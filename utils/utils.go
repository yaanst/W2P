package utils

import (
	"io/ioutil"
	"log"
)

// ---------
// - Const -
// ---------

// WebsiteDir is the path to the directory containing all website
const WebsiteDir string = "./website/"

// MetadataFile is the file name of file which is present in every website in
// the WebsiteMap
const MetadataFile string = "/metadata"

// SeedDir is the path to the directory containing all seeding binary archive
const SeedDir string = "./seed/"

// DefaultPieceLength is the default length in bytes for a piece (8KB)
const DefaultPieceLength = 8192

// -----------
// - Helpers -
// -----------

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
