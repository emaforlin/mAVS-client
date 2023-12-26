package main

import (
	"flag"
	"log"

	"github.com/emaforlin/VoteNGo/pkg/node"
	golog "github.com/ipfs/go-log"
)

func main() {
	//flag.UintVar(&blockchain.Difficulty, "m", 5, "Mining difficulty")
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	//flag.StringVar(&blockchain.User, "u", "default", "peer username")
	flag.Parse()
	golog.SetAllLoggers(golog.LevelInfo)

	host, err := node.MakeBasicHost(*listenF)
	if err != nil {
		log.Fatal(err)
	}
	node.StartNode(host, *target, *listenF)

}
