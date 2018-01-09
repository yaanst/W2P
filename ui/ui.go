package ui

import (
    "fmt"
    "log"
    "strings"
    "net/http"
    "io/ioutil"
    "encoding/json"

    "github.com/yaanst/W2P/node"
    "github.com/yaanst/W2P/utils"
)


// CheckError prints and exits if an error occurs
func CheckError(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

//  ScanWebsiteFolder finds the user's websites names (/scan)
func ScanWebsiteFolder(writer http.ResponseWriter, request *http.Request) {
    if request.Method == "GET" {
        entries, err := ioutil.ReadDir(utils.WebsiteFolder)
        var folders []string
        for _, entry := range entries {
            if entry.IsDir() {
                folders = append(folders, entry.Name())
            }
        }
        json_data, err := json.Marshal(folders)
        CheckError(err)
        fmt.Fprint(writer, string(json_data))
    }
}

// ListWebsites lists all known websites (/list)
func ListWebsites(node node.Node) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        if request.Method == "GET" {
            json_data, err := json.Marshal(node.WebsiteMap)
            CheckError(err)
            fmt.Fprint(writer, string(json_data))
        }
    }
}

// ListWebsitesFiltered lists all known websites matching the keyword (/filter)
func ListWebsitesFiltered(node node.Node) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        if request.Method == "POST" {
            request.ParseForm()
            keywords := strings.Join(request.Form["keywords"], "")
            websites := node.SearchWebsites(keywords)
        }
    }
}


// ShareWebsite creates and shares a new website (/share)
func ShareWebsite(node node.Node) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        if request.Method == "POST" {
            request.ParseForm()
            name := strings.Join(request.Form["name"], "")
            keywords := strings.Join(request.Form["keywords"], "")

            if name != "" {

                website := node.NewWebsite(name)
            }
        }
    }
}

// UpdateWebsite updates an existing website (/update)
func UpdateWebsite(node node.Node) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        if request.Method == "POST" {
            request.ParseForm()
            name := strings.Join(request.Form["name"], "")
            keywords := strings.Join(request.Form["keywords"], "")

            websites := node.UpdateWebsite(name, keywords)
            json_data, err := json.Marshal(websites)
            CheckError(err)
            fmt.Fprint(writer, string(json_data))
        }
    }
}

