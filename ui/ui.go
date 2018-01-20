package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
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
func ListWebsites(node *node.Node) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "GET" {
			jsonData, err := json.Marshal(node.WebsiteMap)
			utils.CheckError(err)
			fmt.Fprint(writer, string(jsonData))
		}
	}
}

// FilterWebsites finds websites' names matching a given keyword
func FilterWebsites(node *node.Node) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "POST" {
			request.ParseForm()
			keyword := strings.Join(request.Form["keywords"], "")
			websites := node.Search(keyword)
			jsonData, err := json.Marshal(websites)
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

			log.Println("[WEBSITES] Importing new website '" + name + "'")

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

			success := node.UpdateWebsite(name, keywords)
			fmt.Fprint(writer, success)
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

// OpenBrowser starts the user's browser on the UI's URL
func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("Cannot open browser, unsupported platform")
	}
	utils.CheckError(err)
}

// StartServer starts listening and serving on addr
func StartServer(uiPort string, node *node.Node) {
	http.Handle("/", ServeUI())
	http.Handle("/w/", ServeWebsites())
	http.Handle("/list", ListWebsites(node))
	http.HandleFunc("/scan", ScanWebsiteFolder)
	http.Handle("/filter", FilterWebsites(node))
	http.HandleFunc("/share", ImportWebsite(node))
	http.HandleFunc("/update", UpdateWebsite(node))

	go OpenBrowser("http://127.0.0.1:" + uiPort)
	http.ListenAndServe("127.0.0.1:"+uiPort, nil)
}
