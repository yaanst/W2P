package structs

import (
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
	mux sync.Mutex
	p   []*Peer
}

// WebsiteMap is a map from a Website PubKey to the Website
type WebsiteMap struct {
	mux sync.Mutex
	w   map[string]*Website
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

// -----------
// - Methods -
// -----------

// Peer

// PeerEquals compare two peers
func PeerEquals(p1, p2 *Peer) bool {
	return p1.IP.String() == p2.IP.String() && p1.Port == p2.Port
}

// Peers

// Contains check if the Peers slice contains a Peer
func (peers *Peers) Contains(peer *Peer) bool {
	peers.mux.Lock()
	for _, p := range peers.p {
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
		peers.p = append(peers.p, peer)
		peers.mux.Unlock()
	}
}

// Remove removes a Peer from the Peers
func (peers *Peers) Remove(peer *Peer) {
	peers.mux.Lock()
	for i, p := range peers.p {
		if PeerEquals(p, peer) {
			// cut slice in 2 part and append without i-th element inbetween
			peers.p = append(peers.p[:i], peers.p[i+1:]...)
			break
		}
	}
	peers.mux.Unlock()
}

// WebsiteMap

// Set adds/updates a website to the website map
func (wm *WebsiteMap) Set(website *Website) {
	wm.mux.Lock()
	wm.w[website.Name] = website
	wm.mux.Unlock()
}
