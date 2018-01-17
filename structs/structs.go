package structs

import (
	"io"
	"os"
	"fmt"
	"net"
	"sync"
	"strings"
	"io/ioutil"
	"archive/tar"
	"encoding/hex"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"path/filepath"

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
	R   map[string]*Peer
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

	var website *Website
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
    utils.CheckError(err)

	err = ioutil.WriteFile(utils.MetadataDir+w.Name, jsonData, 0644)
    utils.CheckError(err)
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

		target := filepath.Join(utils.WebsiteDir+w.Name, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			_, err := os.Stat(target)
			if err != nil {
				err = os.MkdirAll(target, 0755)
                utils.CheckError(err)
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
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

// Routing table

// Get returns the value associated to dst or nil
func (rt *RoutingTable) Get(dst string) *Peer {
    rt.mux.Lock()
    defer rt.mux.Unlock()
    return rt.R[dst]
}

// Set adds a new entry or updates an existing one in the routing table
func (rt *RoutingTable) Set(dst string, via *Peer) {
    rt.mux.Lock()
    defer rt.mux.Unlock()
    rt.R[dst] = via
}

// Delete removes an element from the routing table
func (rt *RoutingTable) Delete(dst string) {
    rt.mux.Lock()
    defer rt.mux.Unlock()
    delete(rt.R, dst)
}
