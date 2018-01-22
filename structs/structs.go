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
	mux sync.RWMutex
	P   []*Peer
}

// WebsiteMap is a map from a Website PubKey to the Website
type WebsiteMap struct {
	mux sync.RWMutex
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
	R   map[string]*Peer
}

// Counter is a simple async counter
type Counter struct {
	mux sync.RWMutex
	C   int
}

// ----------------
// - Constructors -
// ----------------

// ParsePeer construct a Peer from a string of format "addr:port"
func ParsePeer(peerString string) *Peer {
	udpAddr, err := net.ResolveUDPAddr("udp4", peerString)
	utils.CheckError(err)
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
	utils.CheckError(err)

	website := &Website{}
	err = json.Unmarshal(jsonData, website)
	utils.CheckError(err)

	return website
}

// NewRoutingTable constructs a RoutingTable object
func NewRoutingTable() *RoutingTable {
	return &RoutingTable{
		R: make(map[string]*Peer),
	}
}

// NewCounter constructs a Counter object
func NewCounter() *Counter {
	return &Counter{
		C: 0,
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
	peers.mux.RLock()
	defer peers.mux.RUnlock()
	for _, p := range peers.P {
		if PeerEquals(p, peer) {
			return true
		}
	}
	return false
}

// Add adds a Peer to the Peers if not already present
func (peers *Peers) Add(peer *Peer) {
    set := make(map[string]bool)
    var newPeers []*Peer

    peers.mux.Lock()
	defer peers.mux.Unlock()

    // remove duplicates
    for _, p := range peers.P {
        set[p.String()] = true
    }
    set[peer.String()] = true

    // rebuild slice
    for p := range set {
        newPeers = append(newPeers, ParsePeer(p))
    }
	peers.P = newPeers
}

// Remove removes a Peer from the Peers
func (peers *Peers) Remove(peer *Peer) {
	peers.mux.Lock()
    defer peers.mux.Unlock()
	for i, p := range peers.P {
		if PeerEquals(p, peer) {
			// cut slice in 2 part and append without i-th element inbetween
			peers.P = append(peers.P[:i], peers.P[i+1:]...)
			break
		}
	}
}

// GetAll returns a copy of the list of Peer
func (peers *Peers) GetAll() []Peer {
	var peerList []Peer
	peers.mux.RLock()
	defer peers.mux.RUnlock()
	for _, p := range peers.P {
		peerList = append(peerList, *p)
	}

	return peerList
}

// Count returns the number of peers
func (peers *Peers) Count() int {
	peers.mux.RLock()
	defer peers.mux.RUnlock()
	return len(peers.P)
}

// WebsiteMap

// Set adds/updates a website to the website map
func (wm *WebsiteMap) Set(website *Website) {
	wm.mux.Lock()
    defer wm.mux.Unlock()
	wm.W[website.Name] = website
}

// Get return the Website struct given its name
func (wm *WebsiteMap) Get(name string) *Website {
	wm.mux.RLock()
	defer wm.mux.RUnlock()
	website := wm.W[name]

	return website
}

// GetIndices returns of copy of all indices
func (wm *WebsiteMap) GetIndices() []string {
	var indices []string

	wm.mux.RLock()
	defer wm.mux.RUnlock()

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

	wm.mux.RLock()
	defer wm.mux.RUnlock()
	for _, website := range wm.W {
		if utils.Contains(website.GetKeywords(), keyword) {
			websites = append(websites, website)
		}
	}

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

// Count returns the number of websites stored in the WebsiteMap
func (wm *WebsiteMap) Count() int {
    wm.mux.RLock()
    defer wm.mux.RUnlock()
    return len(wm.W)
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

// DiffSeeders returns the seeders present in a website and not the other
func (w *Website) DiffSeeders(rWeb *Website) []*Peer {
    var seeders []*Peer

    for _, s := range rWeb.Seeders.GetAll() {
        if !w.Seeders.Contains(&s) {
            seeders = append(seeders, &s)
        }
    }
    for _, s := range w.Seeders.GetAll() {
        if !rWeb.Seeders.Contains(&s) {
            seeders = append(seeders, &s)
        }
    }
    return seeders
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
	utils.CheckError(err)

	err = ioutil.WriteFile(utils.MetadataDir+w.Name, jsonData, 0644)
	utils.CheckError(err)
}

// Owned checks if the private key for this website is present which means this
// node owns the website
func (w *Website) Owned() bool {
	_, err := os.Stat(utils.KeyDir + w.Name)
	return (err == nil)
}

// Sign scans the website folder hashing all files in order to create the
// contents.json file with the website's signature
func (w *Website) Sign() {
	var hashes []byte
	var contents = make(map[string]string)

	err := filepath.Walk(utils.WebsiteDir+w.Name, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() && info.Name() != "contents.json" {
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

// Verify verifies if the Website is signed by the owner
func (w *Website) Verify() bool {
	var hashes []byte
	var contents = make(map[string]string)

	path := utils.WebsiteDir + w.Name + "/contents.json"
	data, err := ioutil.ReadFile(path)
	utils.CheckError(err)

	err = json.Unmarshal(data, &contents)
	utils.CheckError(err)

	// Verifying each file's hash
	err = filepath.Walk(utils.WebsiteDir+w.Name, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() && info.Name() != "contents.json" {
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

	target := utils.WebsiteDir + w.Name

	err = filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = path

		log.Println("[BUNDLE]\t\tBundling '" + header.Name + "'")

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
	utils.CheckError(err)

	gzr, err := gzip.NewReader(archive)
	defer gzr.Close()
	utils.CheckError(err)

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		utils.CheckError(err)

		target := header.Name
		log.Println("[UNBUNDLE]\t\tUnbundling '" + target + "'")

		switch header.Typeflag {
		case tar.TypeDir:
			_, err := os.Stat(target)
			if err != nil {
				err = os.MkdirAll(target, 0755)
				utils.CheckError(err)
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			defer f.Close()
			utils.CheckError(err)

			_, err = io.Copy(f, tr)
			utils.CheckError(err)
		}
	}
}

// GenPieces generates the pieces from the website archive and set it in
// the Website object
func (w *Website) GenPieces(pieceLength int) {
	w.PieceLength = pieceLength

	data, err := ioutil.ReadFile(utils.SeedDir + w.Name)
	utils.CheckError(err)

	dataSize := len(data)
	var chunk []byte
	var pieces string

	for i := 0; i < len(data); i += pieceLength {
		offsetStart := i
		offsetEnd := i + pieceLength
		if offsetEnd >= dataSize {
			offsetEnd = dataSize
		}
		chunk = data[offsetStart:offsetEnd]

		sum := sha256.Sum256(chunk)
		hash := hex.EncodeToString(sum[:])
		pieces = pieces + hash
	}

	w.Pieces = pieces
}

// ClearSeeders removes all seeders for a website
func (w *Website) ClearSeeders() {
    w.Seeders.mux.Lock()
    defer w.Seeders.mux.Unlock()
    w.Seeders.P = make([]*Peer, 0, 5)
}


// Counter

// Read returns the current value
func (c *Counter) Read() int {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.C
}

// Inc adds 1 to the current value
func (c *Counter) Inc() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.C++
}

// Dec substracts 1 from the current value
func (c *Counter) Dec() {
	c.mux.Lock()
	defer c.mux.Unlock()
	if (c.C - 1) < 0 {
		c.C = 0
	} else {
		c.C--
	}
}

// Routing table

// Get returns the value peer through which to send the packet or dst
func (rt *RoutingTable) Get(dst string) *Peer {
	rt.mux.Lock()
	defer rt.mux.Unlock()
	via := rt.R[dst]
	if via == nil {
		via = ParsePeer(dst)
	}
	return via
}

// Set adds a new entry or updates an existing one in the routing table
func (rt *RoutingTable) Set(dst string, via *Peer) {
	rt.mux.Lock()
	defer rt.mux.Unlock()
	rt.R[dst] = via
}
