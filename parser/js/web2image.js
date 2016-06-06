var dropvar = document.getElementById('drop');
dropvar.addEventListener('dragover', handleDragOver, false);
dropvar.addEventListener('drop', handleDrop, false);

// Global Variables
var json_sorted;
var file_name;

function handleDrop(evt) {
    $("#drop").fadeOut();
    evt.stopPropagation();
    evt.preventDefault();

    var file_reader = new FileReader();
    file_reader.onload  = function (e) {
        var content = e.target.result;
        var json = JSON.parse(content);
        json_sorted = sort(json);

        var table = document.getElementById("settings_table");
        table.rows[2].cells[1].innerHTML = json.length;

        build_content(json_sorted);
        $("#content_table").treetable({ expandable: true });
    };

    file_reader.readAsText(evt.dataTransfer.files[0]);
    populate_settings(evt.dataTransfer.files[0]);
    $("#content").fadeIn();
}

function handleDragOver(evt) {
    evt.stopPropagation();
    evt.preventDefault();
    evt.dataTransfer.dropEffect = 'copy';
}

// Event Hanlder for save State
$("#save_state").click(function(){
    save_state(json_sorted,file_name);
})

// Get Original Json and Update It
function save_state(orig_json,file_name){

    // Ensure same format as orignal
    var contents = '[';

    // For each Website Object Passed
    var obj_count = orig_json.length;
    $.each(orig_json, function(index, value){

        var table = document.getElementById('content_table');
        obj:

        // Look Through table for Object then Report Changes
        for (var i = 1, n = table.rows.length; i < n; i++) {

            // Ignore Category Rows
            if (table.rows[i].cells.length > 2) {
                var base_url = table.rows[i].cells[1].innerHTML;
                if (value['Base_url'] === base_url.match(/<a [^>]+>([^<]+)<\/a>/)[1]) {

                    // Update Reviewed
                    if (table.rows[i].cells[0].getElementsByTagName('input')[0].checked) {
                        value['Reviewed'] = 1;
                    } else {
                        value['Reviewed'] = 0;
                    }

                    // Update Comments
                    value['Comments'] = table.rows[i].cells[7].getElementsByTagName('textarea')[0].value;
                    break obj;
                }
            }
        }
        // Pretty print to file
        contents += JSON.stringify(value,null,2);

        // Ensure same format as orignal
        if (index != obj_count-1){
            contents += ',';
        }
    })
    // Ensure same format as orignal
    contents += ']';

    // Setup File Type
    var blob = new Blob([contents], {type:'application/json'});
    var downloadLink = document.createElement("a");

    // Setup Download Link
    downloadLink.download = file_name;
    downloadLink.innerHTML = "Download File";
    downloadLink.href = window.URL.createObjectURL(blob);
    downloadLink.onclick = deletelink;
    downloadLink.style.display = "none";
    document.body.appendChild(downloadLink);

    // Download File
    downloadLink.click();
}

// Remove Link
function deletelink(event) {
    document.body.removeChild(event.target);
}

// Sort json by Category, then by Subcategory
function sort(json) {
    return json.sort(function(a,b) {
        var x = a['Category'];
        var y = b['Category'];
        var s = a['Subcategory'];
        var t = b['Subcategory'];
        return compare([compare(x,y),compare(s,t)],[compare(y,x),compare(t,s)])
    });
}

// Sort
function compare(x,y) {
    return ((x < y) ? -1 : ((x > y) ?1 : 0));
}

// Search Website Table
$("#search_websites").keyup(function() {
    _this = this;
    $.each($("#content_table tbody tr"), function() {
        if($(this).text().toLowerCase().indexOf($(_this).val().toLowerCase()) === -1)
           $(this).hide();
        else
           $(this).show();
    });
});

// Build Settings Header
function populate_settings(filename) {
    var table = document.getElementById("settings_table");
    table.rows[1].cells[1].innerHTML = filename.name;
    file_name = filename.name;
}

