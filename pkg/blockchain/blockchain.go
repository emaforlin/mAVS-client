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
var BLOCKCHAIN BlockChain

type Block struct {
	Index     uint
	TimeStamp time.Time
	data      map[string]interface{}
	Hash      string
	PrevHash  string
	pow       int
}

type BlockChain struct {
	genesisBlock Block
	Chain        []Block
	difficulty   uint
}

func (b Block) calculateHash() string {
	data, err := json.Marshal(b.data)
	if err != nil {
		log.Fatal(err)
	}

	blockData := strconv.Itoa(int(b.Index)) + b.PrevHash + string(data) + b.TimeStamp.String() + strconv.Itoa(b.pow)

	blockHash := sha256.Sum256([]byte(blockData))

	return fmt.Sprintf("%x", blockHash)
}

func (b Block) mine(difficulty uint) {
	for !strings.HasPrefix(b.Hash, strings.Repeat("0", int(difficulty))) {
		b.pow++
		b.Hash = b.calculateHash()
	}
}

func CreateBlockchain(d uint) *BlockChain {
	genBlock := Block{
		Index:     0,
		Hash:      "0",
		TimeStamp: time.Now(),
	}
	log.Printf("Starting a blockchain with difficulty %d\n", d)

	return &BlockChain{
		genesisBlock: genBlock,
		Chain:        []Block{genBlock},
		difficulty:   d,
	}
}

func (b *BlockChain) AddBlock(author string, amount int) {

	blockData := map[string]interface{}{
		"author": author,
		"amount": amount,
	}

	lastBlock := b.Chain[len(b.Chain)-1]
	var i uint
	if len(b.Chain) == 1 {
		i = 1
	} else {
		i = lastBlock.Index + 1
	}

	newBlock := Block{
		Index:     i,
		data:      blockData,
		PrevHash:  lastBlock.Hash,
		TimeStamp: time.Now(),
	}

	newBlock.mine(b.difficulty)
	b.Chain = append(b.Chain, newBlock)
}

func (b *BlockChain) IsValid() bool {
	for i := range b.Chain[1:] {
		prevBlock := b.Chain[i]
		currBlock := b.Chain[i+1]

		if currBlock.PrevHash != prevBlock.Hash || currBlock.Index != prevBlock.Index+1 || currBlock.Hash != currBlock.calculateHash() {
			return false
		}
	}
	return true
}
