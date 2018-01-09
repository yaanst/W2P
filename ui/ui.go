package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/yaanst/W2P/node"
	"github.com/yaanst/W2P/structs"
	"github.com/yaanst/W2P/utils"
)

// CheckError checks for an error and print it if any
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// ReadWebsiteFolder lists all folders in the WebsiteFolder
func ReadWebsiteFolder(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		folders := utils.ScanDir(node.WebsiteDir)
		jsonData, err := json.Marshal(folders)
		CheckError(err)
		fmt.Fprint(writer, string(jsonData))
	}
}

// ListWebsites lists all known websites
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

// UpdateWebsite update a website already present in the Node
func UpdateWebsite(node node.Node) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "POST" {

		}
	}
}
