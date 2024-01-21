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

	"github.com/emaforlin/mAVS/internal/app/blockchain"
	"github.com/emaforlin/mAVS/internal/component/cli"
	"github.com/emaforlin/mAVS/internal/component/handlers"
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

var NodeLogger zerolog.Logger = zerolog.New(os.Stdout).With().Timestamp().Str("component", "node").Logger()

type Node struct {
	ctx    context.Context
	host   host.Host
	logger zerolog.Logger
	bc     *blockchain.Blockchain
}

func New(c context.Context) Node {
	NodeLogger.Debug().Msg("Creating a new Node")
	return Node{
		ctx:    c,
		logger: NodeLogger,
	}
}

var args cli.Config

func (n *Node) Start(listenPort uint16, difficulty uint) error {
	n.logger.Info().Msg("Starting node...")

	args = cli.GetArgs() // retrieve args

	h, err := makeBasicHost()
	if err != nil {
		// error handle
		n.logger.Fatal().Err(err)
		return err
	}

	n.host = h

	n.logger.Info().Msg(fmt.Sprint("Host creted. We are: ", n.host.ID()))
	n.logger.Info().Msg(fmt.Sprint(n.host.Addrs()))

	n.bc = blockchain.CreateBlockchain(n.logger.With().Str("component", "blockchain").Logger(), difficulty)

	handlers.SetBlockchain(n.bc)
	n.host.SetStreamHandler(protocol.ID(args.ProtocolID), handlers.HandleStream)

	kademliaDHT, err := dht.New(n.ctx, n.host)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	n.logger.Debug().Msg("Bootstrapping the DHT")
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
			var p peer.AddrInfo
			if peerinfo != nil {
				p = *peerinfo
			}
			if err := n.host.Connect(n.ctx, p); err != nil {
				n.logger.Fatal().AnErr("error connecting to peer", err)
			} else {
				n.logger.Info().Msg(fmt.Sprint("connection established with bootstrap node ", p))
			}
		}()
	}
	wg.Wait()

	// We use a rendezvous point "meet me at Agriculture Park" to announce our location.
	n.logger.Info().Msg("Announcing ourselves...")
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(n.ctx, routingDiscovery, args.RendezvousString)
	n.logger.Debug().Msg("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	n.logger.Debug().Msg("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(n.ctx, args.RendezvousString)
	if err != nil {
		panic(err)
	}

	for peer := range peerChan {
		if peer.ID == n.host.ID() {
			continue
		}
		n.logger.Debug().Msg(fmt.Sprint("Found peer:", peer))

		n.logger.Debug().Msg(fmt.Sprint("Connecting to:", peer))
		stream, err := n.host.NewStream(n.ctx, peer.ID, protocol.ID(args.ProtocolID))

		if err != nil {
			n.logger.Warn().AnErr("Connection failed", err)
			continue
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			go handlers.WriteData(rw, n.bc)
			go handlers.ReadData(rw, n.bc)
		}
		n.logger.Info().Msg(fmt.Sprint("Connected to ", peer))
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
