package comm

import (
	"github.com/yaanst/W2P/structs"
)

// Message is the basic struct that will be exchanged throughout the network
type Message struct {
	Orig *structs.Peer
	Dest *structs.Peer
	Meta *Meta
	Data *Data
}

// Data are the messages containing binary data for file exchange
type Data struct {
	ID   string
	Data []byte
}

// Meta are the messages containing the information about all websites
type Meta struct {
	WebsiteMap structs.WebsiteMap
}
