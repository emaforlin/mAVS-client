package main

import (
	"context"
	"os"

	"github.com/emaforlin/mAVS/internal/app/node"
	"github.com/emaforlin/mAVS/internal/component/cli"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	config, err := cli.ParseFlags()
	if err != nil {
		log.Fatal().Err(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logfile, err := os.OpenFile("mavs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	multi := zerolog.MultiLevelWriter(os.Stdout, logfile)
	node.NodeLogger = zerolog.New(multi).With().Timestamp().Logger()
	node.NodeLogger.Debug().Msg("Starting logger")

	n := node.New(ctx)

	if err := n.Start(0, uint(config.MiningDifficulty)); err != nil {
		log.Fatal().Err(err)
	}
}
