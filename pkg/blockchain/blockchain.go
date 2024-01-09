package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"

	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Block struct {
	Index     uint
	TimeStamp time.Time
	Data      map[string]interface{}
	Hash      string
	PrevHash  string
	Pow       int
}

type Blockchain struct {
	logger       zerolog.Logger
	GenesisBlock Block
	Chain        []Block
	Difficulty   uint
	Lenght       uint
}

func CreateBlockchain(l zerolog.Logger, d uint) *Blockchain {
	firstHash := strings.Repeat("0", int(d))
	genBlock := Block{
		Index:     0,
		Hash:      firstHash,
		PrevHash:  "0",
		TimeStamp: time.Now(),
	}

	l.Debug().Msg("Creating blockchain")

	return &Blockchain{
		logger:       l,
		GenesisBlock: genBlock,
		Chain:        []Block{genBlock},
		Difficulty:   d,
	}
}

func (b *Block) CalculateHash() string {
	data, err := json.Marshal(b.Data)
	if err != nil {
		log.Fatal(err)
	}
	ts, err := b.TimeStamp.MarshalText()
	if err != nil {
		log.Fatal(err)
	}
	blockData := strconv.Itoa(int(b.Index)) + b.PrevHash + string(data) + string(ts) + strconv.Itoa(b.Pow)

	blockHash := sha256.Sum256([]byte(blockData))

	return fmt.Sprintf("%x", blockHash)
}

func (b *Block) Mine(difficulty uint) {
	for !strings.HasPrefix(b.Hash, strings.Repeat("0", int(difficulty))) {
		b.Pow++
		b.Hash = b.CalculateHash()
	}
}

func (b *Blockchain) AddBlock(body map[string]interface{}) {
	lastBlock := &b.Chain[len(b.Chain)-1]

	newBlock := &Block{
		Index:     lastBlock.Index + 1,
		Data:      body,
		PrevHash:  lastBlock.Hash,
		TimeStamp: time.Now(),
	}

	newBlock.Mine(b.Difficulty)
	b.Chain = append(b.Chain, *newBlock)
	b.Lenght = uint(len(b.Chain))

	b.logger.Info().Msg(fmt.Sprintf("Block added %s", b.Chain[len(b.Chain)-1].Hash))
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
