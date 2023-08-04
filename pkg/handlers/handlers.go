package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/emaforlin/VoteNGo/pkg/blockchain"
	net "github.com/libp2p/go-libp2p/core/network"
)

var mutex = &sync.Mutex{}

func HandleStream(s net.Stream) {
	log.Println("Got a new stream!")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go ReadData(rw)
	go WriteData(rw)
}

func ReadData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if str == "" {
			return
		}
		if str != "\n" {
			var bc = blockchain.CreateBlockchain(blockchain.Difficulty)
			bc.Difficulty = blockchain.Difficulty
			if err := json.Unmarshal([]byte(str), &bc); err != nil {
				log.Fatal(err)
			}
			mutex.Lock()
			if len(bc.Chain) > len(blockchain.BLOCKCHAIN.Chain) && bc.IsValid() {
				*blockchain.BLOCKCHAIN = *bc
				bytes, err := json.MarshalIndent(blockchain.BLOCKCHAIN, "", "  ")
				if err != nil {
					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()

		}
	}
}

func WriteData(rw *bufio.ReadWriter) {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(&blockchain.BLOCKCHAIN)
			if err != nil {
				log.Fatal(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()
		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		lastState := &blockchain.BLOCKCHAIN
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		sendData = strings.Replace(sendData, "\n", "", -1)
		amount, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}

		blockchain.BLOCKCHAIN.AddBlock(blockchain.User, amount)
		if !blockchain.BLOCKCHAIN.IsValid() {
			mutex.Lock()
			blockchain.BLOCKCHAIN = *lastState
			mutex.Unlock()
		}
		bytes, err := json.Marshal(&blockchain.BLOCKCHAIN)
		if err != nil {
			log.Fatal(err)
		}

		spew.Dump(&blockchain.BLOCKCHAIN)

		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}
}
