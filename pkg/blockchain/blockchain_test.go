package blockchain_test

import (
	"math/rand"
	"testing"

	"github.com/emaforlin/VoteNGo/pkg/blockchain"
)

func FuzzAddBlock(f testing.F) {
	for j := 0; j < 5; j++ {
		testData := map[string]interface{}{
			"author": generateRandomString(8),
			"amount": rand.Float32(),
		}
		f.Add(testData)
	}

	f.Fuzz(func(t *testing.T, data map[string]interface{}) {
		difficulty := rand.Int() % 10
		testBC := blockchain.CreateBlockchain(uint(difficulty))
		if !testBC.IsValid() {
			t.Errorf("Error invalid blockchain: %+v", testBC)
		}
		testBC.AddBlock(data)
		if !testBC.IsValid() {
			t.Errorf("Error invalid blockchain: %+v", testBC)
		}
	})

}

func generateRandomString(lenght int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, lenght)
	for i := range b {
		b[i] = charset[rand.Int()%len(charset)+1]
	}
	return string(b)
}
