package CBR

import (
	"container/heap"
	// "fmt"
)

type City struct {
	Id       int
	CameFrom int
	Transp   int // transp of previous rel
}

type GraphGen struct {
	Graph

	Cities       map[int]City
	DepId, ArrId int

	Closed   map[int]map[int]map[int]map[int]bool   // dijkstra's closed vertices. [hin||ht][u][v][transp]
	Explored map[int]map[int]map[int]map[int]bool   // dijkstra's explored vertices (in the search list but no yet closed). [hin||ht][u][v][transp]
	Parents  map[int]map[int]map[int]map[int][]Node // links to parents of nodes [hin||ht][u][v][transp]. Note that a node might have various parents.

	H *UsualCombinationsHandler
}

func (g GraphGen) GetPath() []int {
	v := g.Cities[g.ArrId]
	var path = []int{g.ArrId}
	for v.Id != g.DepId {
		v = g.Cities[v.CameFrom]
		path = append(path, v.Id)
	}
	var res []int
	for i := len(path) - 1; i >= 0; i-- {
		res = append(res, path[i])
	}
	return res
}

func (g *GraphGen) createHt(i int) {
	// newMap := make(map[int]map[int]int)
	var newHt Ht
	heap.Init(&newHt)

	// get previous ht
	if i != g.DepId {
		prevHt, exists := (*g).Hts[(*g).Cities[i].CameFrom]
		if !exists {
			g.createHt((*g).Cities[i].CameFrom)
			prevHt = (*g).Hts[(*g).Cities[i].CameFrom]
		}
		for _, node := range prevHt {
			newNode := node
			newNode.ht = i
			heap.Push(&newHt, newNode)
		}
	}
	rootHinInt, found := (*g).Hins[i].Top()
	if found {
		rootHin := rootHinInt.(Node)
		rootHin.ht = i
		heap.Push(&newHt, rootHin)
	}
	(*g).Hts[i] = newHt
}

func (g *GraphGen) succCrossEdges(n Node) []Node {

	var res []Node

	if _, exists := (*g).Hts[n.u]; !exists {
		g.createHt(n.u)
	}

	// "pointer" to the root node
	rootHtInt, found := (*g).Hts[n.u].Top()
	if found {
		rootHt := rootHtInt.(Node)
		res = append(res, rootHt)
	}

	return res

}

func getNodeHin(u, v, transp int, nodes []Node) (n Node, found bool) {
	for _, n := range nodes {
		if n.u == u && n.v == v && n.transp == transp {
			return n, true
		}
	}
	return Node{}, false
}

func getNodeHt(u, v, transp int, nodes []Node) (n Node, found bool) {
	for _, n := range nodes {
		if n.u == u && n.v == v && n.transp == transp {
			return n, true
		}
	}
	return Node{}, false
}
