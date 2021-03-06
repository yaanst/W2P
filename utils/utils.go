package utils

import (
	"io/ioutil"
	"log"
	"time"
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

// ConnBufferSize is the os buffer size to receive and send udp packets
const ConnBufferSize int = 10485760 // 10MB

// HeartBeatBufferSize is the default size for buffer waiting for a heartbeat
const HeartBeatBufferSize int = 512

// HeartBeatLimit is the default maximum number of concurrent heartbeats
const HeartBeatLimit int = 50

// HeartBeatTimeout is the default timeout for an answer from a peer
const HeartBeatTimeout time.Duration = time.Duration(10000000000) // 10s

// DataReqTimeout is the timeout before receiving data
const DataReqTimeout time.Duration = time.Duration(10000000000) // 10s

// HashSize is the number of hex character in a sha256 hash (for pieces)
const HashSize int = 64

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
