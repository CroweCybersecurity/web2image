package main

import (
    "bufio"
    "crypto/tls"
    "encoding/json"
    "encoding/csv"
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "time"
)

// Object for Each Website
type web2struct struct {

    // URL - Input URL
    Base_url string

    // final_url - URL after redirects
    Final_url string

    // Headers [] - HTTP Header Object Array
    Headers map[string][]string

    // Title
    Title string

    // Img Location (images/website_port.png)
    Img_name string

    // Match Condition (Header/Title/%Img Match Percentage)
    Condition string

    // Category - Default IIS/Printer/Tomcat
    Category string

    // SubCategory - IIS 7.0
    Subcategory string

    // Reviewed
    Reviewed string

    // Comments
    Comments string
}

// Array of Grouped Sites by Title
type web2group struct {

    // Title that Matched the Websites Together
    Title string

    // Array of Web2Structs
    Websites []web2struct
}

// Array of Grouped Websites
var website_array []web2group

// Tracks amount of Hosts
var host_counter int

// Verbose Output
var verbose bool

/*-----------------------------------------------------------------------
|  Function: Main
|  Description: Setup waigroups and channels for workers.
|  Inputs: Command Line Input
|  Returns: None
*-----------------------------------------------------------------------*/
func main()  {

    var file_urls string
    var file_output string
    var file_category string

    // Setup CLI Inputs
    flag.StringVar(&file_urls, "list" , "websites.txt" , "(Required) File containing urls to scan")
    flag.StringVar(&file_output, "out" , "web2image.json" , "Name of json output")
    flag.StringVar(&file_category, "category", "categories.csv", "Categories File")
    ptr_threads := flag.Int("threads", 5 , "# of Threads to run")
    ptr_timeout := flag.Int("timeout", 7, "HTTP Timeout in seconds")
    ptr_verbose := flag.Bool("verbose", false, "Enable verbose output")
    flag.Parse()

    if len(os.Args) == 1 {
        flag.Usage()
        os.Exit(1)
    }

    threads := *ptr_threads
    verbose = *ptr_verbose
    timeout := *ptr_timeout

    // L33t Banner
    banner := `            _    ___ _
 __ __ _____| |__|_  |_)_ __  __ _ __ _ ___
 \ V  V / -_) '_ \/ /| | '  \/ _' / _' / -_)
  \_/\_/\___|_.__/___|_|_|_|_\__,_\__, \___|
                                  |___/  `

    fmt.Println("\x1b[34;1m",banner, "\x1b[0m")

    // Setup WaitGroups
    wg_webclients := new(sync.WaitGroup)
    wg_group := new(sync.WaitGroup)
    wg_compare := new(sync.WaitGroup)
    wg_output := new(sync.WaitGroup)

    // Setup Data Channels
    chan_file_in := make(chan string, threads)
    chan_webclient_out := make(chan web2struct, 10000)
    chan_group_out := make(chan web2group, 10000)
    chan_compare_out := make(chan web2struct, 10000)

    // Setup Output File
    if _, err := os.Stat(file_output); err == nil {
        os.Remove(file_output)
    }
    file_report, _ := os.OpenFile(file_output, os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600)

    // Setup images/ Directory
    if _,err := os.Stat("images/"); os.IsNotExist(err) {
          os.MkdirAll("images",0711)
    }

    // Prep Json File
    w := bufio.NewWriter(file_report)
    w.WriteString("[")
    w.Flush()

    // Setup and Read Match Categories File
    category_file, _ := os.Open(file_category)
    reader := csv.NewReader(category_file)
    reader.FieldsPerRecord = -1
    categories, _ := reader.ReadAll()

    // Start Output Routine
    go worker_output(chan_compare_out, wg_output, file_report)
    wg_output.Add(1)
    wg_group.Add(1)

    // Start Goroutines for Webclient Pool
    for id := 0; id < threads; id++ {
        wg_webclients.Add(1)
        go pool_webclient(id, timeout, chan_file_in, chan_webclient_out, wg_webclients)
    }

    // Start Routine for Read File
    go readfile(file_urls, chan_file_in)

    // Wait for All Results and Fan in
    wg_webclients.Wait()
    if verbose {fmt.Println("Starting grouping")}
    go worker_group(chan_webclient_out, chan_group_out, wg_group)
    if verbose {fmt.Println("Finished Grouping")}

    // Wait for Grouping to Finish, then Fan Out and Compare
    wg_group.Wait()
    if verbose{fmt.Println("Starting Categorization")}
    for id := 0; id < threads; id++ {
        wg_compare.Add(1)
        go pool_compare_group(id, chan_group_out, chan_compare_out, wg_compare, categories)
    }

    // Wait for Categorization to Finish, then Fan In and Close File
    wg_compare.Wait()
    if verbose {fmt.Println("Categorization done")}
    wg_output.Wait()

    close(chan_webclient_out)
    close(chan_group_out)
    close(chan_compare_out)

    category_file.Close()
}
/*-----------------------------------------------------------------------
|  Function:
|  Description: Read in List of URLs from File
|  Inputs:
|        file_urls - Filename of the Website List
|  Output:
|        chan_file_in - Channel of URLs for WebClient
*-----------------------------------------------------------------------*/
func readfile(file_urls string, chan_file_in chan<- string) {

    file,_ := os.Open(file_urls)
    defer file.Close()
    scanner := bufio.NewScanner(file)
    for scanner.Scan(){
        chan_file_in <- scanner.Text()
    }
    close(chan_file_in)
}

