package main

import (
	"flag"
	"log"
	"time"

	"github.com/yaanst/W2P/node"
    "github.com/husobee/vestigo"
)

func ServeWebsite(w http.ResponseWriter, r *http.Request) {
    name := vestigo.Param(r, "name")
    path := utils.WebsiteDir + name

    fs := http.FileServer(http.Dir(path))
    fs.ServeHTTP(w,r)
}

func main() {
	//clientPort := flag.String("client-port", "4000", "Port on which you connect the browser")
	//publicAddr := flag.String("public-addr", "", "Address used to communicate with other peers in the form of IP:PORT")
	var name, addr, peers string
	flag.StringVar(&name, "name", "test", "Name of the node")
	flag.StringVar(&addr, "addr", "", "Address of the node format IP:PORT")
	flag.StringVar(&peers, "peers", "", "Comma-separated list of peers in the form of IP:PORT")
	flag.Parse()

	log.Println("arg name:", name)
	log.Println("arg addr:", addr)
	log.Println("arg peers:", peers)

	node := node.NewNode(name, addr, peers)
	node.Init()

	go node.AntiEntropy(time.Second)

	node.Listen()

    router := vestigo.NewRouter()
    router.Get("/", http:FileServer(http.Dir("ui/website")))
    router.Get("/w/:name", ui.Ser
}
