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
var BC *blockchain.BlockChain

func HandleStream(s net.Stream) {
	log.Println("Got a new stream!")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go ReadData(rw, BC)
	go WriteData(rw, BC)
}

func ReadData(rw *bufio.ReadWriter, BC *blockchain.BlockChain) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if str == "" {
			return
		}
		if str != "\n" {
			extBc := blockchain.BlockChain{}
			if err := json.Unmarshal([]byte(str), &extBc); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			// if len(extBc.Chain) > len(BC.Chain) && extBc.IsValid() {
			if extBc.Lenght > BC.Lenght && extBc.IsValid() {
				*BC = extBc
				bytes, err := json.MarshalIndent(BC, "", "  ")
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

func WriteData(rw *bufio.ReadWriter, BC *blockchain.BlockChain) {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(&BC)
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
	user := "testing"
	for {
		lastState := *BC
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

		BC.AddBlock(map[string]interface{}{
			"user":   user,
			"amount": amount,
		})
		if !BC.IsValid() {
			mutex.Lock()
			BC = &lastState
			mutex.Unlock()
		}
		bytes, err := json.Marshal(&BC)
		if err != nil {
			log.Fatal(err)
		}

		spew.Dump(&BC)

		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}
}
