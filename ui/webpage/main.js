/***********************************
        Retrieve content
************************************/
// Get the website list
(function fetch_website_list() {
    $.get("/list", function(data) {
        print_websites_list(data);
        setTimeout(fetch_website_list, 1000);
    });
})();

// Get status info
(function fetch_status_info() {
    $.get("/status", function(data) {
        print_status_info(data);
        setTimeout(fetch_status_info, 1000);
    });
})();

/********************
        Buttons
*********************/
var EXTRA_WINDOW = ""
// Triggers "websites" folder scan and show hidden inputs
$(document).on("click", "#share_website_button", function() {
    $.get("/scan", function(data) {
        print_website_folder(data);
    });
    EXTRA_WINDOW = "share";
    $("#websites_section_extra").show();
    $(".update").hide();
    $(".share").show();
});

// Triggers "websites" folder scan and show hidden inputs
$(document).on("click", "#update_website_button", function() {
    $.get("/scan", function(data) {
        print_website_folder(data);
    });
    EXTRA_WINDOW = "update";
    $("#websites_section_extra").show();
    $(".share").hide();
    $(".update").show();
});

// Send name and keywords to share/update website
$(document).on("click", "#websites_extra_button", function() {
    if (EXTRA_WINDOW == "share") {
        $.post("/share", 
            {
                name: $("#extra_folders_select").val(),
                keywords: $("#keywords_input").val()
            },
            function (data, status) {}
        );

    } else if (EXTRA_WINDOW == "update") {
        $.post("/update", 
            {
                name: $("#extra_folders_select").val(),
                keywords: $("#keywords_input").val()
            },
            function (data, status) {
                if (data == "false") {
                    alert("Website could not be updated.\nYou can update only websites already existing and that you own");
                }
            }
        );
    }
    $("#websites_section_extra").hide();
    $(".update").hide();
    $(".share").hide();
});

// Filter the website list based on keywords entered in the input field
$(document).on("click", "#filter_apply_button", function() {
    k = $("#filter_keywords").val();
    $.post("/filter", 
        {
            keywords: k
        },
        function (data, status) {
            print_websites_filtered(data);
        }
    );
    $("#current_filter").html("<b>Current filter:</b> " + k);
    $("#websites_list").hide();
    $("#websites_list_filtered").show();
});

// Clears the filters applied on the website list
$(document).on("click", "#filter_clear_button", function() {
    $("#current_filter").html("");
    $("#filter_keywords").val("");
    $("#websites_list_filtered").hide();
    $("#websites_list").show();
});

/*********************
        Helpers
**********************/
// Format and print the JSON string for websites
function print_websites_list(data) {
    websites = JSON.parse(data);
    if (websites != null) {
        websites = websites.W
        num_websites = Object.keys(websites).length;

        if (num_websites > 0) {
            // sort websites by name
            var sorted =  [];
            for (var name in websites) sorted.push(websites[name]);
            sorted = sorted.sort(function(a,b) {
                return (a.Name).localeCompare(b.Name)
            });

            list = ""
            for (idx in sorted) {
                w = sorted[idx];
                list += `<li><a target="_blank" href="/w/${w.Name}">${w.Name}</a></li>`
                delete w;
            }

            $("#websites_list").html(list);
            delete list;
            delete sorted;
        }
        delete num_websites;
    }
    delete websites;
}

// Format and print the filtered list of websites
function print_websites_filtered(data) {
    websites = JSON.parse(data);
    list = ""
    if (websites != null) {
        websites = websites.sort()
        for (idx in websites) {
            w = websites[idx];
            list += `<li><a target="_blank" href="/w/${w}">${w}</a></li>`
            delete w;
        }
    }
    $("#websites_list_filtered").html(list);
    delete list;
    delete websites;
}

// Format and print the contents of the website folder
function print_website_folder(data) {
    websites = JSON.parse(data);
    websites = websites.sort();

    options = $("#extra_folders_select").innerHTML
    for (idx in websites) {
        w = websites[idx]
        options += `<option value="${w}">${w}</option>`
        delete w;
    }
    $("#extra_folders_select").html(options);
    delete websites;
    delete options;
}

// Format and print the stauts information
function print_status_info(data) {
    info = JSON.parse(data);
    name = "<b>Name:</b> " + info["name"];
    addr = "<b>Address:</b> " + info["addr"];
    peers = "<b>#Peers:</b> " + info["peers"];
    websites = "<b>#Websites:</b> " + info["websites"];
    $("#status_bar_name").html(name);
    $("#status_bar_addr").html(addr);
    $("#status_bar_peers").html(peers);
    $("#status_bar_websites").html(websites);
    delete info;
    delete name;
    delete addr;
    delete peers;
    delete websites;
}