// Build Actual Table with Websites and Group categories
function build_content(json) {

    // Trees to store existing categories
    var categories = {};
    var existing_tree = [];

    // Temp String which will contain entire table
    var temp_string = '';

    // Starting category id, used to assign parents
    var cat_id = 1;
    var sub_id = 1;

    // Loop through Json objects
    for (i = 1; i < json.length+1; i++){
        var found = false;

        var item = json[i-1];

        var category = item.Category;
        var sub_category = item.Subcategory;

        // First check if the category already exists in tree
        for (index = 0; index < existing_tree.length; index++) {

            // Category exists
            if (existing_tree[index].category == category) {

                // Track subcategory for specific category
                existing_tree[index].title_id++;

                var current = existing_tree[index];
                var sub_found = false;

                // Check if subcategory exists
                for (sub = 0; sub < current.subcategory.length; sub++) {
                     if (current.subcategory[sub].sub_cat == sub_category) {
                         sub_found = true;
                         var pid = cat_id;
                         cat_id++;
                         temp_string += '<tr data-tt-id="'+cat_id+'" data-tt-parent-id="'+current.subcategory[sub].sub_id+'">';
                         break;
                     }
                }
                // Subcategory does not exist, create new one
                if (sub_found == false) {
                    cat_id++;
                    var sub_obj = {sub_cat: sub_category, sub_id: cat_id};
                    current.subcategory.push(sub_obj);
                    temp_string += '<tr data-tt-id="'+cat_id+'" data-tt-parent-id="'+current.category_id+'">';
                    temp_string += '<td>'+sub_category+'</td></tr>';
                    cat_id++;
                    var current_sub = current.subcategory.length - 1;
                    temp_string += '<tr data-tt-id="'+cat_id+'" data-tt-parent-id="'+current.subcategory[current_sub].sub_id+'">';
                }
                found = true;
                cat_id++;
                break;
            }
        }
        // Category does not exist, creating new category branch
        if (found == false) {

            // create category
            temp_string += '<tr data-tt-id="'+cat_id+'">';
            temp_string += '<td>'+category+'</td></tr>';

            // Create new Object linking category to id
            categories[category] = cat_id;

            // Track parents
            var pid = cat_id;
            cat_id++;

            // Create subcategory
            temp_string += '<tr data-tt-id="'+cat_id+'" data-tt-parent-id="'+pid+'">';
            temp_string += '<td>'+sub_category+'</td></tr>';

            // Object for tracking overall categories
            var sub_obj = {sub_cat: sub_category, sub_id: cat_id};
            var sub_array = [];
            sub_array.push(sub_obj);
            var branch = {category: category, category_id: pid, subcategory: sub_array, subcategory_id: cat_id, title_id: 0};
            existing_tree.push(branch);

            // create first element in category
            var pid = cat_id;
            cat_id++;
            temp_string += '<tr data-tt-id="'+cat_id+'" data-tt-parent-id="'+pid+'">';
            sub_id++;
            cat_id++;
        }

        // Actual data for specific website line item
        if (item['Reviewed'] == 0) {
            temp_string += '<td><input type="checkbox" name="reviewed" value="1"></td>';
        } else {
            temp_string += '<td><input type="checkbox" name="reviewed" value="1" checked></td>';
        }
        temp_string += '<td><a href="'+item["Base_url"]+'" target="_blank">'+item["Base_url"]+'</a></td>';
        temp_string += '<td><a href="'+item["Final_url"]+'" target="_blank">'+item["Final_url"]+'</a></td>';
        temp_string += '<td>'+item.Title+'</td>';
        temp_string += '<td>'+item.Category+'</td>';
        temp_string += '<td><div class="scroll"><img src="images/'+item["Img_name"]+'" alt="'+item["Img_name"]+'"></div></td>';
        temp_string += '<td><div class="scroll">';
        for (var element in item["Headers"]) {
            temp_string += '<b>'+element+':</b> '+item["Headers"][element]+'<br/>';
        }
        temp_string += '</div></td>';
        temp_string += '<td><textarea rows="4" cols="50">'+item["Comments"]+'</textarea></td>';
        temp_string += '</tr>';
    }
    $("#content_table").append(temp_string);
}
