package cli

import (
	"flag"
	"strings"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	maddr "github.com/multiformats/go-multiaddr"
)

// A new type we need for writing a custom flag parser
type addrList []maddr.Multiaddr

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := maddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

func StringsToAddrs(addrStrings []string) (maddrs []maddr.Multiaddr, err error) {
	for _, addrString := range addrStrings {
		addr, err := maddr.NewMultiaddr(addrString)
		if err != nil {
			return maddrs, err
		}
		maddrs = append(maddrs, addr)
	}
	return
}

type Config struct {
	MiningDifficulty uint
	RendezvousString string
	BootstrapPeers   addrList
	ListenAddresses  addrList
	ProtocolID       string
}

var config Config

func ParseFlags() (Config, error) {
	flag.UintVar(&config.MiningDifficulty, "m", 5, "Sets the mining difficulty")
	flag.StringVar(&config.RendezvousString, "rendezvous", "meet me at Agriculture Park",
		"Unique string used to identify a group of nodes. Share this with other nodes so they can connect to yours")
	flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")
	flag.Var(&config.ListenAddresses, "listen", "Adds a multiaddress to the listen list")

	flag.Parse()

	config.ProtocolID = "/mavs/0.1.1"

	if len(config.BootstrapPeers) == 0 {
		config.BootstrapPeers = dht.DefaultBootstrapPeers
	}

	return config, nil
}

func GetArgs() Config {
	return config
}
