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


func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(block)
	if err!=nil{
		log.Panic(err)
	}
	return buffer.Bytes()
}

func Deserialize(serializedBlock []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(serializedBlock))
	err := decoder.Decode(&block)
	if err!=nil{
		log.Panic(err)
	}
	return &block
}


