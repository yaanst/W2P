package main

import (
    "fmt"
    //"net"
    //"log"
    "flag"
    //"crypto"
    //"io/ioutil"
    //"github.com/dedis/protobuf"
    //"github.com/yaanst/W2P/w2pcrypto"
)

func main() {
    client_port := flag.String("client-port", "4000", "Port on which you connect the browser")
    peers := flag.String("peers", "", "Comma-separated list of peers in the form of IP:PORT")
    public_addr := flag.String("public-addr", "", "Address used to communicate with other peers in the form of IP:PORT")

    fmt.Println(client_port, peers, public_addr)
}
