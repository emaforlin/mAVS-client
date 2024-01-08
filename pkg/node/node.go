package node

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"

	"github.com/emaforlin/VoteNGo/pkg/blockchain"
	"github.com/emaforlin/VoteNGo/pkg/handlers"
	"github.com/emaforlin/mAVS/pkg/cli"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	ctx  context.Context
	host host.Host
	l    zerolog.Logger
	bc   *blockchain.BlockChain
}

func New(c context.Context) Node {
	return Node{
		ctx: c,
		l:   zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}
}

var args cli.Config

func (n *Node) Start(listenPort uint16, difficulty uint) error {
	args = cli.GetArgs()

	h, err := makeBasicHost()
	if err != nil {
		// error handle
		n.l.Fatal().Err(err)
		return err
	}

	n.host = h

	n.l.Info().Msg(fmt.Sprint("Host creted. We are: ", n.host.ID()))
	n.l.Info().Msg(fmt.Sprint(n.host.Addrs()))

	n.bc = blockchain.CreateBlockchain(difficulty)
	handlers.SetBlockchain(n.bc)
	n.host.SetStreamHandler(protocol.ID(args.ProtocolID), handlers.HandleStream)
	// ctx := context.Background()
	kademliaDHT, err := dht.New(n.ctx, n.host)
	if err != nil {
		panic(err)
	}
	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	n.l.Debug().Msg("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(n.ctx); err != nil {
		panic(err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, peerAddr := range args.BootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := n.host.Connect(n.ctx, *peerinfo); err != nil {
				n.l.Fatal().AnErr("error connecting to peer", err)
			} else {
				n.l.Info().Msg(fmt.Sprint("connection established with bootstrap node ", *peerinfo))
			}
		}()
	}
	wg.Wait()

	// We use a rendezvous point "meet me at Agriculture Park" to announce our location.
	n.l.Info().Msg("Announcing ourselves...")
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(n.ctx, routingDiscovery, args.RendezvousString)
	n.l.Debug().Msg("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	n.l.Debug().Msg("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(n.ctx, args.RendezvousString)
	if err != nil {
		panic(err)
	}

	for peer := range peerChan {
		if peer.ID == n.host.ID() {
			continue
		}
		n.l.Debug().Msg(fmt.Sprint("Found peer:", peer))

		n.l.Debug().Msg(fmt.Sprint("Connecting to:", peer))
		stream, err := n.host.NewStream(n.ctx, peer.ID, protocol.ID(args.ProtocolID))

		if err != nil {
			n.l.Warn().AnErr("Connection failed", err)
			continue
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			go handlers.WriteData(rw, n.bc)
			go handlers.ReadData(rw, n.bc)
		}
		n.l.Info().Msg(fmt.Sprint("Connected to ", peer))
	}

	// waits for signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	// shutdown the node when signal received
	fmt.Println("Received signal, shutting down...")
	if err := h.Network().ClosePeer(h.ID()); err != nil {
		panic(err)
	}
	return nil
}

func makeBasicHost() (host.Host, error) {
	r := rand.Reader
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrs([]multiaddr.Multiaddr(args.ListenAddresses)...),
		libp2p.Identity(priv),
	}
	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	return basicHost, nil
}
