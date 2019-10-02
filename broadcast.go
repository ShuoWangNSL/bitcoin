package main

import (
	//"time"
	//"sync"
	"fmt"
)

type Message struct {
	Type    string
	Payload []byte
}

type Broadcaster struct {
	Inchan chan *Message
	Outchans map[int]chan *Message
}

func (b *Broadcaster) Unicast (msg *Message, dest int)  {
	select {
	case b.Outchans[dest] <- msg:
	default:
		fmt.Errorf("Unicast failed")
	}
}

func (b *Broadcaster) Broadcast(msg *Message)  {
	for i := range b.Outchans {
		b.Unicast(msg, i)
	}
}

