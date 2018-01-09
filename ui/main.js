/*********************
        Variables
**********************/
var BASE_URL = "";
(function compute_base_url() {
    proto = window.location.protocol;
    host = window.location.hostname;
    port = window.location.port;
    if (proto != "") {
        BASE_URL += proto + "//";
    }
    if (host != "") {
        BASE_URL += host;
    }
    if (port != "") {
        BASE_URL += ":" + port;
    }
    BASE_URL += "/";
})();

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
        print_website_folder(data, false);
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
                name: $("#share_folders_select").val(),
                keywords: $("#keywords_input").val()
            },
            function (data, status) {}
        );

    } else if (EXTRA_WINDOW == "update") {
        $.post("/update", 
            {
                name: $("#share_folders_select").val(),
                keywords: $("#keywords_input").val()
            },
            function (data, status) {}
        );
    }
    $("#websites_section_extra").hide();
    $(".update").hide();
    $(".share").hide();
});

// Filter the website list based on keywords entered in the input field
$(document).on("click", "#filter_apply_button", function() {
    k = $("#filter_keywords").val();
    console.log(k);
    $.post("/filter", 
        {
            keywords: k
        },
        function (data, status) {
            print_website_filtered(data, true);
            $("#current_filter").html("<b>Current filter:</b> " + k);
            $("#wesbites_list").hide();
            $("#wesbites_list_filtered").show();
        }
    );
});

// Clears the filters applied on the website list
$(document).on("click", "#filter_clear_button", function() {
    console.log("clear");
    $("#current_filter").html("");
    $("#filter_keywords").val("");
    $("websites_list_filtered").hide();
    $("#wesbites_list").show();
});



/*********************
        Helpers
**********************/
// Format and print the JSON string for websites
function print_websites_list(data, filtered) {
    websites = JSON.parse(data);

    // sort websites by name
    var keys =  [];
    for (var name in websites) keys.push(name);
    keys = keys.sort(function(a,b) {
        return a.localeCompare(b)
    });

    list = ""
    if (num_websites > 0) {
        for (name in keys) {
            w = websites[name];
            list += `<li><a target="_blank" href="${BASE_URL + name}">${name}</a></li><br />`
            delete w;
        }
    }
    if (filtered == true) {
        $("#websites_list_filtered").html(list);
    } else {
        $("#websites_list").html(list);
    }
    delete list;
    delete keys;
    delete websites;
}

// Format and print 
function print_website_folder(data) {
    websites = JSON.parse(data);
    websites = websites.sort();
    
    options = $("#folders_select").innerHTML
    for (idx in websites) {
        w = websites[idx]
        options += `<option value="${w}">${w}</option>`
        delete w;
    }
    $("#folders_select"),html(options);
    delete websites;
    delete options;
}
