package main

import (
	"log"
	"flag"
	"time"

	"github.com/yaanst/W2P/ui"
	"github.com/yaanst/W2P/node"
)

func main() {
	var name, addr, peers, uiPort string
	flag.StringVar(&name, "name", "test", "Name of the node")
	flag.StringVar(&addr, "addr", "", "Address of the node format IP:PORT")
	flag.StringVar(&peers, "peers", "", "Comma-separated list of peers in the form of IP:PORT")
	flag.StringVar(&uiPort, "uiPort", "8000", "Port for the browser based UI")
	flag.Parse()

	log.Println("arg name:", name)
	log.Println("arg addr:", addr)
	log.Println("arg peers:", peers)

	node := node.NewNode(name, addr, peers)
	node.Init()

	go node.AntiEntropy(5 * time.Second)

	go node.Listen()

    ui.StartServer(uiPort, node)
}
