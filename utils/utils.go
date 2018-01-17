package utils

import (
	"log"
	"time"
	"io/ioutil"
)

// ---------
// - Const -
// ---------

// UIDir is the path to the directory containing the UI
const UIDir string = "./ui/webpage"

// WebsiteDir is the path to the directory containing all websites
const WebsiteDir string = "./website/"

// MetadataDir is the directory in which we save all serialization of websites
const MetadataDir string = "./metadata/"

// SeedDir is the path to the directory containing all seeding binary archive
const SeedDir string = "./seed/"

// KeyDir is the directory containing crypto keys
const KeyDir string = "./keys/"

// DefaultPieceLength is the default length in bytes for a piece (8KB)
const DefaultPieceLength int = 8192

// ListenBufferSize is the default size in bytes for buffer holding incoming
// messages
const ListenBufferSize int = 65536

// HeartBeatBufferSize is the default size for buffer waiting for a heartbeat
const HeartBeatBufferSize int = 512

// HeartBeatTimeout is the default timeout for an answer from a peer
const HeartBeatTimeout time.Duration = time.Duration(5000000000) // 5s

// -----------
// - Helpers -
// -----------

// CheckError logs and exits if an error occurs
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// ScanDir scans a folder and return a list of subfolder names
func ScanDir(path string) []string {
	entries, err := ioutil.ReadDir(path)
	CheckError(err)

	var subfolders []string
	for _, entry := range entries {
		if entry.IsDir() {
			subfolders = append(subfolders, entry.Name())
		}
	}

	return subfolders
}

// Contains check if a slice of string contains that particular string
func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
