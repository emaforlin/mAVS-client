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

type Block struct {
	index     uint
	timeStamp time.Time
	data      map[string]interface{}
	hash      string
	prevHash  string
	pow       int
}

type BlockChain struct {
	genesisBlock Block
	chain        []Block
	difficulty   uint
	Lenght       uint
}

func CreateBlockchain(d uint) BlockChain {
	genBlock := Block{
		index:     0,
		hash:      "0",
		prevHash:  "0",
		timeStamp: time.Now(),
	}

	return BlockChain{
		genesisBlock: genBlock,
		chain:        []Block{genBlock},
		difficulty:   d,
	}
}

func (b *Block) CalculateHash() string {
	data, err := json.Marshal(b.data)
	if err != nil {
		log.Fatal(err)
	}
	ts, err := b.timeStamp.MarshalText()
	if err != nil {
		log.Fatal(err)
	}
	blockData := strconv.Itoa(int(b.index)) + b.prevHash + string(data) + string(ts) + strconv.Itoa(b.pow)

	blockHash := sha256.Sum256([]byte(blockData))

	return fmt.Sprintf("%x", blockHash)
}

func (b *Block) Mine(difficulty uint) {
	for !strings.HasPrefix(b.hash, strings.Repeat("0", int(difficulty))) {
		b.pow++
		b.hash = b.CalculateHash()
	}
}

func (b *BlockChain) AddBlock(data map[string]interface{}) {
	lastBlock := &b.chain[len(b.chain)-1]

	newBlock := &Block{
		index:     lastBlock.index + 1,
		data:      data,
		prevHash:  lastBlock.hash,
		timeStamp: time.Now(),
	}

	newBlock.Mine(b.difficulty)
	b.chain = append(b.chain, *newBlock)
	b.Lenght = uint(len(b.chain))
}

func (b *BlockChain) IsValid() bool {
	if b.Lenght != uint(len(b.chain)) ||
		b.genesisBlock.timeStamp != b.chain[0].timeStamp {
		return false
	}
	for i := range b.chain[1:] {
		prevBlock := &b.chain[i]
		currBlock := &b.chain[i+1]

		if currBlock.prevHash != prevBlock.hash ||
			currBlock.hash != currBlock.CalculateHash() {
			return false
		}
	}
	return true
}
