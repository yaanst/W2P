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

// WebsiteMap is a map from a Website hash to the Website
type WebsiteMap map[string]*Website

// Website is a structure that represents a website
type Website struct {
	Name         string
	Seeders      *Peers
	Keywords     []string
	PubKey       string
	PiecesLength int
	Pieces       string
	Version      int
}

// -----------
// - Methods -
// -----------

// PeerEquals compare two peers
func PeerEquals(p1, p2 *Peer) bool {
	return p1.IP.String() == p2.IP.String() && p1.Port == p2.Port
}

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
