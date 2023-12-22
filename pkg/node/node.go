package node

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"log"

	"github.com/emaforlin/VoteNGo/pkg/blockchain"
	"github.com/emaforlin/VoteNGo/pkg/handlers"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/multiformats/go-multiaddr"
)

func MakeBasicHost(listenPort int) (host.Host, error) {
	r := rand.Reader
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}
	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)

	log.Printf("I am %s\n", fullAddr)

	return basicHost, nil
}

func StartNode(h host.Host, dest string, l int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blockchain := blockchain.CreateBlockchain(3)
	handlers.BC = &blockchain

	if dest == "" {
		startPeer(ctx, h)
		// Listener code ends
	} else {
		rw, err := startPeerAndConnect(ctx, h, dest)
		if err != nil {
			log.Fatal(err)
		}

		// Create a thread to read and write data.
		go handlers.ReadData(rw, &blockchain)
		go handlers.WriteData(rw, &blockchain)

	}
	select {}
}

func startPeer(ctx context.Context, h host.Host) {
	// Set a stream handler on host A. /p2p/1.0.0 is an example of
	// a user-defined protocol name.
	h.SetStreamHandler("/p2p", handlers.HandleStream)

	var port string

	for _, la := range h.Network().ListenAddresses() {
		if p, err := la.ValueForProtocol(multiaddr.P_TCP); err == nil {
			port = p
			break
		}
	}

	if port == "" {
		log.Println("Was not able to find actual local port.")
		return
	}
	log.Printf("Run '... -d /ip4/127.0.0.1/tcp/%v/p2p/%s' on another console.\n", port, h.ID().Pretty())
	log.Println("You can replace 127.0.0.1 with public IP as well.")
	log.Println("Waiting for incoming connection")
	log.Println()
}

func startPeerAndConnect(ctx context.Context, h host.Host, dest string) (*bufio.ReadWriter, error) {
	log.Println("This node multiaddresses: ")
	for _, la := range h.Addrs() {
		log.Printf(" - %v\n", la)
	}
	log.Println()

	// Turn the destination into multiaddr.

	maddr, err := multiaddr.NewMultiaddr(dest)
	if err != nil {
		return nil, err
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, err
	}

	log.Printf("Multiaddrs: %q.", info.Addrs)
	h.Peerstore().AddAddr(info.ID, info.Addrs[0], peerstore.PermanentAddrTTL)

	// Start the stream with the destination
	// Multiaddress of the destination peer is fetched from the peerstore.
	s, err := h.NewStream(context.Background(), info.ID, "/p2p")
	if err != nil {
		return nil, err
	}
	log.Println("Established connection to destination")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	return rw, nil
}
