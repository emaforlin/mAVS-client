package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/emaforlin/VoteNGo/pkg/blockchain"
	"github.com/emaforlin/VoteNGo/pkg/handlers"
	"github.com/emaforlin/VoteNGo/pkg/node"
	golog "github.com/ipfs/go-log"
	peer "github.com/libp2p/go-libp2p/core/peer"
	pstore "github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

func main() {
	flag.UintVar(&blockchain.Difficulty, "m", 5, "Mining difficulty")
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	flag.StringVar(&blockchain.User, "u", "default", "peer username")
	flag.Parse()

	blockchain.BLOCKCHAIN = blockchain.CreateBlockchain(blockchain.Difficulty)

	golog.SetAllLoggers(golog.LevelInfo)

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind with -l option")
	}

	ha, err := node.MakeBasicHost(*listenF)
	if err != nil {
		log.Fatal(err)
	}

	if *target == "" {
		log.Println("listening for connections")

		// Set a stream handler on host A. /p2p/1.0.0 is an example of
		// a user-defined protocol name.
		ha.SetStreamHandler("/p2p/1.0.0", handlers.HandleStream)

		select {}

		// Listener code ends
	} else {
		ha.SetStreamHandler("/p2p/1.0.0", handlers.HandleStream)

		// The following code extracts target's peer ID from the
		// given multiaddress
		ipfsaddr, err := ma.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}
		pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}

		//peerid, err := peer.IDB58Decode(pid)
		peerid, err := peer.Decode(pid)

		if err != nil {
			log.Fatalln(err)
		}

		// Decapsulate the /ipfs/<peerID> part from the target
		// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
		targetPeerAddr, _ := ma.NewMultiaddr(
			//fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
			fmt.Sprintf("/ipfs/%s", peer.Encode(peerid)))

		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		// We have a peer ID and a targetAddr so we add it to the peerstore
		// so LibP2P knows how to contact it
		ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

		log.Println("opening stream")
		// make a new stream from host B to host A
		// it should be handled on host A by the handler we set above because
		// we use the same /p2p/1.0.0 protocol
		s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

		// Create a thread to read and write data.
		go handlers.ReadData(rw)
		go handlers.WriteData(rw)

		select {} // hang forever

	}
}
