package CBR

import (
	"container/heap"
	"encoding/gob"
	// "fmt"
	"github.com/jcasado94/tfg/common"
	"os"
	"time"
)

type Kstar struct {
	openD []Node
}

type KstarSpec struct {
	Kstar

	res [][]int

	DepartureTime time.Time

	yearsLookup, monthsLookup, daysLookup []int
	difMonth                              bool

	H *SameDayCombinationsHandler
}

func (ks Kstar) Len() int { return len(ks.openD) }

func (ks Kstar) Empty() bool { return len(ks.openD) == 0 }

// func (ks KstarSpec) Less(i, j int) bool {
// 	iDist := ks.openD[i].dist
// 	jDist := ks.openD[j].dist

// 	return iDist < jDist
// }

func (ks Kstar) Less(i, j int) bool {
	iDist := ks.openD[i].dist
	jDist := ks.openD[j].dist

	return iDist < jDist
}

func (ks Kstar) Swap(i, j int) {
	ks.openD[i], ks.openD[j] = ks.openD[j], ks.openD[i]
}

func (ks *Kstar) Push(x interface{}) {
	(*ks).openD = append((*ks).openD, x.(Node))
}

func (ks *Kstar) Pop() interface{} {
	old := (*ks).openD
	n := len(old)
	rel := old[n-1]
	(*ks).openD = old[0 : n-1]
	return rel
}

func (ks Kstar) Top() interface{} {
	return ks.openD[0]
}

const INF = 1000000.0

func (ks KstarSpec) GoKStar(dep, arr int) [][]int {

	ks.H.initVars(dep, arr)

	succesful := ks.H.Astar.goAStar(arr, true)
	if !succesful {
		return [][]int{}
	}

	heap.Init(&ks)

	// create R
	var r = Node{dValue: 0.0, u: END_ID, v: R, ht: R, dist: 0.0}

	heap.Push(&ks, r)

	n := ks.Top().(Node)

	ks.H.Graph.succCrossEdges(n)

	// while A* can still work
	for !ks.H.Astar.Empty() || !ks.Empty() {

		if ks.schedulingMechanismEnabled() {

			if !ks.Empty() {

				// check scheduling condition
				uNode := ks.H.Astar.Top().(AstarNode)
				u := uNode.Id
				n := ks.Top().(Node)

				var d = 0.0

				var succ = append(ks.H.Graph.succCrossEdges(n), ks.H.Graph.succHeapEdges(n)...)

				for _, n2 := range succ {
					delta := getDelta(&n, &n2)
					if delta > d {
						d = delta
					}
				}

				// scheduling condition
				if ks.H.Astar.gScore[arr]+d <= ks.H.Astar.fScore[u] {

					ks.dijkstraStep()

					if len(ks.res) == K {
						// fmt.Println(ks.res)

						return ks.getFinalPath()
						// return ks.res
					}

				} else {
					// resume A* to explore more graph
					ks.resumeAstar(arr)
				}

			} else {
				ks.resumeAstar(arr)
			}

		} else {

			ks.dijkstraStep()

			if len(ks.res) == K {
				// fmt.Println(ks.res)

				return ks.getFinalPath()
				// return ks.res
			}

		}

	}

	// fmt.Println(ks.getFinalPath())

	return ks.getFinalPath()
	// return ks.res

}

func (ks KstarSpec) getFinalPath() [][]int {

	type Pair struct {
		nextCities map[int]*Pair
		transps    []int // type of transp of the rels arriving to the city
	}

	intInSlice := func(a int, sl []int) bool {
		for _, x := range sl {
			if x == a {
				return true
			}
		}
		return false
	}

	var mainCity = &Pair{transps: []int{}, nextCities: make(map[int]*Pair)}

	relsPaths := ks.res
	var res [][]int

	for _, relsPath := range relsPaths {

		transpAnt := 0
		found := true
		i := len(relsPath) - 2
		city := mainCity

		for i > 0 {

			rel := relsPath[i]
			dep := ks.H.Graph.Rels[rel].DepNode

			if _, exists := city.nextCities[dep]; !exists {
				found = false
				city.nextCities[dep] = &Pair{nextCities: make(map[int]*Pair), transps: []int{transpAnt}}
				city = city.nextCities[dep]
			} else if !intInSlice(transpAnt, city.nextCities[dep].transps) {
				found = false
				city.nextCities[dep].transps = append(city.nextCities[dep].transps, transpAnt)
				city = city.nextCities[dep]
			} else {
				city = city.nextCities[dep]
			}

			if i == 1 {
				arr := ks.H.Graph.Rels[rel].ArrNode
				transp := ks.H.Graph.Rels[rel].Transp
				if _, exists := city.nextCities[arr]; !exists {
					found = false
					city.nextCities[arr] = &Pair{nextCities: make(map[int]*Pair), transps: []int{transp}}
				} else if !intInSlice(transp, city.nextCities[arr].transps) {
					found = false
					city.nextCities[arr].transps = append(city.nextCities[arr].transps, transp)
				}
			}

			transpAnt = ks.H.Graph.Rels[rel].Transp
			i--

		}

		if !found {
			res = append(res, relsPath)
		}

	}

	return res
}

