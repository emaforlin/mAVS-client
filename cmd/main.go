package main

import (
	"log"

	"github.com/emaforlin/VoteNGo/pkg/cli"
	"github.com/emaforlin/VoteNGo/pkg/node"

	golog "github.com/ipfs/go-log"
)

func main() {
	l := log.Default()
	config, err := cli.ParseFlags()
	if err != nil {
		l.Fatal(err)
	}

	golog.SetAllLoggers(golog.LevelInfo)

	n := node.New(l)
	if err := n.Start(uint16(config.ListenPort), config.Difficulty, config.TargetAddr); err != nil {
		l.Fatal(err)
	}

}
