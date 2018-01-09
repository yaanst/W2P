package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/yaanst/W2P/node"
	"github.com/yaanst/W2P/utils"
)

// ScanWebsiteFolder finds the user's websites names (/scan)
func ScanWebsiteFolder(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		folders := utils.ScanDir(utils.WebsiteDir)
		jsonData, err := json.Marshal(folders)
		utils.CheckError(err)
		fmt.Fprint(writer, string(jsonData))
	}
}

// ListWebsites lists all known websites (/list)
func ListWebsites(node node.Node) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "GET" {
			jsonData, err := json.Marshal(node.WebsiteMap)
			utils.CheckError(err)
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

			keywordsString := strings.Join(request.Form["keywords"], "")
			keywords := strings.Split(keywordsString, ",")

			node.UpdateWebsite(name, keywords)
		}
	}

// ServeWebsite looks for the files belonging to a website and serves them
func ServeWebsite(writer http.ResponseWriter, request *http.Request) {
    splitPath := strings.Split(request.URL.EscapedPath(), "/")
    websiteName := splitPath[:len(splitPath) - 1]

    fs := http.FileServer(http.Dir(utils.WebsiteFolder + websiteName))
    http.Handle("/", fs)
}
