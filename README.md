# Web2Image

Go tool to convert web pages to screenshots and categorize the websites.

### Prerequisite

Install WebKit2Png
```
apt-get install python-qt4 libqt4-webkit xvfb
pip install webkit2png
```

### Install
Download Zip and Build
```
go build ./web2image.go
```

### Usage

```
> web2image -h

Usage of ./web2image:
  -category string
    	Categories File (default "categories.csv")
  -list string
    	File containing urls to scan (default "websites.txt")
  -out string
    	Name of json output (default "web2image.json")
  -threads int
    	# of Threads to run (default 5)
  -timeout int
    	HTTP Timeout in seconds (default 7)
  -verbose
    	Enable verbose output
```

### Example

```
> web2image -list=websites.txt -out=example.json --verbose=true
             _    ___ _
 __ __ _____| |__|_  |_)_ __  __ _ __ _ ___
 \ V  V / -_) '_ \/ /| | '  \/ _' / _' / -_)
  \_/\_/\___|_.__/___|_|_|_|_\__,_\__, \___|
                                  |___/   
Worker 0 URL: http://google.com
Worker 1 URL: http://localhost
Worker 2 URL: http://localhost:8080
Redirect from: http://google.com -> http://www.google.com/
[+] Rendered: http_localhost_8080.png
[+] Rendered: http_localhost.png
[+] Rendered: http_google_com.png
Starting grouping
Finished Grouping
Websites being grouped based on title
Title does not exist: Apache Tomcat
Title does not exist: Apache2 Debian Default Page It works
Title does not exist: Google
Finished Grouping Websites
Starting Categorization
Categorizing Group: Apache Tomcat
Categorizing Group: Apache2 Debian Default Page It works
Categorizing Group: Google
Categorization done

```

### Parser

To use parser, drag and drop json to index.html

### Authors

* **Mitchell Hennigan**

### License

This project is licensed under the GNU General Public License - see the [LICENSE](LICENSE) file for details

### Acknowledgments

* Adamn - [Python-WebKit2Png](https://github.com/adamn/python-webkit2png)
