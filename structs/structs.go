package structs

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
    "errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yaanst/W2P/utils"
	"github.com/yaanst/W2P/w2pcrypto"
)

// -----------
// - Structs -
// -----------

// Peer is a peer represented by a pointer to a UDP address
type Peer net.UDPAddr

// Peers is a collection of Peer
type Peers struct {
	mux sync.Mutex
	P   []*Peer
}

// WebsiteMap is a map from a Website PubKey to the Website
type WebsiteMap struct {
	mux sync.Mutex
	W   map[string]*Website
}

// Website is a structure that represents a website
type Website struct {
	Name        string
	Seeders     *Peers
	Keywords    []string
	PubKey      *w2pcrypto.PublicKey
	PieceLength int
	Pieces      string
	Version     int
}

// RoutingTable is a table which keeps in memory possible route for a dest
// if a Peer is not directly reachable
type RoutingTable struct {
	mux sync.Mutex
	R   map[string]*Peers
}

// ----------------
// - Constructors -
// ----------------

// ParsePeer construct a Peer from a string of format "addr:port"
func ParsePeer(peerString string) *Peer {
	udpAddr, err := net.ResolveUDPAddr("udp4", peerString)

	if err != nil {
		log.Fatal(err)
	}

	peer := Peer(*udpAddr)

	return &peer
}

// ParsePeers construct a collection of type Peers from a string
// format of string: addr:port,addr2:port2,addr3:port3
func ParsePeers(peersString string) *Peers {
	if peersString != "" {
		addrList := strings.Split(peersString, ",")

		peers := NewPeers()

		for _, addr := range addrList {
			peer := ParsePeer(addr)
			peers.Add(peer)
		}

		return peers
	}
	return NewPeers()
}

// NewPeers constructs a new Peers object (list of peer with a mutex)
func NewPeers() *Peers {
	return &Peers{
		P: make([]*Peer, 0, 5),
	}
}

// NewWebsiteMap constructs a WebsiteMap object
func NewWebsiteMap() *WebsiteMap {
	return &WebsiteMap{
		W: make(map[string]*Website),
	}
}

// NewWebsite constructs a new Website data structure
func NewWebsite(name string, keywords []string) *Website {
	privKey, pubKey := w2pcrypto.CreateKey()
	privKey.Save(name)

	seeders := NewPeers()

	return &Website{
		Name:     name,
		Seeders:  seeders,
		Keywords: keywords,
		PubKey:   pubKey,
		Version:  1,
	}
}

// LoadWebsite constructs a Website from a metadata file
func LoadWebsite(name string) *Website {
	jsonData, err := ioutil.ReadFile(utils.MetadataDir + name)
	if err != nil {
		log.Fatal(err)
	}

	var website *Website
	err = json.Unmarshal(jsonData, website)
	if err != nil {
		log.Fatal(err)
	}

	return website
}

// NewRoutingTable constructs a RoutingTable object
func NewRoutingTable() *RoutingTable {
	return &RoutingTable{
		R: make(map[string]*Peers),
	}
}

// -----------
// - Methods -
// -----------

// Peer

// PeerEquals compare two peers
func PeerEquals(p1, p2 *Peer) bool {
	return p1.String() == p2.String()
}

func (p *Peer) String() string {
	return fmt.Sprintf("%v:%v", p.IP.String(), p.Port)
}

// Peers

// Contains check if the Peers slice contains a Peer
func (peers *Peers) Contains(peer *Peer) bool {
	peers.mux.Lock()
	for _, p := range peers.P {
		if PeerEquals(p, peer) {
			peers.mux.Unlock()
			return true
		}
	}
	peers.mux.Unlock()
	return false
}

// Add adds a Peer to the Peers if not already present
func (peers *Peers) Add(peer *Peer) {
	if !peers.Contains(peer) {
		peers.mux.Lock()
		peers.P = append(peers.P, peer)
		peers.mux.Unlock()
	}
}

// Remove removes a Peer from the Peers
func (peers *Peers) Remove(peer *Peer) {
	peers.mux.Lock()
	for i, p := range peers.P {
		if PeerEquals(p, peer) {
			// cut slice in 2 part and append without i-th element inbetween
			peers.P = append(peers.P[:i], peers.P[i+1:]...)
			break
		}
	}
	peers.mux.Unlock()
}

// GetAll returns a copy of the list of Peer
func (peers *Peers) GetAll() []Peer {
	var peerList []Peer
	peers.mux.Lock()
	for _, p := range peers.P {
		peerList = append(peerList, *p)
	}
	peers.mux.Unlock()

	return peerList
}

// WebsiteMap

// Set adds/updates a website to the website map
func (wm *WebsiteMap) Set(website *Website) {
	wm.mux.Lock()
	wm.W[website.Name] = website
	wm.mux.Unlock()
}

// Get return the Website struct given its name
func (wm *WebsiteMap) Get(name string) *Website {
	wm.mux.Lock()
	website := wm.W[name]
	wm.mux.Unlock()

	return website
}

// GetIndices returns of copy of all indices
func (wm *WebsiteMap) GetIndices() []string {
	var indices []string

	wm.mux.Lock()
	defer wm.mux.Unlock()

	for idx, w := range wm.W {
		if w != nil {
			indices = append(indices, idx)
		}
	}

	return indices
}

// SearchKeyword search for all the websites that have this keyword
func (wm *WebsiteMap) SearchKeyword(keyword string) []*Website {
	var websites []*Website

	wm.mux.Lock()
	for _, website := range wm.W {
		if utils.Contains(website.GetKeywords(), keyword) {
			websites = append(websites, website)
		}
	}
	wm.mux.Unlock()

	return websites
}

