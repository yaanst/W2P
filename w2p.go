package main

import (
    "fmt"
    "flag"
    "github.com/yaanst/W2P/w2pcrypto"
)

func main() {
    client_port := flag.String("client-port", "4000", "Port on which you connect the browser")
    peers := flag.String("peers", "", "Comma-separated list of peers in the form of IP:PORT")
    public_addr := flag.String("public-addr", "", "Address used to communicate with other peers in the form of IP:PORT")

    fmt.Println("args;",client_port, peers, public_addr)

    // Examples
    privkey := w2pcrypto.CreateKey()
    fmt.Println("key1: %+v\n",privkey)
    w2pcrypto.SaveKey("first.key", privkey)

    privkey2 := w2pcrypto.LoadKey("first.key")
    fmt.Println("key2: %+v\n",privkey2)
}