/*-----------------------------------------------------------------------
|  Function: Pool WebClient
|  Description: GORoutine for Looping Through Available URLS and getting
|        GET ifnroamtion and screenshot. Once done, builds a Web2Structs object.
|  Inputs:
|        id - ID of the Individual Worker
|        chan_file_in - Channel of Available URLs
|        wg_webclients - Waitgroup to Signal Worker is Done
|  Output:
|        chan_webclient_out - Channel of Ready to Group Web2Struct Objects
*-----------------------------------------------------------------------*/
func pool_webclient(id int, timeout int, chan_file_in <-chan string, chan_webclient_out chan<- web2struct, wg_webclients *sync.WaitGroup) {

    defer wg_webclients.Done()

    for url := range chan_file_in {

        if verbose{fmt.Println("Worker", id, "URL:",url)}

        host_counter++

        // Follow Redirect Funciton and Grab Header
        final_url, header, title := webclient_follow(timeout, url)

        // Render Function
        img_name := webclient_render(timeout, url)

        // Set Empty Categories for Now
        condition := ""
        category := ""
        subcategory := ""
        reviewed := "0"
        comments := ""

        // Build Web2Struct Object
        web2_obj := web2struct{Base_url: url, Final_url: final_url, Headers: header, Title: title, Img_name: img_name, Condition: condition, Category: category, Subcategory: subcategory, Reviewed: reviewed, Comments: comments}
        chan_webclient_out <- web2_obj
    }
}

