package CBR

import (
	"container/heap"
	// "fmt"
	"github.com/jcasado94/tfg/common"
	"github.com/jmcvetta/neoism"
	"sort"
	"time"
)

const START_ID = -1
const END_ID = -2
const R = -3
const K = 15

type Graph struct {
	Hins map[int]Hin
	Hts  map[int]Ht
}

type Rel struct {
	Id               int
	DepTime, ArrTime time.Time
	Weight           float64
	Transp           int
	ArrNode          int
	DepNode          int
	CameFrom         int
}

type GraphSpec struct {
	Graph

	Rels map[int]Rel
	/*
		notice that a node can't be in a hin and a ht, except for the hin roots, but those we just treat them like ht's.
	*/
	Closed   map[int]map[int]map[int]bool    // dijkstra's closed vertices. [hin||ht][u][v]
	Dists    map[int]map[int]map[int]float64 // dijkstra's vertex dinstances. [hin||ht][u][v]
	Explored map[int]map[int]map[int]bool    // dijkstra's explored vertices (in the search list but no yet closed). [hin||ht][u][v]
	Parents  map[int]map[int]map[int][]Node  // links to parents of nodes [hin||ht][u][v]. Note that a node might have various parents.

	H *SameDayCombinationsHandler
}

// given the database and the relationships returns the Rel object initialized
func (g GraphSpec) GetRel(rel *neoism.Relationship) Rel {

	location := g.H.Location
	Db := g.H.Db

	rel.Db = Db
	relId := rel.Id()
	propsRel, _ := rel.Properties()
	start, _ := rel.Start()
	startId := start.Id()
	end, _ := rel.End()
	endId := end.Id()
	props, _ := rel.Properties()
	weight := props["price"].(float64)
	transp := int(props["transp"].(float64))

	depYear := int(propsRel["depYear"].(float64))
	depMonth := int(propsRel["depMonth"].(float64))
	depDay := int(propsRel["depDay"].(float64))
	depHour := int(propsRel["depHour"].(float64))
	depMin := int(propsRel["depMin"].(float64))
	arrYear := int(propsRel["arrYear"].(float64))
	arrMonth := int(propsRel["arrMonth"].(float64))
	arrDay := int(propsRel["arrDay"].(float64))
	arrHour := int(propsRel["arrHour"].(float64))
	arrMin := int(propsRel["arrMin"].(float64))
	timeDep := time.Date(depYear, time.Month(depMonth), depDay, depHour, depMin, 0, 0, location)
	timeArr := time.Date(arrYear, time.Month(arrMonth), arrDay, arrHour, arrMin+common.TRANSFER_TIME, 0, 0, location)

	newRel := Rel{Id: relId, Transp: transp, Weight: weight, DepTime: timeDep, ArrTime: timeArr, ArrNode: endId, DepNode: startId}

	return newRel

}

func (g GraphSpec) GetPath() []int {
	v := g.Rels[END_ID]
	var path = []int{END_ID}
	for v.Id != START_ID {
		v = g.Rels[v.CameFrom]
		path = append(path, v.Id)
	}
	var res []int
	for i := len(path) - 1; i >= 0; i-- {
		res = append(res, path[i])
	}
	return res
}

type Node struct {
	u             int
	v             int
	transp        int // only for GEN
	dValue        float64
	hin, hinIndex int
	ht, htIndex   int
	dist          float64
	// parent information
	indParent int
}

func getDelta(n *Node, m *Node) float64 {
	if isCross(n, m) {
		return m.dValue
	} else {
		return m.dValue - n.dValue
	}
}

func isCross(n *Node, m *Node) bool {
	return m.isTop()
}

func (n Node) isTop() bool {
	return n.ht != 0 && n.htIndex == 0
}

func (n Node) isHin() bool {
	return n.ht == 0
}

