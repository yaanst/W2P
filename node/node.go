package node

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"log"
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
	HBCounter    *structs.Counter
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
	hbCounter := structs.NewCounter()

	connAddr := net.UDPAddr(*addr)
	conn, err := net.ListenUDP("udp4", &connAddr)
	utils.CheckError(err)

	return &Node{
		Name:         name,
		Addr:         addr,
		Conn:         conn,
		Peers:        peers,
		HBCounter:    hbCounter,
		RoutingTable: rt,
		WebsiteMap:   wm,
	}
}

// NewConnAndPeer creates a new connection on a free port in the range given
// as well as the peer for this connection
func NewConnAndPeer(ip net.IP, port, maxPort int) (*net.UDPConn, *structs.Peer) {
	// to avoid conflict when creating a new node on same machine
	minPort := port + 100

	tempPeer := &structs.Peer{
		IP:   ip,
		Port: minPort,
	}
	lAddr := net.UDPAddr(*tempPeer)

	conn, err := net.ListenUDP("udp4", &lAddr)

	for err != nil {
		tempPeer.Port++
		if tempPeer.Port > maxPort {
			tempPeer.Port = minPort
		}
		lAddr := net.UDPAddr(*tempPeer)
		conn, err = net.ListenUDP("udp4", &lAddr)
	}

	return conn, tempPeer
}

// -----------
// - Methods -
// -----------

