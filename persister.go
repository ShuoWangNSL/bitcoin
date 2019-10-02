package main

import (
	"sync"
	//"fmt"
	"os"
	"encoding/gob"
	"strconv"
	"path/filepath"
)

type Persister struct {
	mu        		sync.Mutex
	persistPath 	string
	persistHeight 	int
	confidenceDepth int
}

func NewPersister(seq int) *Persister{
	var persister Persister
	persister.confidenceDepth = 3
	curPath , _ := os.Getwd()
	//fmt.Println(curPath)
	persister.persistPath = filepath.Join(curPath, "blockchain_"+strconv.Itoa(seq))
	os.Mkdir(persister.persistPath, os.ModePerm)
	persister.persistHeight = 0
	return &persister
}

func (ps *Persister) PersistBlock(block *Block) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	blockFilename := filepath.Join(ps.persistPath, "block_"+strconv.Itoa(ps.persistHeight))
	//fmt.Println(blockFilename)
	ps.persistHeight++
	file, err := os.Create(blockFilename)
	defer file.Close()
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(block)
	}
}

func (ps *Persister) Persist(blocks []*Block) {
	curHeight := len(blocks)
	confidenceHeight := ps.persistHeight + ps.confidenceDepth
	//fmt.Printf("curHeight: %d\n" ,curHeight)
	if curHeight > confidenceHeight {
		for i := ps.persistHeight; i < confidenceHeight; i++ {
			//fmt.Printf("persist: %d\n" ,i)
			ps.PersistBlock(blocks[i])
		}
	}
}

func (ps *Persister) Load(h int) *Block {
	var block Block
	blockFilename := filepath.Join(ps.persistPath, "block_"+strconv.Itoa(h))
	file, err := os.Open(blockFilename)
	defer file.Close()
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(&block)
	}
	return &block
}

func (ps *Persister) Revert(cutoffHeight int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.persistHeight <= cutoffHeight {
		return
	}
	var blockFilename string
	for ; ps.persistHeight > cutoffHeight; ps.persistHeight-- {
		blockFilename = filepath.Join(ps.persistPath, "block_"+strconv.Itoa(ps.persistHeight - 1))
		os.Remove(blockFilename)
	}
}