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

// type BlockAccessor interface {
// 	Hash()
// 	TimeStamp()
// 	Data()
// }

type Block struct {
	Index     uint
	TimeStamp time.Time
	Data      any
	Hash      string
	PrevHash  string
	Pow       int
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

func CreateBlock(data any) Block {
	return Block{
		Index:     0,
		TimeStamp: time.Time{},
		Data:      data,
		Hash:      "",
		PrevHash:  "",
		Pow:       0,
	}
}
