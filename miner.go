package main
import (
	"bytes"
	"fmt"
	"crypto/sha256"
	"sync"
	"math"
	"time"
	"math/rand"
)

type Miner struct {
	mu        sync.Mutex
	persistPath string
	persistHeight int
	confidenceDepth int
}

func (miner *Miner) PoW (block *Block) error {
	var hash [32]byte
	tgt := make([]byte, 1)
	tgt[0] = block.Target
	random := rand.Uint32()
	maxi := uint32(math.MaxUint32)

	for i := random; i < maxi; i++ {
		block.Nonce = i
		hash = sha256.Sum256(block.Serialize())
		if bytes.Compare(hash[:], tgt) < 0 {
			return nil
		}
		time.Sleep(1000 * time.Millisecond)
	}

	for i := uint32(0); i < random; i++ {
		block.Nonce = i
		hash = sha256.Sum256(block.Serialize())
		if bytes.Compare(hash[:], tgt) < 0 {
			return nil
		}
		time.Sleep(1000 * time.Millisecond)
	}
	return fmt.Errorf("No solution")
}