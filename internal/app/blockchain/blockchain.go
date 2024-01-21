package blockchain

import (
	"fmt"

	"strings"
	"time"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

type BlockchainService interface {
	AddBlock()
	IsValid()
}

type Blockchain struct {
	GenesisBlock Block
	Chain        []Block
	Difficulty   uint
}

func CreateBlockchain(l zerolog.Logger, d uint) *Blockchain {
	firstHash := strings.Repeat("0", int(d))
	genBlock := Block{
		Index:     0,
		Hash:      firstHash,
		PrevHash:  "0",
		TimeStamp: time.Now(),
	}
	logger = l
	logger.Debug().Msg("Creating blockchain")

	return &Blockchain{
		GenesisBlock: genBlock,
		Chain:        []Block{genBlock},
		Difficulty:   d,
	}
}

func (b *Blockchain) AddBlock(newBlock Block) {
	lastBlock := &b.Chain[len(b.Chain)-1]

	newBlock.Index = lastBlock.Index + 1
	newBlock.PrevHash = lastBlock.Hash
	newBlock.TimeStamp = time.Now()

	newBlock.Mine(b.Difficulty)
	b.Chain = append(b.Chain, newBlock)

	logger.Info().Msg(fmt.Sprintf("Added block number %d: %s", newBlock.Index, newBlock.Hash))
}

func (b *Blockchain) IsValid() bool {
	for i := range b.Chain[1:] {
		prevBlock := &b.Chain[i]
		currBlock := &b.Chain[i+1]

		if currBlock.PrevHash != prevBlock.Hash || currBlock.Hash != currBlock.CalculateHash() {
			return false
		}
	}
	return true
}

func (b Blockchain) Lenght() int {
	return len(b.Chain)
}
