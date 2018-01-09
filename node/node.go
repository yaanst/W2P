package node

import (
	"fmt"
	"github.com/yaanst/W2P/structs"
	//	"github.com/yaanst/W2P/comm"
)

// ---------
// - Const -
// ---------

// WebsiteDir is the path to the directory containing all website
const WebsiteDir string = "./website/"

// SeedDir is the path to the directory containing all seeding binary archive
const SeedDir string = "./seed/"

// -----------
// - Structs -
// -----------

// Node is the main componoment which will handle message, update its state,
// send messages etc...
type Node struct {
	Name         string
	Addr         *structs.Peer
	Peers        *structs.Peers
	RoutingTable *structs.RoutingTable
	WebsiteMap   *structs.WebsiteMap
}

// ----------------
// - Constructors -
// ----------------

// NewNode construct a fresh new Node (no rt and no wm)
func NewNode(name, addrString, peersString string) *Node {
	addr := structs.ParsePeer(addrString)
	peers := structs.ParsePeers(peersString)
	rt := structs.NewRoutingTable()
	wm := structs.NewWebsiteMap()

	return &Node{
		Name:         name,
		Addr:         addr,
		Peers:        peers,
		RoutingTable: rt,
		WebsiteMap:   wm,
	}
}

// -----------
// - Methods -
// -----------

// Init initialize a Node adding website already present
func (n *Node) Init() {
	// TODO implement
}

// NewWebsite creates a new Website object using the directory it is in
func (n *Node) NewWebsite(name string) *structs.Website {
	seeders := structs.NewPeers()
	seeders.Add(n.Addr)

	// TODO finish implementation

	return &structs.Website{
		Name:    name,
		Seeders: seeders,
	}
}

func main() {
	fmt.Println("vim-go")
}