// Init initialize a Node adding website already present on disk and checking
// wether we have their metadata, also checking every dir is present
func (n *Node) Init() {
	var dirPerm os.FileMode = 0755
	if _, err := os.Stat(utils.MetadataDir); err != nil {
		err := os.MkdirAll(utils.MetadataDir, dirPerm)
		utils.CheckError(err)
	}

	if _, err := os.Stat(utils.SeedDir); err != nil {
		err := os.MkdirAll(utils.SeedDir, dirPerm)
		utils.CheckError(err)
	}

	if _, err := os.Stat(utils.WebsiteDir); err != nil {
		err := os.MkdirAll(utils.WebsiteDir, dirPerm)
		utils.CheckError(err)
	}

	if _, err := os.Stat(utils.KeyDir); err != nil {
		err := os.MkdirAll(utils.KeyDir, dirPerm)
		utils.CheckError(err)
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
	log.Println("[WEBSITES]\tLoading website '" + name + "'")
	website := structs.LoadWebsite(name)

	n.WebsiteMap.Set(website)
	log.Println("[WEBSITES]\tSuccesfully loaded website '" + name + "' !")
}

// AddNewWebsite constructs a new website that has no metadatafile and adds
// it to the WebsiteMap
func (n *Node) AddNewWebsite(name string, keywords []string) {
	log.Println("[WEBSITES]\tAdding new website '" + name + "'")
	website := structs.NewWebsite(name, keywords)

	if website != nil && website.Owned() {
		log.Println("[WEBSITES]\t\tSigning website '" + name + "'")
		website.Sign()

		log.Println("[WEBSITES]\t\tBundling website '" + name + "'")
		website.Bundle()

		log.Println("[WEBSITES]\t\tGenerating pieces for website '" + name + "'")
		website.GenPieces(utils.DefaultPieceLength)
		website.Seeders.Add(n.Addr)

		log.Println("[WEBSITES]\t\tSaving Metadata for website '" + name + "'")
		website.SaveMetadata()

		n.WebsiteMap.Set(website)
		log.Println("[WEBSITES]\tSuccesfully added website '" + name + "' !")
	}
}

// UpdateWebsite update a Website in the WebsiteMap when user modified
// his website
func (n *Node) UpdateWebsite(name string, keywords []string) bool {
	log.Println("[WEBSITES]\tUpdating website '" + name + "'")
	website := n.WebsiteMap.Get(name)

	if website != nil && website.Owned() {
		log.Println("[WEBSITES]\t\tRe-signing website '" + name + "'")
		website.Sign()

		log.Println("[WEBSITES]\t\tOverwritting bundle of website '" + name + "'")
		website.Bundle()

		website.SetKeywords(keywords)

		log.Println("[WEBSITES]\t\tGenerating new pieces for website '" + name + "'")
		website.GenPieces(utils.DefaultPieceLength)
		website.IncVersion()

		log.Println("[WEBSITES]\t\tSaving new Metadata for website '" + name + "'")
		website.SaveMetadata()

		log.Println("[WEBSITES]\tSuccesfully updated website '" + name + "' !")

		return true
	}
	return false
}

// SendWebsiteMap shares the node's WebsiteMap with other nodes
func (n *Node) SendWebsiteMap() {
	for _, p := range n.Peers.GetAll() {
		message := comm.NewMeta(n.Addr, &p, n.WebsiteMap)
		via := n.RoutingTable.Get(p.String())
		message.Send(n.Conn, via)
		log.Println("[SENT]\tWebsiteMap to", p.String())
		n.CheckPeer(&p)
	}
}

// HeartBeat sends a hearbeat message to peer and waits for an answer or timeout
func (n *Node) HeartBeat(peer *structs.Peer, reachable chan bool) {
	// Create a random local address for a new connection
	conn, tempPeer := NewConnAndPeer(n.Addr.IP, n.Addr.Port, n.Addr.Port+10000)
	defer conn.Close()

	// Set Read timeout
	conn.SetReadDeadline(time.Now().Add(utils.HeartBeatTimeout))

	message := comm.NewHeartbeat(tempPeer, peer)
	buffer := make([]byte, utils.HeartBeatBufferSize)

	via := n.RoutingTable.Get(peer.String())
	message.Send(conn, via)

	log.Println("[SENT]\tHeartbeat to", peer.String())

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
	for n.HBCounter.Read() >= utils.HeartBeatLimit {
		time.Sleep(50 * time.Millisecond)
	}
	n.HBCounter.Inc()
	defer n.HBCounter.Dec()

	c := make(chan bool)
	go n.HeartBeat(peer, c)

	reachable := <-c
	if !reachable {
		log.Println("[HEARTBEAT]\tPeer", peer, "is down")
		n.Peers.Remove(peer)
		n.WebsiteMap.RemovePeer(peer)
	} else {
		log.Println("[HEARTBEAT]\tPeer", peer, "is up")
		n.Peers.Add(peer)
		n.RoutingTable.Set(peer.String(), peer) // Reset RoutingTable entry
	}
}

// DiscoverPeers checks for any unknown peer in the WM in order to add them
func (n *Node) DiscoverPeers(w *structs.Website) {
	for _, s := range w.GetSeeders() {
		if !n.Peers.Contains(&s) {
			n.Peers.Add(&s)
		}
	}
}

// MergeWebsiteMap merges a WebsiteMap into the local one
func (n *Node) MergeWebsiteMap(remoteWM *structs.WebsiteMap) {
	localWM := n.WebsiteMap

	rIndices := remoteWM.GetIndices()
	for _, rKey := range rIndices {
		lWeb := localWM.Get(rKey)
		rWeb := remoteWM.Get(rKey)
		go n.DiscoverPeers(rWeb)

		if lWeb == nil {
			log.Print("[WEBSITEMAP]\tAdding website '" + rWeb.Name + "'")
			localWM.Set(rWeb)
			n.RetrieveWebsite(rWeb.Name)
		} else if lWeb.PubKey.String() != rWeb.PubKey.String() {
			log.Fatalf("[WEBSITEMAP]\tPublic keys not matching for local/remote website %v\n", lWeb.Name)
		} else if rWeb.Version > lWeb.Version {
			log.Print("[WEBSITEMAP]\tUpdating website '" + lWeb.Name + "'")
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
			n.RetrieveWebsite(rWeb.Name)
		}
	}
}

// Listen listens for messages from other peers and acts on them
func (n *Node) Listen() {
	buffer := make([]byte, utils.ListenBufferSize)

	log.Println("[LISTENING]\ton", n.Addr.String())

	for {
		_, senderAddr, err := n.Conn.ReadFromUDP(buffer)
		sender := structs.ParsePeer(senderAddr.String())
		utils.CheckError(err)

		message := comm.DecodeMessage(buffer)
		orig := message.Orig
		dest := message.Dest

		// Forward message
		if !structs.PeerEquals(dest, n.Addr) {
			via := n.RoutingTable.Get(dest.String())
			message.Send(n.Conn, via)
		}

		// Update RoutingTable
		if !structs.PeerEquals(orig, sender) {
			n.RoutingTable.Set(orig.String(), sender)
		} else {
			n.RoutingTable.Set(orig.String(), orig)
		}

		// HeartBeat
		if message.Meta == nil && message.Data == nil {
			log.Println("[RECEIVE]\tHeartbeat from " + orig.String())
			heartbeat := comm.NewHeartbeat(n.Addr, orig)
			via := n.RoutingTable.Get(orig.String())
			heartbeat.Send(n.Conn, via)

			// WebsiteMapUpdate
		} else if message.Meta != nil {
			log.Println("[RECEIVE]\tWebsiteMap from " + orig.String())
			go n.CheckPeer(orig)
			go n.MergeWebsiteMap(message.Meta.WebsiteMap)

			// Data
		} else if message.Data != nil {
			msgData := message.Data

			// DataRequest
			if msgData.Data == nil {
				log.Println("[RECEIVE]\tDataRequest: '" + msgData.Piece + "' for '" +
					msgData.Website + "' from " + orig.String())
				go n.SendPiece(message, msgData.Website, msgData.Piece)
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
func (n *Node) RetrieveWebsite(name string) {
	log.Println("[PIECES]\tRetrieving pieces for website '" + name + "'")
	website := n.WebsiteMap.Get(name)

	pieces := website.Pieces
	numPieces := len(pieces) / utils.HashSize
	chans := make([]chan []byte, numPieces)

	for i := 0; i < numPieces; i++ {
		piece := pieces[i*utils.HashSize : (i+1)*utils.HashSize]
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

	log.Println("[PIECES]\tSuccessful retrieval of website '" + name + "'")

	// archive is now complete we can unbundle it and seed it
	// TODO need to checksum this !

	website.AddSeeder(n.Addr)

	log.Println("[WEBSITES]\tUnbundling website '" + name + "'")
	website.Unbundle()

	log.Println("[WEBSITES]\tSaving metadata for '" + name + "'")
	website.SaveMetadata()
}

// RetrievePiece retrieves a piece from a website archive and input it in a channel
func (n *Node) RetrievePiece(website *structs.Website, piece string, c chan []byte) {
	for _, seeder := range website.GetSeeders() {
		log.Println("[PIECES]\tRetrieving piece '" + piece + "' for website '" +
			website.Name + "' by " + seeder.String())

		conn, tempPeer := NewConnAndPeer(n.Addr.IP, n.Addr.Port, n.Addr.Port+10000)
		defer conn.Close()

		conn.SetReadDeadline(time.Now().Add(utils.HeartBeatTimeout))

		message := comm.NewDataRequest(tempPeer, &seeder, website.Name, piece)

		via := n.RoutingTable.Get(seeder.String())
		message.Send(conn, via)

		log.Println("[SENT]\t\tDatarequest: '" + piece + "' for website '" +
			website.Name + "' to " + seeder.String())

		// Maybe make a const for buffer size
		buf := make([]byte, 65507)
		_, err := conn.Read(buf)
		if err != nil {
			n.CheckPeer(&seeder)
		} else {
			reply := comm.DecodeMessage(buf)
			// do some validity checks here
			data := reply.Data.Data

			sum := sha256.Sum256(data)
			hash := hex.EncodeToString(sum[:])

			if hash != piece {
				log.Println("[PIECES]\tBad piece '" + piece + "' for website '" +
					website.Name + "' by " + seeder.String())
			} else {
				log.Println("[PIECES]\tGood piece '" + piece + "' for website '" +
					website.Name + "' by " + seeder.String())
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
	utils.CheckError(err)

	archiveSize := len(archiveData)
	pieces := website.Pieces
	numPieces := len(pieces) / utils.HashSize

	var data []byte
	for i := 0; i < numPieces; i++ {
		piece := pieces[i*utils.HashSize : (i+1)*utils.HashSize]
		if piece == pieceToSend {
			offsetStart := i * website.PieceLength
			offsetEnd := (i + 1) * website.PieceLength
			if offsetEnd >= archiveSize {
				offsetEnd = archiveSize
			}

			data = archiveData[offsetStart:offsetEnd]
			// need to check for checksum here
			reply := comm.NewDataReply(request, data)
			reply.Send(n.Conn, reply.Dest)
			log.Println("[SENT]\tPiece '" + piece + "' for website '" +
				website.Name + "' to " + reply.Dest.String())
			return
		}
	}
}

// AntiEntropy sends websitemap to all known peers at given time interval
func (n *Node) AntiEntropy(timeout time.Duration) {
	ticker := time.NewTicker(timeout)

	for range ticker.C {
		if n.Peers.Count() > 0 {
			n.SendWebsiteMap()
		}
	}
}
