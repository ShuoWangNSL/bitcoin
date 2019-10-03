package main

import (
	"crypto/sha256"
	"bytes"
	"sync"
	"fmt"

)

type Blockchain struct {
	id int
	curHeight int
	hashes [][32]byte
	blocks []*Block
	persister *Persister
	miner *Miner
	broacaster *Broadcaster
	mu sync.Mutex
	newblock chan struct{}
}

type PartialChain struct{
	Blocks []*Block
	Hashes [][32]byte
	Start int
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
		curHeight:	0,
	}
	chain.newblock = make(chan struct{}, 1)
	chain.persister = NewPersister(seq)
	genesis := &Block{
		PrevBlockHash: [32]byte{},
		Target :16,
		Nonce : 2,
	}
	chain.blocks = append(chain.blocks, genesis)
	chain.hashes = append(chain.hashes, sha256.Sum256(genesis.SerializeBlock()))
	return chain
}

func (chain *Blockchain) Mine() {
	for {
		var block Block
		block.PrevBlockHash = chain.hashes[chain.curHeight]
		block.Target = 32
		ok, hash := chain.miner.PoW(&block,chain.newblock)
		if !ok {
			continue
		}
		blockInfo := &BlockInfo{&block, hash, chain.curHeight + 1}
		fmt.Printf("blockchain_%d mined: block_%d\n",chain.id,chain.curHeight + 1)
		chain.addBlock(blockInfo, true)
	}

}

func (chain *Blockchain) Listen() {
	for {
		select {
		case msg := <- chain.broacaster.Inchan:
			chain.Receive(msg)
		default:
		}
	}
}

func (chain *Blockchain) addBlock(blockInfo *BlockInfo,  selfMined bool) {
	chain.mu.Lock()
	defer chain.mu.Unlock()
	if chain.curHeight + 1 != blockInfo.Height{
		return
	}
	chain.curHeight++
	chain.hashes = append(chain.hashes, blockInfo.Hash)
	chain.blocks = append(chain.blocks, blockInfo.B)
	if !selfMined {
		chain.newblock <- struct{}{}
	}

	chain.SendBlockInfo(blockInfo)
	//fmt.Printf("blockchain_%d add and send: block_%d\n",chain.id,chain.curHeight)
	chain.persister.Persist(chain.blocks)
}

func (chain *Blockchain) Validate(blockInfo *BlockInfo) int {
	chain.mu.Lock()
	defer chain.mu.Unlock()

	if blockInfo.Height <= chain.curHeight {
		return 0
	}
	if blockInfo.Height > chain.curHeight + 1 {
		return 2
	}
	prevhash := chain.hashes[chain.curHeight]
	if !bytes.Equal(prevhash[:], blockInfo.B.PrevBlockHash[:]) {
		return 0
	}
	hash := sha256.Sum256(blockInfo.B.SerializeBlock())
	if !bytes.Equal(hash[:], blockInfo.Hash[:]) {
		return 0
	}
	tgt := make([]byte, 1)
	tgt[0] = blockInfo.B.Target
	if bytes.Compare(hash[:], tgt) >= 0 {
		return 0
	}
	return 1
}

func (chain *Blockchain) SendBlockInfo(blockInfo *BlockInfo) {
	payload := blockInfo.SerializeBlockInfo()
	msg := &Message{
		Src		:	chain.id,
		Type	: "blockInfo",
		Payload	: payload,
	}
	chain.broacaster.Broadcast(msg)
}


func (chain *Blockchain) Receive(msg *Message) {
	switch msg.Type {
	case "blockInfo":
		blockInfo := DeserializeBlockInfo(msg.Payload)
		chain.ReceiveBlockInfo(blockInfo, msg.Src)
	case "hashesRequest":
		chain.ReplyHashesRequest(msg.Src)
	case "replyHashesRequest":
		var hashes [][32]byte
		Deserialize(msg.Payload, &hashes)
		ok, start := chain.CheckHashes(hashes)
		fmt.Println(start)
		if ok {
			chain.SendBlocksRequest(msg.Src, start)
		}
	case "blocksRequest":
		var start int
		Deserialize(msg.Payload, &start)
		chain.SendPartialChain(msg.Src, start)
	case "partialChain":
		var pc PartialChain
		Deserialize(msg.Payload, &pc)
		chain.ApplyPartialChain(&pc)
	}
}

func (chain *Blockchain) ReceiveBlockInfo(blockInfo *BlockInfo, src int) {
	switch chain.Validate(blockInfo) {
	case 1:
		fmt.Printf("blockchain_%d heard: block_%d\n",chain.id,blockInfo.Height)
		chain.addBlock(blockInfo, false)
	case 2:
		//chain.SendHashesRequest(src)
	}
}

func (chain *Blockchain) SendHashesRequest(src int) {
	fmt.Printf("blockchain_%d SendHashesRequest\n",chain.id)
	msg := &Message{
		Src		:	chain.id,
		Type	: "hashesRequest",
	}
	chain.broacaster.Unicast(msg, src)
}

func (chain *Blockchain) ReplyHashesRequest(dest int) {
	fmt.Printf("blockchain_%d ReplyHashesRequest\n",chain.id)
	msg := &Message{
		Src		:	chain.id,
		Type	: "replyHashesRequest",
		Payload	: Serialize(chain.hashes),
	}
	chain.broacaster.Unicast(msg, dest)
}

func (chain *Blockchain) CheckHashes(hashes [][32]byte) (bool, int) {
	l := len(hashes)
	H := chain.curHeight + 1
	if l <= H {
		return false, 0
	}
	var i int
	for i = 0; i < H; i++ {
		if !bytes.Equal(chain.hashes[i][:], hashes[i][:]) {
			break
		}
	}
	if i == 0 {
		return false, 0
	}
	return true, i
}

func (chain *Blockchain) SendBlocksRequest(src int, start int) {
	fmt.Printf("blockchain_%d SendBlocksRequest\n",chain.id)
	msg := &Message{
		Src		:	chain.id,
		Type	: "blocksRequest",
		Payload : Serialize(start),
	}
	chain.broacaster.Unicast(msg, src)
}

func (chain *Blockchain) SendPartialChain(src int, start int) {
	fmt.Printf("blockchain_%d SendPartialChain\n",chain.id)
	if start >= chain.curHeight {
		return
	}
	pc := &PartialChain{
		chain.blocks[start:],
		chain.hashes[start:],
		start,
	}
	msg := &Message{
		Src		:	chain.id,
		Type	: "partialChain",
		Payload : Serialize(pc),
	}
	chain.broacaster.Unicast(msg, src)
}

func (chain *Blockchain) ApplyPartialChain(pc *PartialChain) {
	fmt.Printf("blockchain_%d ApplyPartialChain\n",chain.id)
	if pc.Start + len(pc.Blocks) <= chain.curHeight + 1 || pc.Start == 0 {
		return
	}
	//prevhash := chain.hashes[pc.start - 1]
	var i int
	for i = 0; i < len(pc.Blocks) ;i++ {
		if false { // check validity
			return
		}
	}
	chain.mu.Lock()
	chain.curHeight = pc.Start + len(pc.Blocks) - 1
	chain.blocks = append(chain.blocks, pc.Blocks...)
	chain.hashes = append(chain.hashes, pc.Hashes...)
	fmt.Printf("chain_%d append partial chain from block_%d to block_%d\n", chain.id, pc.Start, chain.curHeight)
	chain.mu.Unlock()
	chain.newblock <- struct{}{}
	chain.persister.Persist(chain.blocks)
}


