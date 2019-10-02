package main

import (
	"crypto/sha256"
	//"fmt"
	"bytes"
	"sync"
)

type Blockchain struct {
	id int
	blocks []*Block // Todo: discard persisted blocks
	persister *Persister
	miner *Miner
	broacaster *Broadcaster
	mu sync.Mutex
	newblock chan *Block
}

func NewBlockchain(seq int, inchan chan *Message, outchans map[int]chan *Message) *Blockchain {
	bc := &Broadcaster{
		Inchan: inchan,
		Outchans: outchans,
	}
	chain := &Blockchain {
		id : 		seq,
		blocks : 	[]*Block{},
		broacaster: bc,
	}
	chain.newblock = make(chan *Block)
	chain.persister = NewPersister(seq)
	genesis := &Block{
		PrevBlockHash: [32]byte{},
		Target :32,
		Nonce : 0,
	}
	chain.blocks = append(chain.blocks, genesis)
	return chain
}

func (chain *Blockchain) Mine() {
	chain.mu.Lock()
	defer chain.mu.Unlock()
	var block Block
	len := len(chain.blocks)
	block.PrevBlockHash = sha256.Sum256(chain.blocks[len - 1].Serialize())
	block.Target = 32
	chain.miner.PoW(&block)
}

func (chain *Blockchain) addBlock(block *Block) {
	chain.mu.Lock()
	defer chain.mu.Unlock()
	chain.blocks = append(chain.blocks, block)
	chain.persister.Persist(chain.blocks)
}

func (chain *Blockchain) Validate(block *Block) bool {
	chain.mu.Lock()
	defer chain.mu.Unlock()
	len := len(chain.blocks)
	prevhash := sha256.Sum256(chain.blocks[len - 1].Serialize())
	if !bytes.Equal(prevhash[:], block.PrevBlockHash[:]) {
		return false
	}
	hash := sha256.Sum256(block.Serialize())
	tgt := make([]byte, 1)
	tgt[0] = block.Target
	if bytes.Compare(hash[:], tgt) >= 0 {
		return false
	}
	return true
}

func (chain *Blockchain) SendBlock(block *Block) {
	msg := &Message{
		Type	: "block",
		Payload	: block.Serialize(),
	}
	chain.broacaster.Broadcast(msg)
}


func (chain *Blockchain) Receive(msg *Message) {
	switch msg.Type {
	case "block":
		block := Deserialize(msg.Payload)
		chain.ReceiveBlock(block)
	}
}

func (chain *Blockchain) ReceiveBlock(block *Block) {
	if !chain.Validate(block) {
		select {
		case chain.newblock <- block:
		default:
		}
	}
}


