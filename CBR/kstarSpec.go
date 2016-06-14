package CBR

import (
	"container/heap"
	"encoding/gob"
	// "fmt"
	"github.com/jcasado94/tfg/common"
	"github.com/jmcvetta/neoism"
	"os"
	"sync"
	"time"
)

var mutexSpecHeuristic sync.Mutex

type Kstar struct {
	openD []Node
}

type KstarSpec struct {
	Kstar

	res [][]int

	depId, arrId  int
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

	t1 := time.Now()
	succesful := ks.H.Astar.GoAStar(arr, true)
	ks.H.Astar.seconds += time.Now().Sub(t1).Seconds()
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

	type TranspDay struct {
		transp int
		day    int
	}

	type Pair struct {
		nextCities  map[int]*Pair
		transpsDays []TranspDay // type of transp of the rels arriving to the city
	}

	inSlice := func(transp int, day int, sl []TranspDay) bool {
		for _, x := range sl {
			if x.transp == transp && x.day == day {
				return true
			}
		}
		return false
	}

	var mainCity = &Pair{transpsDays: []TranspDay{}, nextCities: make(map[int]*Pair)}

	relsPaths := ks.res
	var res [][]int

	for _, relsPath := range relsPaths {

		dayAnt := 0
		transpAnt := 0
		found := true
		i := len(relsPath) - 2
		city := mainCity

		for i > 0 {

			rel := relsPath[i]
			dep := ks.H.Graph.Rels[rel].DepNode

			if _, exists := city.nextCities[dep]; !exists {
				found = false
				city.nextCities[dep] = &Pair{nextCities: make(map[int]*Pair), transpsDays: []TranspDay{TranspDay{transpAnt, dayAnt}}}
				city = city.nextCities[dep]
			} else if !inSlice(transpAnt, dayAnt, city.nextCities[dep].transpsDays) {
				found = false
				city.nextCities[dep].transpsDays = append(city.nextCities[dep].transpsDays, TranspDay{transpAnt, dayAnt})
				city = city.nextCities[dep]
			} else {
				city = city.nextCities[dep]
			}

			if i == 1 {
				arr := ks.H.Graph.Rels[rel].ArrNode
				transp := ks.H.Graph.Rels[rel].Transp
				day := ks.H.Graph.Rels[rel].DepTime.Day()
				if _, exists := city.nextCities[arr]; !exists {
					found = false
					city.nextCities[arr] = &Pair{nextCities: make(map[int]*Pair), transpsDays: []TranspDay{TranspDay{transp, day}}}
				} else if !inSlice(transp, day, city.nextCities[arr].transpsDays) {
					found = false
					city.nextCities[arr].transpsDays = append(city.nextCities[arr].transpsDays, TranspDay{transp, day})
				}
			}

			transpAnt = ks.H.Graph.Rels[rel].Transp
			dayAnt = ks.H.Graph.Rels[rel].DepTime.Day()
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

	t1 := time.Now()
	ks.H.Astar.GoAStar(arr, false)
	ks.H.Astar.seconds += time.Now().Sub(t1).Seconds()

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
						node, found = ks.H.Graph.Hins[h].getNode(u, v)
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

					d := node.dist + getDelta(&node, &s)
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

		n2.usedInDijkstra = true
		if n2.v != R {
			ks.H.Graph.Hins[n2.hin][n2.hinIndex] = n2
		}

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
	ks.updateHeuristic(path, ks.H.Db)
	ks.res = append(ks.res, path)
}

func (ks KstarSpec) updateHeuristic(path []int, db *neoism.Database) {
	// get path price
	totalPrice := 0.0
	for i := 1; i < len(path)-1; i++ {
		rel, _ := db.Relationship(path[i])
		rel.Db = db
		props, _ := rel.Properties()
		totalPrice += props["price"].(float64)
	}
	var heuristics map[int]map[int]float64

	mutexSpecHeuristic.Lock()
	dataFile, _ := os.Open(common.FILE_SPEC_HEURISTIC)
	dataDecoder := gob.NewDecoder(dataFile)
	_ = dataDecoder.Decode(&heuristics)
	dataFile.Close()
	if heuristics[ks.depId][ks.arrId] > totalPrice || heuristics[ks.depId][ks.arrId] == 0.0 {
		heuristics[ks.depId][ks.arrId] = totalPrice
		dataFile, err := os.Create(common.FILE_SPEC_HEURISTIC)
		common.PanicErr(err)
		dataEncoder := gob.NewEncoder(dataFile)
		dataEncoder.Encode(heuristics)
	}
	dataFile.Close()
	mutexSpecHeuristic.Unlock()

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

	(*h).Kstar.depId = dep
	(*h).Kstar.arrId = arr

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
	(*h).Astar.admissible, (*h).Astar.consistent = true, true
	(*h).Astar.seconds = 0

	(*h).Astar.gScore = make(map[int]float64)
	(*h).Astar.fScore = make(map[int]float64)

	(*h).Astar.incomingEdges = make(map[int][]int)

	(*h).Astar.gScore[START_ID] = 0
	(*h).Astar.fScore[START_ID] = (*h).Astar.getHeuristicValue(dep, arr)
	(*h).Astar.openVertices[START_ID] = true

	dataFile, err := os.Open(common.FILE_SPEC_HEURISTIC)
	common.PanicErr(err)
	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&(*h).Astar.heuristic)
	common.PanicErr(err)
	dataFile.Close()

	(*h).Astar.H = h
	(*h).Kstar.H = h
	(*h).Graph.H = h

}
