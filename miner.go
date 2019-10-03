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

func (miner *Miner) PoW (block *Block, newblock chan struct{}) (bool, [32]byte) {
	var hash [32]byte
	tgt := make([]byte, 1)
	tgt[0] = block.Target
	random := rand.Uint32()
	maxi := uint32(math.MaxUint32)

	for i := random; i < maxi; i++ {
		select {
		case <-newblock:
			return false, [32]byte{}
		default:
		}
		block.Nonce = i
		hash = sha256.Sum256(block.SerializeBlock())
		if bytes.Compare(hash[:], tgt) < 0 {
			return true, hash
		}
		time.Sleep(5000 * time.Millisecond)
	}

	for i := uint32(0); i < random; i++ {
		select {
		case <-newblock:
			return false, [32]byte{}
		default:
		}
		block.Nonce = i
		hash = sha256.Sum256(block.SerializeBlock())
		if bytes.Compare(hash[:], tgt) < 0 {
			return true, hash
		}
		time.Sleep(5000 * time.Millisecond)
	}
	fmt.Errorf("No solution")
	return false, [32]byte{}
}