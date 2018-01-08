package node

import (
	"fmt"
	"github.com/yaanst/W2P/structs"
	"strings"
	//	"github.com/yaanst/W2P/comm"
)

// Node is the main componoment which will handle message, update its state,
// send messages etc...
type Node struct {
	Addr       *structs.Peer
	Peers      *structs.Peers
	WebsiteMap *structs.WebsiteMap
}

// NewWebsite creates a new Website object using the directory it is in
func (n *Node) NewWebsite(path string) *structs.Website {
	splitPath := strings.Split(path, "/")
	name := splitPath[len(splitPath)-1]

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