func (g *GraphSpec) createHt(i int) {
	// newMap := make(map[int]map[int]int)
	var newHt Ht
	heap.Init(&newHt)

	// get previous ht
	if i != START_ID {
		prevHt, exists := (*g).Hts[(*g).Rels[i].CameFrom]
		if !exists {
			g.createHt((*g).Rels[i].CameFrom)
			prevHt = (*g).Hts[(*g).Rels[i].CameFrom]
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

func (g *GraphSpec) succCrossEdges(n Node) []Node {

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

func (g *Graph) succHeapEdges(n Node) []Node {

	if n.v == R {
		return []Node{}
	}

	var res []Node

	hin := (*g).Hins[n.hin]
	hinNode := hin[n.hinIndex]
	left := g.getLeftHin(hinNode)
	right := g.getRightHin(hinNode)
	if left != 0 {
		res = append(res, hin[left])
	}
	if right != 0 {
		res = append(res, hin[right])
	}

	if n.isHin() {
		return res
	}

	// get Ht nodes
	ht := (*g).Hts[n.ht]
	htNode := ht[n.htIndex]
	left = g.getLeftHt(htNode)
	right = g.getRightHt(htNode)
	if left != 0 {
		res = append(res, ht[left])
	}
	if right != 0 {
		res = append(res, ht[right])
	}

	return res
}

// returns the index of the right child in the hin queue, if it exists. Otherwise, returns 0.
func (g Graph) getRightHin(n Node) int {
	sort.Sort(g.Hins[n.hin]) // might not be sorted, therefore nodes not in order
	hin := g.Hins[n.hin]
	var i int
	if n.hinIndex == 0 {
		i = 1
	} else {
		i = 2*n.hinIndex + 1
	}
	if len(hin) > i {
		return i
	} else {
		return 0
	}
}

// returns the index of the left child in the hin queue, if it exists. Otherwise, returns 0.
func (g Graph) getLeftHin(n Node) int {
	sort.Sort(g.Hins[n.hin]) // might not be sorted, therefore nodes not in order
	hin := g.Hins[n.hin]
	var i int
	if n.hinIndex == 0 {
		return 0
	} else {
		i = 2 * n.hinIndex
	}
	if len(hin) > i {
		return i
	} else {
		return 0
	}
}

// returns the index of the right child in the ht queue, if it exists. Otherwise, returns 0.
func (g Graph) getRightHt(n Node) int {
	sort.Sort(g.Hts[n.ht]) // might not be sorted, therefore nodes not in order
	ht := g.Hts[n.ht]
	i := 2*n.htIndex + 1
	if len(ht) > i {
		return i
	} else {
		return 0
	}
}

// returns the index of the left child in the ht queue, if it exists. Otherwise, returns 0.
func (g Graph) getLeftHt(n Node) int {
	sort.Sort(g.Hts[n.ht]) // might not be sorted, therefore nodes not in order
	ht := g.Hts[n.ht]
	i := 2 * n.htIndex
	if len(ht) > i {
		return i
	} else {
		return 0
	}
}

func (n Node) getHtOrHin() int {
	if n.isHin() {
		return n.hin
	} else {
		return n.ht
	}
}

type Hin []Node

func (pq Hin) Len() int { return len(pq) }

func (pq Hin) Empty() bool { return len(pq) == 0 }

func (pq Hin) Less(i, j int) bool {
	return pq[i].dValue < pq[j].dValue
}

func (pq Hin) Swap(i, j int) {
	iNew := pq[j]
	iNew.hinIndex = i
	jNew := pq[i]
	jNew.hinIndex = j
	pq[i], pq[j] = iNew, jNew
}

func (pq *Hin) Push(x interface{}) {
	node := x.(Node)
	node.hinIndex = len(*pq)
	*pq = append(*pq, node)
}

func (pq *Hin) Pop() interface{} {
	old := *pq
	n := len(old)
	rel := old[n-1]
	*pq = old[0 : n-1]
	return rel
}

func (pq Hin) Top() (node interface{}, found bool) {
	if len(pq) == 0 {
		return Node{}, false
	}
	return pq[0], true
}

func (pq Hin) getNode(u, v int) (n Node, found bool) {
	for _, n := range pq {
		if n.u == u && n.v == v {
			return n, true
		}
	}
	return Node{}, false
}

type Ht []Node

func (pq Ht) Len() int { return len(pq) }

func (pq Ht) Empty() bool { return len(pq) == 0 }

func (pq Ht) Less(i, j int) bool {
	return pq[i].dValue < pq[j].dValue
}

func (pq Ht) Swap(i, j int) {
	iNew := pq[j]
	iNew.htIndex = i
	jNew := pq[i]
	jNew.htIndex = j
	pq[i], pq[j] = iNew, jNew
}

func (pq *Ht) Push(x interface{}) {
	node := x.(Node)
	node.htIndex = len(*pq)
	*pq = append(*pq, node)
}

func (pq *Ht) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	*pq = old[0 : n-1]
	return node
}

func (pq Ht) Top() (node interface{}, found bool) {
	if len(pq) == 0 {
		return Node{}, false
	}
	return pq[0], true
}

func (pq Ht) getNode(u, v int) (n Node, found bool) {
	for _, n := range pq {
		if n.u == u && n.v == v {
			return n, true
		}
	}
	return Node{}, false
}
