package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"

	"strconv"
	"strings"
	"time"
)

var Difficulty uint
var User string
var BLOCKCHAIN *BlockChain

type Block struct {
	Index     uint
	TimeStamp time.Time
	Data      map[string]interface{}
	Hash      string
	PrevHash  string
	Pow       int
}

type BlockChain struct {
	GenesisBlock Block
	Chain        []Block
	Difficulty   uint
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

func CreateBlockchain(d uint) *BlockChain {
	genBlock := Block{
		Index:     0,
		Hash:      "0",
		PrevHash:  "0",
		TimeStamp: time.Now(),
	}

	return &BlockChain{
		GenesisBlock: genBlock,
		Chain:        []Block{genBlock},
		Difficulty:   d,
	}
}

func (b *BlockChain) AddBlock(author string, amount int) {
	blockData := map[string]interface{}{
		"author": author,
		"amount": amount,
	}

	lastBlock := &b.Chain[len(b.Chain)-1]

	newBlock := &Block{
		Index:     lastBlock.Index + 1,
		Data:      blockData,
		PrevHash:  lastBlock.Hash,
		TimeStamp: time.Now(),
	}

	newBlock.Mine(b.Difficulty)
	b.Chain = append(b.Chain, *newBlock)
}

func (b *BlockChain) IsValid() bool {
	for i := range b.Chain[1:] {
		prevBlock := &b.Chain[i]
		currBlock := &b.Chain[i+1]

		if currBlock.PrevHash != prevBlock.Hash || currBlock.Hash != currBlock.CalculateHash() {
			return false
		}
	}
	return true
}
