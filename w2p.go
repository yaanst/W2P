package main

import (
	"flag"
	"log"
	"time"

	"github.com/yaanst/W2P/node"
)

func main() {
	//clientPort := flag.String("client-port", "4000", "Port on which you connect the browser")
	//publicAddr := flag.String("public-addr", "", "Address used to communicate with other peers in the form of IP:PORT")
	var name, addr, peers string
	flag.StringVar(&peers, "peers", "", "Comma-separated list of peers in the form of IP:PORT")
	flag.StringVar(&addr, "addr", "", "Address of the node format IP:PORT")
	flag.StringVar(&name, "name", "", "Name of the node")

	log.Println("arg name: " + name)
	log.Println("arg addr: " + addr)
	log.Println("arg peers: " + peers)

	node := node.NewNode(name, addr, peers)
	node.Init()

	go node.AntiEntropy(time.Second)

	node.Listen()
}
