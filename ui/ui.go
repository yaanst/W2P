package ui

import (
	"fmt"
	"strings"
	"net/http"
	"encoding/json"

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
func ListWebsites(node *node.Node) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "GET" {
			jsonData, err := json.Marshal(node.WebsiteMap)
			utils.CheckError(err)
			fmt.Fprint(writer, string(jsonData))
		}
	}
}

// ImportWebsite imports a new website from the UI and add it to the
// seeding websites (/share)
func ImportWebsite(node *node.Node) http.HandlerFunc {
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
func UpdateWebsite(node *node.Node) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "POST" {
			request.ParseForm()
			name := strings.Join(request.Form["name"], "")

			keywordsString := strings.Join(request.Form["keywords"], "")
			keywords := strings.Split(keywordsString, ",")

			node.UpdateWebsite(name, keywords)
		}
	}
}

// ServeUI serves the UI page
func ServeUI() http.Handler {
    return http.FileServer(http.Dir(utils.UIDir))
}
// ServeWebsites serves the website folder
func ServeWebsites() http.Handler {
    return http.StripPrefix("/w/", http.FileServer(http.Dir(utils.WebsiteDir)))
}

// StartServer starts listening and serving on addr
func StartServer(uiPort string, node *node.Node) {
    http.Handle("/", ServeUI())
    http.Handle("/w/", ServeWebsites())
    http.Handle("/list", ListWebsites(node))
    http.HandleFunc("/scan", ScanWebsiteFolder)
    http.HandleFunc("/share", ImportWebsite(node))
    http.HandleFunc("/update", UpdateWebsite(node))

    http.ListenAndServe("127.0.0.1:" + uiPort, nil)
}