func (ks KstarSpec) schedulingMechanismEnabled() bool {
	return len(ks.H.Astar.pq) > 0
}

func (ks *KstarSpec) resumeAstar(arr int) {

	ks.H.Astar.goAStar(arr, false)

	// delete and rebuild the hts
	for i := range ks.H.Graph.Hts {
		delete(ks.H.Graph.Hts, i)
		ks.H.Graph.createHt(i)
	}

	// look for new unexplored children of closed vertices
	for h, hm := range ks.H.Graph.Closed {
		for u, um := range hm {
			for v, _ := range um {
				var node Node
				if h == R {
					node = Node{dValue: 0.0, u: END_ID, v: R, ht: R}
				} else {
					var found bool
					node, found = ks.H.Graph.Hts[h].getNode(u, v)
					if !found {
						node, _ = ks.H.Graph.Hins[h].getNode(u, v)
					}
				}

				succ := append(ks.H.Graph.succCrossEdges(node), ks.H.Graph.succHeapEdges(node)...)

				for _, s := range succ {

					h := s.getHtOrHin()

					if ks.H.Graph.Closed[h][s.u][s.v] || ks.H.Graph.Explored[h][s.u][s.v] {
						continue
					}

					// found a new added vertex. check scheduling condition.

					doDijkstraStep := true

					d := node.dist
					if !isCross(&node, &s) {
						// heap edge
						d += s.dValue - node.dValue
					} else {
						d += s.dValue
					}
					if ks.schedulingMechanismEnabled() {

						uNode := ks.H.Astar.Top().(AstarNode)
						u := uNode.Id
						f := d + ks.H.Astar.gScore[arr]

						doDijkstraStep = f <= ks.H.Astar.fScore[u]

					}

					if !doDijkstraStep {

						ks.resumeAstar(arr)

					} else {

						// proceed with dijkstra

						if _, exists := ks.H.Graph.Explored[h]; !exists {
							ks.H.Graph.Explored[h] = make(map[int]map[int]bool)
						}
						if _, exists := ks.H.Graph.Explored[h][s.u]; !exists {
							ks.H.Graph.Explored[h][s.u] = make(map[int]bool)
						}
						if _, exists := ks.H.Graph.Parents[h]; !exists {
							ks.H.Graph.Parents[h] = make(map[int]map[int][]Node)
						}
						if _, exists := ks.H.Graph.Parents[h][s.u]; !exists {
							ks.H.Graph.Parents[h][s.u] = make(map[int][]Node)
						}

						ks.H.Graph.Explored[h][s.u][s.v] = true
						ks.H.Graph.Parents[h][s.u][s.v] = append(ks.H.Graph.Parents[h][s.u][s.v], node)
						s.indParent = len(ks.H.Graph.Parents[h][s.u][s.v]) - 1

						s.dist = d

						heap.Push(ks, s)

					}
				}
			}
		}
	}

}

