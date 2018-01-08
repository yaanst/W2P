package comm

import (
	"bytes"
	"encoding/gob"
	"log"
	"net"

	"github.com/yaanst/W2P/structs"
)

// -----------
// - Structs -
// -----------

// Message is the basic struct that will be exchanged throughout the network
type Message struct {
	Orig *structs.Peer
	Dest *structs.Peer
	Meta *Meta
	Data *Data
}

// Data are the messages containing binary data for file exchange
type Data struct {
	Website string
	Piece   string
	Data    []byte
}

// Meta are the messages containing the information about all websites
type Meta struct {
	WebsiteMap *structs.WebsiteMap
}

// ----------------
// - Constructors -
// ----------------

// NewDataRequest construct a data request for a piece in a specific website
func NewDataRequest(orig, dest *structs.Peer, website, piece string) *Message {
	data := &Data{
		Website: website,
		Piece:   piece,
	}

	return &Message{
		Orig: orig,
		Dest: dest,
		Data: data,
	}
}

// NewDataReply construct a reply to a data request with a piece in a specific website
func NewDataReply(request *Message, data []byte) *Message {
	dataMessage := &Data{
		Website: request.Data.Website,
		Piece:   request.Data.Piece,
		Data:    data,
	}

	return &Message{
		Orig: request.Dest,
		Dest: request.Orig,
		Data: dataMessage,
	}
}

// NewMeta construct a Message that contains a WebsiteMap
func NewMeta(orig, dest *structs.Peer, wm *structs.WebsiteMap) *Message {
	meta := &Meta{
		WebsiteMap: wm,
	}

	return &Message{
		Orig: orig,
		Dest: dest,
		Meta: meta,
	}
}

// NewHeartbeat construct a simple heartbeat message
func NewHeartbeat(orig, dest *structs.Peer) *Message {
	return &Message{
		Orig: orig,
		Dest: dest,
	}
}

// -----------
// - Methods -
// -----------

// Send sends a message to the Peer at dest (dest is NOT final destination)
func (m *Message) Send(conn *net.UDPConn, dest *structs.Peer) {
	b := EncodeMessage(m)

	conn.WriteToUDP(b, dest)
}

// EncodeMessage serializes a Message in order to send it
func EncodeMessage(m *Message) []byte {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)

	err := e.Encode(m)
	if err != nil {
		log.Fatal(err)
	}

	return b.Bytes()
}

// DecodeMessage deserializes a Message in order to receive it
func DecodeMessage(b []byte) *Message {
	m := &Message{}

	bb := bytes.Buffer{}
	bb.Write(b)

	d := gob.NewDecoder(&bb)

	err := d.Decode(m)
	if err != nil {
		log.Fatal(err)
	}

	return m
}
