package structs

import (
	"fmt"
	"log"
	"net"
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
	Mux sync.Mutex
	P   []*Peer
}

// WebsiteMap is a map from a Website PubKey to the Website
type WebsiteMap struct {
	Mux sync.Mutex
	W   map[string]*Website
}

// Website is a structure that represents a website
type Website struct {
	Name        string
	Seeders     *Peers
	Keywords    []string
	PubKey      string
	PieceLength int
	Pieces      string
	Version     int
}

// RoutingTable is a table which keeps in memory possible route for a dest
// if a Peer is not directly reachable
type RoutingTable struct {
	Mux sync.Mutex
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
	addrList := strings.Split(peersString, ",")

	peers := NewPeers()

	for _, addr := range addrList {
		peer := ParsePeer(addr)
		peers.Add(peer)
	}

	return peers
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
func NewWebsite(name string, seeders *Peers, keywords []string) *Website {
	pieceLength := utils.DefaultPieceLength

	pieces := ""

	privKey, pubKey := w2pcrypto.CreateKey()

	// TODO implement

	return &Website{
		Name:        name,
		Seeders:     seeders,
		Keywords:    keywords,
		PieceLength: pieceLength,
		Version:     1,
	}
}

// LoadWebsite constructs a Website from a metadata file
func LoadWebsite(name string) *Website {

	// TODO implement

	return &Website{}
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
	peers.Mux.Lock()
	for _, p := range peers.P {
		if PeerEquals(p, peer) {
			peers.Mux.Unlock()
			return true
		}
	}
	peers.Mux.Unlock()
	return false
}

// Add adds a Peer to the Peers if not already present
func (peers *Peers) Add(peer *Peer) {
	if !peers.Contains(peer) {
		peers.Mux.Lock()
		peers.P = append(peers.P, peer)
		peers.Mux.Unlock()
	}
}

// Remove removes a Peer from the Peers
func (peers *Peers) Remove(peer *Peer) {
	peers.Mux.Lock()
	for i, p := range peers.P {
		if PeerEquals(p, peer) {
			// cut slice in 2 part and append without i-th element inbetween
			peers.P = append(peers.P[:i], peers.P[i+1:]...)
			break
		}
	}
	peers.Mux.Unlock()
}

// WebsiteMap

// Set adds/updates a website to the website map
func (wm *WebsiteMap) Set(website *Website) {
	wm.Mux.Lock()
	wm.W[website.Name] = website
	wm.Mux.Unlock()
}

// Website

// Save write/overwrite a metadata file in the website folder
func (w *Website) Save() {
}
