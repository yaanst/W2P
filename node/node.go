package node

import (
	"fmt"
	"github.com/yaanst/W2P/structs"
	//	"github.com/yaanst/W2P/comm"
)

// Node is the main componoment which will handle message, update its state,
// send messages etc...
type Node struct {
	Peers      structs.Peers
	WebsiteMap structs.WebsiteMap
}

func main() {
	fmt.Println("vim-go")
}