func (ks *KstarSpec) dijkstraStep() {

	next := heap.Pop(ks).(Node)

	h := next.getHtOrHin()

	if _, exists := ks.H.Graph.Closed[h]; !exists {
		ks.H.Graph.Closed[h] = make(map[int]map[int]bool)
	}
	if _, exists := ks.H.Graph.Closed[h][next.u]; !exists {
		ks.H.Graph.Closed[h][next.u] = make(map[int]bool)
	}
	ks.H.Graph.Closed[h][next.u][next.v] = true

	succ := append(ks.H.Graph.succCrossEdges(next), ks.H.Graph.succHeapEdges(next)...)

	for _, n2 := range succ {

		h2 := n2.getHtOrHin()
		n2.dist = next.dist + getDelta(&next, &n2)

		if _, exists := ks.H.Graph.Parents[h2]; !exists {
			ks.H.Graph.Parents[h2] = make(map[int]map[int][]Node)
		}
		if _, exists := ks.H.Graph.Parents[h2][n2.u]; !exists {
			ks.H.Graph.Parents[h2][n2.u] = make(map[int][]Node)
		}
		ks.H.Graph.Parents[h2][n2.u][n2.v] = append(ks.H.Graph.Parents[h2][n2.u][n2.v], next)
		n2.indParent = len(ks.H.Graph.Parents[h2][n2.u][n2.v]) - 1

		if _, exists := ks.H.Graph.Explored[h2]; !exists {
			ks.H.Graph.Explored[h2] = make(map[int]map[int]bool)
		}
		if _, exists := ks.H.Graph.Explored[h2][n2.u]; !exists {
			ks.H.Graph.Explored[h2][n2.u] = make(map[int]bool)
		}
		ks.H.Graph.Explored[h2][n2.u][n2.v] = true

		heap.Push(ks, n2)

	}

	path := ks.getPath(ks.getTetaSeq(&next))

	if len(path) > 3 { // avoid direct trips
		ks.res = append(ks.res, path)
	}
}

func (ks KstarSpec) getPath(tetaSeq [][2]int) []int {

	res := []int{END_ID}

	for res[len(res)-1] != START_ID {
		for len(tetaSeq) > 0 && tetaSeq[len(tetaSeq)-1][1] == res[len(res)-1] {
			// pop
			st := tetaSeq[len(tetaSeq)-1][0]
			res = append(res, st)
			tetaSeq = tetaSeq[:len(tetaSeq)-1]
		}
		prevTree := ks.H.Graph.Rels[res[len(res)-1]].CameFrom
		res = append(res, prevTree)
	}

	return res

}

func (ks KstarSpec) getTetaSeq(node *Node) [][2]int {

	if node.v == R {
		return [][2]int{}
	}

	var res [][2]int
	res = append(res, [2]int{node.u, node.v})

	lastNode := node

	for {
		parent := ks.H.Graph.Parents[lastNode.getHtOrHin()][lastNode.u][lastNode.v][lastNode.indParent]
		if parent.v == R {
			return res
		}
		if isCross(&parent, lastNode) {
			res = append(res, [2]int{parent.u, parent.v})
		}
		lastNode = &parent
	}

	return res
}

func (h *SameDayCombinationsHandler) initVars(dep, arr int) {

	(*h).Graph.Rels = make(map[int]Rel)
	(*h).Graph.Hins = make(map[int]Hin)
	(*h).Graph.Hts = make(map[int]Ht)
	(*h).Graph.Closed = make(map[int]map[int]map[int]bool)
	(*h).Graph.Explored = make(map[int]map[int]map[int]bool)
	(*h).Graph.Parents = make(map[int]map[int]map[int][]Node)

	(*h).Graph.Rels[START_ID] = Rel{Id: START_ID, ArrNode: dep, Weight: 0.0, ArrTime: (*h).Kstar.DepartureTime}

	(*h).Astar.pq = make([]AstarNode, 1)
	(*h).Astar.pq[0] = AstarNode{Id: START_ID, Transfers: -1, Prevs: make(map[int]bool)}
	heap.Init(&(*h).Astar)

	(*h).Astar.closedVertices = make(map[int]bool)
	(*h).Astar.openVertices = make(map[int]bool)

	(*h).Astar.gScore = make(map[int]float64)
	(*h).Astar.fScore = make(map[int]float64)

	(*h).Astar.incomingEdges = make(map[int][]int)

	(*h).Astar.gScore[START_ID] = 0
	(*h).Astar.fScore[START_ID] = (*h).Astar.getHeuristicValue(dep, arr)
	(*h).Astar.openVertices[START_ID] = true

	dataFile, err := os.Open(common.FILE_AVERAGES)
	common.PanicErr(err)
	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&(*h).Astar.tripAverages)
	common.PanicErr(err)
	dataFile.Close()

	(*h).Astar.H = h
	(*h).Kstar.H = h
	(*h).Graph.H = h

}
