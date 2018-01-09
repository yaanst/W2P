package node

import (
	"os"
	"strings"

	"github.com/yaanst/W2P/structs"
	"github.com/yaanst/W2P/utils"
	//"github.com/yaanst/W2P/w2pcrypto"
	//	"github.com/yaanst/W2P/comm"
)

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

// Init initialize a Node adding website already present on disk and checking
// wether we have their metadata, also checking every dir is present
func (n *Node) Init() {
	websitesNames := utils.ScanDir(utils.WebsiteDir)

	for _, name := range websitesNames {
		if _, err := os.Stat(utils.MetadataDir + name); err == nil {
			n.AddWebsite(name)
		}
	}
}

// AddWebsite constructs a Website object that already has a metadatafile and
// adds it to the WebsiteMap
func (n *Node) AddWebsite(name string) {
	website := structs.LoadWebsite(name)

	n.WebsiteMap.Set(website)
}

// AddNewWebsite constructs a new website that has no metadatafile and adds
// it to the WebsiteMap
func (n *Node) AddNewWebsite(name string, keywords []string) {
	website := structs.NewWebsite(name, keywords)

	website.Bundle()

	website.GenPieces(utils.DefaultPieceLength)
	website.Seeders.Add(n.Addr)
	website.SaveMetadata()

	n.WebsiteMap.Set(website)
}

// UpdateWebsite update a Website in the WebsiteMap when user modified
// his website
func (n *Node) UpdateWebsite(name string, keywords []string) {
	website := n.WebsiteMap.Get(name)

	website.Bundle()

	website.SetKeywords(keywords)
	website.GenPieces(utils.DefaultPieceLength)
	website.IncVersion()
	website.SaveMetadata()

}

// Search search for keywords match among all the websites on the network
func (n *Node) Search(search string) []string {
	terms := strings.Split(search, " ")

	var websites []*structs.Website
	for _, term := range terms {
		websites = append(websites, n.WebsiteMap.SearchKeyword(term)[:]...)
	}

	var results []string
	for _, w := range websites {
		results = append(results, w.Name)
	}

	return results
}
