package main

type Network struct {
	InChans     map[int]chan *Message
	OutChans    map[int] map[int]chan *Message
}

func NewNetwork (nodes []int, graph map[int][]int) *Network {
	var network Network
	network.InChans = make(map[int] chan *Message)
	network.OutChans = make(map[int] map[int]chan *Message)
	for i := 0; i < len(nodes); i++ {
		dst := nodes[i]
		network.InChans[dst] = make(chan *Message, 50)
	}
	for i := 0; i < len(nodes); i++ {
		src := nodes[i]
		network.OutChans[src] = make(map[int]chan *Message)
		for j := 0; j < len(graph[src]); j++ {
			dst := graph[src][j]
			network.OutChans[src][dst] = network.InChans[dst]
		}
	}
	return &network
}
