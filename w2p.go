package main

import (
	"flag"
	"fmt"

	"github.com/yaanst/W2P/node"
)

func main() {
	//clientPort := flag.String("client-port", "4000", "Port on which you connect the browser")
	//publicAddr := flag.String("public-addr", "", "Address used to communicate with other peers in the form of IP:PORT")
	var name, addr, peers string
	flag.StringVar(&peers, "peers", "", "Comma-separated list of peers in the form of IP:PORT")
	flag.StringVar(&addr, "addr", "", "Comma-separated list of peers in the form of IP:PORT")
	flag.StringVar(&name, "name", "", "Comma-separated list of peers in the form of IP:PORT")

	fmt.Println("args: ", name, addr, peers)

	node := node.NewNode(name, addr, peers)
	node.Init()

	node.Listen()
}