// RemovePeer removes a peer from all websites' seeders list
func (wm *WebsiteMap) RemovePeer(peer *Peer) {
	wm.mux.Lock()
	defer wm.mux.Unlock()

	for _, w := range wm.W {
		w.Seeders.Remove(peer)
	}
}

// Website

// GetKeywords return the list of keywords of a website
func (w *Website) GetKeywords() []string {
	return w.Keywords
}

// SetKeywords sets the keywords for the Website
func (w *Website) SetKeywords(keywords []string) {
	w.Keywords = keywords
}

// AddSeeder adds a seeder for this particular website
func (w *Website) AddSeeder(peer *Peer) {
	w.Seeders.Add(peer)
}

// GetSeeders gets all the seeders
func (w *Website) GetSeeders() []Peer {
	return w.Seeders.GetAll()
}

// IncVersion increment the version of a Website by 1
func (w *Website) IncVersion() {
	w.Version++
}

// SaveMetadata write/overwrite a metadata file in the website folder
func (w *Website) SaveMetadata() {
	jsonData, err := json.Marshal(w)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(utils.MetadataDir+w.Name, jsonData, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// IsOurs checks if the private key for this website is present which means this
// node owns the website
func (w *Website) IsOurs() bool {
    _, err := os.Stat(utils.KeyDir + w.Name)
    return (err == nil)
}

// Sign scans the website folder hashing all files in order to create the
// contents.json file with the website's signature
func (w *Website) Sign() {
    var hashes []byte
    var contents map[string]string = make(map[string]string)

    err := filepath.Walk(utils.WebsiteDir + w.Name, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.Mode().IsRegular() {
            data, err := ioutil.ReadFile(path)
            if err != nil {
                return err
            }
            hash := sha256.Sum256(data)
            hashes = append(hashes, hash[:]...)
            contents[path] = hex.EncodeToString(hash[:])
        }
        return nil
    })
    utils.CheckError(err)

    privKey := w2pcrypto.LoadPrivateKey(w.Name)
    sig := privKey.SignMessage(hashes)
    contents["signature"] = sig

    jsonData, err := json.Marshal(contents)
    utils.CheckError(err)

    path := utils.WebsiteDir + w.Name + "/contents.json"
    err = ioutil.WriteFile(path, jsonData, 0600)
    utils.CheckError(err)
}

func (w *Website) Verify() bool {
    var hashes []byte
    var contents map[string]string = make(map[string]string)

    path := utils.WebsiteDir + w.Name + "/contents.json"
    data, err := ioutil.ReadFile(path)
    utils.CheckError(err)

    err = json.Unmarshal(data, contents)
    utils.CheckError(err)

    // Verifying each file's hash
    err = filepath.Walk(utils.WebsiteDir + w.Name, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.Mode().IsRegular() {
            data, err := ioutil.ReadFile(path)
            if err != nil {
                return err
            }

            hash := sha256.Sum256(data)
            hashStr := hex.EncodeToString(hash[:])
            if hashStr != contents[path] {
                err = errors.New("VerificationError")
                return err
            }

            hashes = append(hashes, hash[:]...)
            contents[path] = hex.EncodeToString(hash[:])
        }
        return nil
    })
    if err != nil {
        return false
    }

    // Verifying signature
    return w.PubKey.VerifySignature(hashes, contents["signature"])
}

// Bundle creates a compressed archive of a website folder for seeding
func (w *Website) Bundle() {
	file, err := os.Create(utils.SeedDir + w.Name)
	defer file.Close()
    utils.CheckError(err)

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	err = filepath.Walk(utils.WebsiteDir+w.Name, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = path

		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		// if not a file (e.g. a dir) don't copy content of it
		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path)
		defer f.Close()
		if err != nil {
			return err
		}

		_, err = io.Copy(tw, f)
		if err != nil {
			return err
		}

		return nil
	})
    utils.CheckError(err)
}

// Unbundle uncompress and unarchive a website to display it
func (w *Website) Unbundle() {
	archive, err := os.Open(utils.SeedDir + w.Name)
	defer archive.Close()
	if err != nil {
		log.Fatal(err)
	}

	gzr, err := gzip.NewReader(archive)
	defer gzr.Close()
	if err != nil {
		log.Fatal(err)
	}

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		target := filepath.Join(utils.WebsiteDir+w.Name, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			_, err := os.Stat(target)
			if err != nil {
				err = os.MkdirAll(target, 0755)
				if err != nil {
					log.Fatal(err)
				}
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			defer f.Close()
			if err != nil {
				log.Fatal(err)
			}

			_, err = io.Copy(f, tr)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// GenPieces generates the pieces from the website archive and set it in
// the Website object
func (w *Website) GenPieces(pieceLength int) {
	w.PieceLength = pieceLength

	data, err := ioutil.ReadFile(utils.SeedDir + w.Name)
	if err != nil {
		log.Fatal(err)
	}

	rest := data
	var chunk []byte
	var pieces string
	for i := pieceLength; i < len(data); i += pieceLength {
		chunk = rest[:i]

		sum := sha256.Sum256(chunk)
		hash := hex.EncodeToString(sum[:])
		pieces = pieces + hash

		rest = rest[i:]
	}

	sum := sha256.Sum256(rest)
	hash := hex.EncodeToString(sum[:])
	pieces = pieces + hash

	w.Pieces = pieces
}
