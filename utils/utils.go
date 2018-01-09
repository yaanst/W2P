package utils

import (
	"io/ioutil"
	"log"
)

// ScanDir scans a folder and return a list of subfolder names
func ScanDir(path string) []string {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	var subfolders []string
	for _, entry := range entries {
		if entry.IsDir() {
			subfolders = append(subfolders, entry.Name())
		}
	}

	return subfolders
}
