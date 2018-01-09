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

// CheckError logs and exits if an error occurs
func CheckError(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

//  ScanWebsiteFolder finds the user's websites names (/scan)
func ScanWebsiteFolder(writer http.ResponseWriter, request *http.Request) {
    if request.Method == "GET" {
        folders := utils.ScanDir(node.WebsiteDir)
		    jsonData, err := json.Marshal(folders)
		    CheckError(err)
		    fmt.Fprint(writer, string(jsonData))
    }
}

// ListWebsites lists all known websites (/list)
func ListWebsites(node node.Node) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
		    if request.Method == "GET" {
			      jsonData, err := json.Marshal(node.WebsiteMap)
			      CheckError(err)
            fmt.Fprint(writer, string(jsonData))
        }
    }
}

// ImportWebsite imports a new website from the UI and add it to the
// seeding websites
func ImportWebsite(node node.Node) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        if request.Method == "POST" {
            request.ParseForm()
      			name := strings.Join(request.Form["name"], "")

			      keywordsString := strings.Join(request.Form["keywords"], "")
			      keywords := strings.Split(keywordsString, ",")

			      if name != "" {
				        node.AddNewWebsite(name, keywords)
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