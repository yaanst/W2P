package node

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yaanst/W2P/comm"
	"github.com/yaanst/W2P/structs"
	"github.com/yaanst/W2P/utils"
)

// -----------
// - Structs -
// -----------

// Node is the main componoment which will handle message, update its state,
// send messages etc...
type Node struct {
	Name         string
	Addr         *structs.Peer
	Conn         *net.UDPConn
	Peers        *structs.Peers
	RoutingTable *structs.RoutingTable
	WebsiteMap   *structs.WebsiteMap
}

// ----------------
// - Constructors -
// ----------------

// NewNode construct a fresh new Node (enmpty rt and no wm)
func NewNode(name, addrString, peersString string) *Node {
	addr := structs.ParsePeer(addrString)
	peers := structs.ParsePeers(peersString)
	rt := structs.NewRoutingTable()
	wm := structs.NewWebsiteMap()

	connAddr := net.UDPAddr(*addr)
	conn, err := net.ListenUDP("udp4", &connAddr)
	utils.CheckError(err)

	return &Node{
		Name:         name,
		Addr:         addr,
		Conn:         conn,
		Peers:        peers,
		RoutingTable: rt,
		WebsiteMap:   wm,
	}
}

// -----------
// - Methods -
// -----------

// Init initialize a Node adding website already present on disk and checking
// wether we have their metadata, also checking every dir is present
func (n *Node) Init() {
	if _, err := os.Stat(utils.MetadataDir); err != nil {
		err := os.MkdirAll(utils.MetadataDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(utils.SeedDir); err != nil {
		err := os.MkdirAll(utils.SeedDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(utils.WebsiteDir); err != nil {
		err := os.MkdirAll(utils.WebsiteDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(utils.KeyDir); err != nil {
		err := os.MkdirAll(utils.KeyDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	websitesNames := utils.ScanDir(utils.WebsiteDir)

	for _, name := range websitesNames {
		if _, err := os.Stat(utils.MetadataDir + name); err == nil {
			n.AddWebsite(name)
		}
	}
}

// AddWebsite constructs a Website object that already has a metadatafile and
// adds it to the WebsiteMap
func (n *Node) AddWebsite(name string) {
	website := structs.LoadWebsite(name)

	n.WebsiteMap.Set(website)
}

// AddNewWebsite constructs a new website that has no metadatafile and adds
// it to the WebsiteMap
func (n *Node) AddNewWebsite(name string, keywords []string) {
	website := structs.NewWebsite(name, keywords)

    if website != nil && website.IsOurs() {
        website.Sign()
        website.Bundle()

        website.GenPieces(utils.DefaultPieceLength)
        website.Seeders.Add(n.Addr)
        website.SaveMetadata()

        n.WebsiteMap.Set(website)
    }
}

// UpdateWebsite update a Website in the WebsiteMap when user modified
// his website
func (n *Node) UpdateWebsite(name string, keywords []string) bool {
    website := n.WebsiteMap.Get(name)

    if website != nil && website.IsOurs() {
        website.Sign()
        website.Bundle()

        website.SetKeywords(keywords)
        website.GenPieces(utils.DefaultPieceLength)
        website.IncVersion()
        website.SaveMetadata()
        return true
    }
    return false
}

// SendWebsiteMap shares the node's WebsiteMap with other nodes
func (n *Node) SendWebsiteMap() {
	//TODO: Send to ALL or subset or random but more frequently?
	for _, p := range n.Peers.GetAll() {
		message := comm.NewMeta(n.Addr, &p, n.WebsiteMap)
		message.Send(n.Conn, &p)
		log.Println("[SENT] WebsiteMap to", p.String())
	}
}

// HeartBeat sends a hearbeat message to peer and waits for an answer or timeout
func (n *Node) HeartBeat(peer *structs.Peer, reachable chan bool) {
	// Create a random local address for a new connection
	tempPeer := &structs.Peer{
		IP:   n.Addr.IP,
		Port: n.Addr.Port + 100 + rand.Intn(10000),
		Zone: n.Addr.Zone,
	}
	lAddr := net.UDPAddr(*tempPeer)

	conn, err := net.ListenUDP("udp4", &lAddr)
	defer conn.Close()
	utils.CheckError(err)

	conn.SetReadDeadline(time.Now().Add(utils.HeartBeatTimeout))

	message := comm.NewHeartbeat(tempPeer, peer)
	buffer := make([]byte, utils.HeartBeatBufferSize)

	message.Send(conn, peer)

	log.Println("[SENT] Heartbeat to", peer.String())

	size, err := conn.Read(buffer)
	if err != nil {
		reachable <- false
	} else if size > 0 {
		reachable <- true
	}
	return
}

// CheckPeer checks if peer is up and removes it from every location if not
func (n *Node) CheckPeer(peer *structs.Peer) {
	c := make(chan bool)
	go n.HeartBeat(peer, c)

	reachable := <-c
	if !reachable {
		log.Println("[HEARTBEAT] Peer", peer, "is down")
		n.Peers.Remove(peer)
		n.WebsiteMap.RemovePeer(peer)
	} else {
		log.Println("[HEARTBEAT] Peer", peer, "is up")
		n.Peers.Add(peer)
	}
}

// MergeWebsiteMap merges a WebsiteMap into the local one
func (n *Node) MergeWebsiteMap(remoteWM *structs.WebsiteMap) {
	localWM := n.WebsiteMap

	rIndices := remoteWM.GetIndices()
	for _, rKey := range rIndices {
		lWeb := localWM.Get(rKey)
		rWeb := remoteWM.Get(rKey)

        if lWeb.PubKey.String() != rWeb.PubKey.String() {
            log.Fatalf("Public keys not matching for local/remote website %v\n", lWeb.Name)
        }


		if lWeb != nil {
            log.Print("[WEBSITEMAP] Updating website", lWeb.Name)
			if rWeb.Version > lWeb.Version {
                lWeb.Version = rWeb.Version
				lWeb.SetKeywords(rWeb.GetKeywords())
				lWeb.Pieces = rWeb.Pieces

				// Add missing seeders for lWeb
				for _, rPeer := range rWeb.GetSeeders() {
					if !lWeb.Seeders.Contains(&rPeer) {
						lWeb.Seeders.Add(&rPeer)
					}
				}
                // Remove extra seeders for lWeb
				for _, lPeer := range lWeb.GetSeeders() {
					if !rWeb.Seeders.Contains(&lPeer) {
						lWeb.Seeders.Remove(&lPeer)
					}
				}
			}
		} else {
            log.Print("[WEBSITEMAP] Adding website", rWeb)
            localWM.Set(rWeb)
        }
	}
}

// Listen listens for messages from other peers and acts on them
func (n *Node) Listen() {
	buffer := make([]byte, utils.ListenBufferSize)

	log.Println("[LISTENING] on", n.Addr.String())

	for {
		_, _, err := n.Conn.ReadFromUDP(buffer)
		utils.CheckError(err)

		message := comm.DecodeMessage(buffer)
		orig := message.Orig
		dest := message.Dest

		// Forward message
		if !structs.PeerEquals(dest, n.Addr) {
			//TODO routing table
			continue
		}

		// HeartBeat
		if message.Meta == nil && message.Data == nil {
			log.Println("[RECEIVE] Heartbeat from " + orig.String())
			heartbeat := comm.NewHeartbeat(n.Addr, orig)
			heartbeat.Send(n.Conn, orig) //TODO use routing table

			// WebsiteMapUpdate
		} else if message.Meta != nil {
			log.Println("[RECEIVE] WebsiteMap from " + orig.String())
			go n.CheckPeer(orig)
			go n.MergeWebsiteMap(message.Meta.WebsiteMap)

			// Data
		} else if message.Data != nil {
			msgData := message.Data

			// DataRequest
			if msgData.Data != nil {
				log.Println("[RECEIVE] DataRequest from " + orig.String())
				go n.SendPiece(message, msgData.Website, msgData.Piece) //TODO: (gets data and sends it back)
			} else {
				//TODO ??? (I think it is not needed because you open temporary
				//connections to download pices in RetrievePiece()

			}
		}
	}
}

// Search search for keywords match among all the websites on the network
func (n *Node) Search(search string) []string {
	terms := strings.Split(search, " ")

	var websites []*structs.Website
	for _, term := range terms {
		websites = append(websites, n.WebsiteMap.SearchKeyword(term)[:]...)
	}

	var results []string
	for _, w := range websites {
		results = append(results, w.Name)
	}

	return results
}

// RetrieveWebsite retrieve the archive of a website in order to display it itself
func (n *Node) RetrieveWebsite(name string, ch chan int) {
	website := n.WebsiteMap.Get(name)

	pieces := website.Pieces
	numPieces := len(pieces) / 8
	chans := make([]chan []byte, numPieces)

	for i := 0; i <= numPieces; i++ {
		piece := pieces[:i*8]
		chans[i] = make(chan []byte, 1)
		go n.RetrievePiece(website, piece, chans[i])
	}

	archive, err := os.Create(utils.SeedDir + website.Name)
	defer archive.Close()
	if err != nil {
		log.Fatal(err)
	}

	// write all pieces in archive at correct pos once retrieven
	var mutex sync.Mutex
	ok := make(chan int, numPieces)
	for i, c := range chans {
		go func() {
			data := <-c
			mutex.Lock()
			archive.WriteAt(data[:], int64(i*utils.DefaultPieceLength))
			mutex.Unlock()
			ok <- 1
		}()
	}

	// wait for all pieces to be written in archive
	for okPiece := 0; okPiece < numPieces; {
		okPiece += <-ok
	}

	// archive is now complete we can unbundle it and seed it
	// TODO need to checksum this !

	website.AddSeeder(n.Addr)

	website.Unbundle()
}

// RetrievePiece retrieves a piece from a website archive and input it in a channel
func (n *Node) RetrievePiece(website *structs.Website, piece string, c chan []byte) {
	for _, seeder := range website.GetSeeders() {
		rAddr := net.UDPAddr(seeder)
		// Create a random local address for a new connection
		tempPeer := &structs.Peer{
			IP:   n.Addr.IP,
			Port: n.Addr.Port + 10100 + rand.Intn(20000),
			Zone: n.Addr.Zone,
		}
		lAddr := net.UDPAddr(*tempPeer)

		// try 2 deifferent ports
		conn, err := net.DialUDP("udp4", &lAddr, &rAddr)
		defer conn.Close()
		if err != nil {
			tempPeer.Port++
			lAddr := net.UDPAddr(*tempPeer)
			conn, err = net.DialUDP("udp4", &lAddr, &rAddr)
			if err != nil {
				log.Fatal(err)
			}
		}

		conn.SetReadDeadline(time.Now().Add(utils.HeartBeatTimeout))

		message := comm.NewDataRequest(tempPeer, &seeder, website.Name, piece)

		message.Send(conn, &seeder)

		// Maybe make a const for buffer size
		buf := make([]byte, 65507)
		_, err = conn.Read(buf)
		if err != nil {
			n.CheckPeer(&seeder)
		} else {
			reply := comm.DecodeMessage(buf)
			// do some validity checks here
			data := reply.Data.Data

			sum := sha256.Sum256(data)
			hash := hex.EncodeToString(sum[:])

			if hash != piece {
				log.Println("bad piece " + piece + " for " + website.Name)
			} else {
				c <- data
				return
			}
		}
	}
}

// SendPiece sends a data reply with the data for the requested piece
func (n *Node) SendPiece(request *comm.Message, name, pieceToSend string) {
	website := n.WebsiteMap.Get(name)

	archiveData, err := ioutil.ReadFile(utils.SeedDir + website.Name)
	if err != nil {
		log.Fatal(err)
	}

	pieces := website.Pieces
	numPieces := len(pieces) / 8

	var data []byte
	for i := 0; i <= numPieces; i++ {
		piece := pieces[:i*8]
		if piece == pieceToSend {
			offset := i * website.PieceLength
			data = archiveData[offset : offset+website.PieceLength]
			// need to check for checksum here
			reply := comm.NewDataReply(request, data)
			reply.Send(n.Conn, n.Addr)
			return
		}
	}
}

// AntiEntropy sends websitemap to all known peers at given time interval
func (n *Node) AntiEntropy(timeout time.Duration) {
	ticker := time.NewTicker(timeout)

	for range ticker.C {
		n.SendWebsiteMap()
	}
}
