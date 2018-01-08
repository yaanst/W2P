package structs

import (
	"fmt"
	"net"
	"sync"
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

// NewPeers constructs a new Peers object (list of peer with a mutex)
func NewPeers() *Peers {
	return &Peers{
		P: make([]*Peer, 5),
	}
}

// NewWebsiteMap constructs a WebsiteMap object
func NewWebsiteMap() *WebsiteMap {
	return &WebsiteMap{
		W: make(map[string]*Website),
	}
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
