package main
import (
	"bytes"
	"encoding/gob"
	"log"
)
type Block struct {
	PrevBlockHash [32]byte
	Target uint8
	Nonce uint32
	Payload [1048576]byte //empty payload for simulating transmission
}

type BlockInfo struct{
	B *Block
	Hash [32]byte
	Height int
}


func (block *Block) SerializeBlock() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(block)
	if err!=nil{
		log.Panic(err)
	}
	return buffer.Bytes()
}

func DeserializeBlock(serializedBlock []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(serializedBlock))
	err := decoder.Decode(&block)
	if err!=nil{
		log.Panic(err)
	}
	return &block
}

func (blockInfo *BlockInfo) SerializeBlockInfo() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(blockInfo)
	if err!=nil{
		log.Panic(err)
	}
	return buffer.Bytes()
}


func DeserializeBlockInfo(serializedBlock []byte) *BlockInfo {
	var blockInfo BlockInfo
	decoder := gob.NewDecoder(bytes.NewReader(serializedBlock))
	err := decoder.Decode(&blockInfo)
	if err!=nil{
		log.Panic(err)
	}
	return &blockInfo
}
