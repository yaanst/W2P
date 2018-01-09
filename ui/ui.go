package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/yaanst/W2P/node"
	"github.com/yaanst/W2P/utils"
	"github.com/yaanst/W2P/w2pcrypto"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// List all folders in the WebsiteFolder
func ReadWebsiteFolder(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		folders := utils.ScanDir(node.WebsiteDir)
		json_data, err := json.Marshal(folders)
		CheckError(err)
		fmt.Fprint(writer, string(json_data))
	}
}

// List all known websites
func ListWebsites(node node.Node) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "GET" {
			json_data, err := json.Marshal(node.WebsiteMap)
			CheckError(err)
			fmt.Fprint(writer, string(json_data))
		}
	}
}

// Start seed a new website
func SeedWebsite(node node.Node) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "POST" {
			request.ParseForm()
			name := strings.Join(request.Form["name"], "")
			keywords := strings.Join(request.Form["keywords"], "")

			if name != "" {
				privkey := w2pcrypto.CreateKey()
				pubkey := &privkey.PublicKey
				w2pcrypto.SaveKey(name+".key", privkey)

				website := node.NewWebsite(name)
				website.Keywords = strings.Split(keywords, ",")
				website.PubKey = pubkey
				website.Version = 1

				// Probel, pubkey has no string representation
			}
		}
	}
}

// /update
func UpdateWebsite(node node.Node) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "POST" {

		}
	}
}