/*-----------------------------------------------------------------------
|  Function: Webclient Follow
|  Description: Performs a GET request on website, then pulls header and
|        all necessary info for matching.
|  Inputs:
|        url - Base URL of the Website
|        time_sec - Timeout for GET Request
|  Returns:
|        final_url - URL After Redirects
|        header - Array of Header Information
|        title - Title of the Website
*-----------------------------------------------------------------------*/
func webclient_follow(time_sec int, url string) (final_url string, header map[string][]string, title string) {

    // Disable Cert Check
    tr := &http.Transport {
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    timeout := time.Duration(time.Duration(time_sec) * time.Second)
    client := &http.Client {
        Transport: tr,
        Timeout: timeout,
    }

    // Get Request to URL
    response, err := client.Get(url)

    // Return Object for Websites that Cannot be Accessed
    if err != nil {
        if verbose{fmt.Println("Error:",err)}
        final_url = url
        header = make(map[string][]string)
        fault := err.Error()
        header["Error"] = []string{fault}

        title = "Invalid"
        return
    }
    final_url = response.Request.URL.String()

    if url != final_url {
        if verbose{fmt.Println("Redirect from:", url, "->", final_url)}
    }
    header = response.Header
    contents, err := ioutil.ReadAll(response.Body)

    // Title Capture
    reg,_ := regexp.Compile(`<title>(.+)?</title>`)
    found := reg.FindStringSubmatch(string(contents))
    if len(found) > 0 {
        title = found[1]

        // Check if ASCII title
        for _, c := range title {
            if c > 127 {
                title = url
                break
            }
        }
        // Blank Title
        if len(title) == 0 {
            title = "no title"
        }
    // No Title
    } else {
        title = "no title"
    }

    return
}

/*-----------------------------------------------------------------------
|  Function: WebClient Render
|  Description: Takes screenshot of website, can be substituted for any Tool
|  Inputs:
|        final_url - URL of the Webpage After Redirects
|  Returns:
|        filename - Filename of the Rendered Picture
*-----------------------------------------------------------------------*/
func webclient_render(time_sec int, final_url string) (filename string) {

    // Strip Prefix on URL
    regex, _ := regexp.Compile("://")
    filename = regex.ReplaceAllString(final_url, "_")

    // Change (. or /) to _
    dot_regex, _ := regexp.Compile("\\.|/!$|:|/")
    filename = dot_regex.ReplaceAllString(filename, "_")

    // Add .PNG File Extension and Remove / to prevent directories
    end_regex, _ := regexp.Compile("_?/?$")
    filename = end_regex.ReplaceAllString(filename, ".png")

    file_path := "images/" + filename
    timeout := strconv.Itoa(time_sec)

    // Execute Render Command
    cmd := exec.Command("/usr/bin/python", "webkit2png/scripts.py", final_url, "-o", file_path, "-t", timeout)
    error := cmd.Start()
    if error != nil {
        log.Fatal(error)
    }
    cmd.Wait()
    fmt.Println("[+] Rendered:", filename)

    return
}

/*-----------------------------------------------------------------------
|  Function: Worker Group
|  Description: Single Instance to Group Web2Objects with Similiar Titles
|  Inputs:
|        chan_webclient_out - Channel of Complete Web2Objects
|        wg_group - Signal indivual Worker is Done
|  Returns: None
|        chan_group_out - Channel of Grouped Websites
*-----------------------------------------------------------------------*/
func worker_group(chan_webclient_out <-chan web2struct, chan_group_out chan<- web2group, wg_group *sync.WaitGroup) {

    if verbose {fmt.Println("Websites being grouped based on title")}

    F:
    for {
      select {
          case web2_obj := <- chan_webclient_out:
              web_arr := make([]web2struct, 1)
              web_arr[0] = web2_obj

              // Remove Special Characters from title for categories
              remove_special, _ := regexp.Compile("[]$&+,:;=?@#|'<>.^*()/%!_-]")
              fuzzy_title := remove_special.ReplaceAllString(web2_obj.Title, "")

              group_obj := web2group{Title: fuzzy_title, Websites: web_arr}
              exist := false

              for index, element := range website_array {
                  if fuzzy_title == element.Title {

                      // If Title Exists, Group Already Exists
                      if verbose {fmt.Println("Title exists:", fuzzy_title)}
                      (website_array[index]).Websites = append((website_array[index]).Websites, web2_obj)

                      exist = true
                      break
                  }
              }

              // Title Does not Exist, Create New Group
              if !exist {
                  if verbose {fmt.Println("Title does not exist:", fuzzy_title)}
                  website_array = append(website_array, group_obj)
              }

        default:
            break F
        }
    }

    // Once all Websites have been assigned a group, Write website_array to Channel
    if verbose {fmt.Println("Finished Grouping Websites")}
    for _, web := range website_array {
        chan_group_out <- web
    }

    defer wg_group.Done()
}

/*-----------------------------------------------------------------------
|  Function: Pool Compare Group
|  Description: GoRoutines of Magic Group Comparision Function which compares each group
|  Inputs:
|        id - ID of the Comparison Worker
|        chan_group_out - Channel of Finished Website Objects
|        wg_compare - Waitgroup to Singal Worker is Finished
|        category_file - Filename of the Categories.csv file
|  Output:
|        chan_compare_out - Website Web2Object with Updated Category Values
*-----------------------------------------------------------------------*/
func pool_compare_group(id int, chan_group_out <-chan web2group, chan_compare_out chan<- web2struct, wg_compare *sync.WaitGroup, categories [][]string) {

    F:
    for {
      select {
          case web_group_obj := <- chan_group_out:

              if verbose{fmt.Println("Categorizing Group:", web_group_obj.Title)}

              // Search Line by Line
              for _, line := range categories {

                  // Default IIS,Title,IIS,IIS\d+|IIS
                  matchcategory := line[0]
                  matchloc := strings.ToLower(line[1])
                  matchregex := strings.ToLower(line[2])

                  // Assign Category
                  var search string
                  if matchloc == strings.ToLower("Title") {
                        search = strings.ToLower(((web_group_obj).Websites[0]).Title)
                  } else if matchloc == strings.ToLower("Server") {
                        if val, ok := ((web_group_obj).Websites[0]).Headers["Server"]; ok {
                            search = strings.ToLower(val[0])
                        }
                  }

                  // Match Based on Either Title or Server
                  matched, _ := regexp.MatchString(matchregex, search)
                  if matched {
                      for index, _ := range web_group_obj.Websites {
                          web_group_obj.Websites[index].Category = matchcategory
                          if (matchcategory == "Unknown"){
                              web_group_obj.Websites[index].Subcategory = "Other"
                          } else {
                              web_group_obj.Websites[index].Subcategory = (web_group_obj.Websites[0]).Title
                          }
                          web_group_obj.Websites[index].Condition = matchloc
                      }
                      break
                  }
                  // ELSE PLACEHOLDER FOR IMAGE COMPARISON IF NOT MATCHED
              }

              // Write Categorized Elements to Channel
              for _, group_element := range web_group_obj.Websites {
                  chan_compare_out <- group_element
              }

          default:
              defer wg_compare.Done()
              break F
      }
    }
}
/*-----------------------------------------------------------------------
|  Function: Worker Output
|  Description: Appends Json to the File, then Closes File
|  Inputs:
|        chan_compare_out - Website Web2Object with Updated Category Values
|        wg_output - WaitGroup to signal Output is Finished Before Closing Program
|        compare_done - Channel That Singals when all Comparision Workers are Done
|        file_report - Filename of the output .json
*-----------------------------------------------------------------------*/
func worker_output(chan_compare_out <-chan web2struct, wg_output *sync.WaitGroup, file_report *os.File) {

    w := bufio.NewWriter(file_report)

    var write_finish bool
    write_finish = false

    F:
    for {
      select {
          case web2_obj := <- chan_compare_out:

              // Format Obj for json
              json_obj, _ := json.MarshalIndent(web2_obj, "","    ")

              // Append to File
              file_report.Write(json_obj)
              host_counter--
              if host_counter != 0 {
                  w.WriteString(",\n")
              } else {
                  w.WriteString("\n")
                  write_finish = true

              }
              w.Flush()

          default:
              if write_finish {
                  w.WriteString("]")
                  w.Flush()

                  defer file_report.Close()
                  break F
              }
      }
    }
    defer wg_output.Done()
}
