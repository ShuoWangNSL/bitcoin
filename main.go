package main

import "time"

func main(){
	nodes := []int {0, 1, 2, 3, 4}
	graph := make(map[int][]int)
	graph[0] = []int{1,2,3,4}
	graph[1] = []int{0,2,3}
	graph[2] = []int{0,1,4}
	graph[3] = []int{0,1}
	graph[4] = []int{0,2}

	network := NewNetwork(nodes, graph)
	chains := make(map[int] *Blockchain)
	for i := 0; i < len(nodes); i++ {
		seq := nodes[i]
		go func(seq int) {
			chains[seq] = NewBlockchain(seq, network.InChans[seq], network.OutChans[seq])
		}(seq)
	}
}