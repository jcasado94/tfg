package CBR

import (
	"container/heap"
	"encoding/gob"
	// "fmt"
	"github.com/jcasado94/tfg/common"
	"os"
)

type KstarGen struct {
	Kstar
	res [][][2]int
	H   *UsualCombinationsHandler
}

func (ks KstarGen) Less(i, j int) bool {
	iDist, jDist := ks.openD[i].dist, ks.openD[j].dist

	return iDist < jDist
}

func (ks KstarGen) GoKStar(dep, arr int) [][][2]int {

	ks.H.initVars(dep, arr)

	succesful := ks.H.Astar.goAStar(arr, true)
	if !succesful {
		return [][][2]int{}
	}

	heap.Init(&ks)

	// create R
	var r = Node{dValue: 0.0, u: arr, v: R, ht: R, dist: 0.0}

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

						// return ks.getFinalPath()
						return ks.res
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

				// return ks.getFinalPath()
				return ks.res
			}

		}

	}

	// return ks.getFinalPath()
	return ks.res

}

func (ks KstarGen) schedulingMechanismEnabled() bool {
	return len(ks.H.Astar.pq) > 0
}

func (ks *KstarGen) resumeAstar(arr int) {

	ks.H.Astar.goAStar(arr, false)

	// delete and rebuild the hts
	for i := range ks.H.Graph.Hts {
		delete(ks.H.Graph.Hts, i)
		ks.H.Graph.createHt(i)
	}

	// look for new unexplored children of closed vertices
	for h, hm := range ks.H.Graph.Closed {
		for u, um := range hm {
			for v, vm := range um {
				for t, _ := range vm {
					var node Node
					if h == R {
						node = Node{dValue: 0.0, u: ks.H.Graph.ArrId, v: R, ht: R, dist: 0.0}
					} else {
						var found bool
						node, found = getNodeHt(u, v, t, ks.H.Graph.Hts[h])
						if !found {
							node, _ = getNodeHin(u, v, t, ks.H.Graph.Hins[h])
						}
					}

					succ := append(ks.H.Graph.succCrossEdges(node), ks.H.Graph.succHeapEdges(node)...)

					for _, s := range succ {

						h := s.getHtOrHin()

						if ks.H.Graph.Closed[h][s.u][s.v][s.transp] || ks.H.Graph.Explored[h][s.u][s.v][s.transp] {
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
								ks.H.Graph.Explored[h] = make(map[int]map[int]map[int]bool)
							}
							if _, exists := ks.H.Graph.Explored[h][s.u]; !exists {
								ks.H.Graph.Explored[h][s.u] = make(map[int]map[int]bool)
							}
							if _, exists := ks.H.Graph.Explored[h][s.u][s.v]; !exists {
								ks.H.Graph.Explored[h][s.u][s.v] = make(map[int]bool)
							}
							if _, exists := ks.H.Graph.Parents[h]; !exists {
								ks.H.Graph.Parents[h] = make(map[int]map[int]map[int][]Node)
							}
							if _, exists := ks.H.Graph.Parents[h][s.u]; !exists {
								ks.H.Graph.Parents[h][s.u] = make(map[int]map[int][]Node)
							}
							if _, exists := ks.H.Graph.Parents[h][s.u][s.v]; !exists {
								ks.H.Graph.Parents[h][s.u][s.v] = make(map[int][]Node)
							}

							ks.H.Graph.Explored[h][s.u][s.v][s.transp] = true
							ks.H.Graph.Parents[h][s.u][s.v][s.transp] = append(ks.H.Graph.Parents[h][s.u][s.v][s.transp], node)
							s.indParent = len(ks.H.Graph.Parents[h][s.u][s.v][s.transp]) - 1
							s.dist = d

							heap.Push(ks, s)

						}
					}
				}
			}
		}
	}

}

func (ks *KstarGen) dijkstraStep() {

	next := heap.Pop(ks).(Node)

	h := next.getHtOrHin()

	if _, exists := ks.H.Graph.Closed[h]; !exists {
		ks.H.Graph.Closed[h] = make(map[int]map[int]map[int]bool)
	}
	if _, exists := ks.H.Graph.Closed[h][next.u]; !exists {
		ks.H.Graph.Closed[h][next.u] = make(map[int]map[int]bool)
	}
	if _, exists := ks.H.Graph.Closed[h][next.u][next.v]; !exists {
		ks.H.Graph.Closed[h][next.u][next.v] = make(map[int]bool)
	}
	ks.H.Graph.Closed[h][next.u][next.v][next.transp] = true

	succ := append(ks.H.Graph.succCrossEdges(next), ks.H.Graph.succHeapEdges(next)...)

	for _, n2 := range succ {

		h2 := n2.getHtOrHin()
		n2.dist = next.dist + getDelta(&next, &n2)

		if _, exists := ks.H.Graph.Parents[h2]; !exists {
			ks.H.Graph.Parents[h2] = make(map[int]map[int]map[int][]Node)
		}
		if _, exists := ks.H.Graph.Parents[h2][n2.u]; !exists {
			ks.H.Graph.Parents[h2][n2.u] = make(map[int]map[int][]Node)
		}
		if _, exists := ks.H.Graph.Parents[h2][n2.u][n2.v]; !exists {
			ks.H.Graph.Parents[h2][n2.u][n2.v] = make(map[int][]Node)
		}
		ks.H.Graph.Parents[h2][n2.u][n2.v][n2.transp] = append(ks.H.Graph.Parents[h2][n2.u][n2.v][n2.transp], next)
		n2.indParent = len(ks.H.Graph.Parents[h2][n2.u][n2.v][n2.transp]) - 1

		if _, exists := ks.H.Graph.Explored[h2]; !exists {
			ks.H.Graph.Explored[h2] = make(map[int]map[int]map[int]bool)
		}
		if _, exists := ks.H.Graph.Explored[h2][n2.u]; !exists {
			ks.H.Graph.Explored[h2][n2.u] = make(map[int]map[int]bool)
		}
		if _, exists := ks.H.Graph.Explored[h2][n2.u][n2.v]; !exists {
			ks.H.Graph.Explored[h2][n2.u][n2.v] = make(map[int]bool)
		}
		ks.H.Graph.Explored[h2][n2.u][n2.v][n2.transp] = true

		heap.Push(ks, n2)

	}

	path := ks.getPath(ks.getTetaSeq(&next))

	if len(path) > 2 { // avoid direct trips
		ks.res = append(ks.res, path)
	}

}

func (ks KstarGen) getPath(tetaSeq [][3]int) [][2]int { // {nodeId, arrivingTransp}

	res := [][2]int{[2]int{ks.H.Graph.ArrId, ks.H.Graph.Cities[ks.H.Graph.ArrId].Transp}}

	done := false
	for !done {
		for len(tetaSeq) > 0 && tetaSeq[len(tetaSeq)-1][1] == res[len(res)-1][0] {
			// pop
			transp := tetaSeq[len(tetaSeq)-1][2]
			res[len(res)-1][1] = transp

			u := tetaSeq[len(tetaSeq)-1][0]
			res = append(res, [2]int{u, ks.H.Graph.Cities[u].Transp})
			tetaSeq = tetaSeq[:len(tetaSeq)-1]
		}
		if res[len(res)-1][0] == ks.H.Graph.DepId {
			done = true
			break
		}
		prevTree := ks.H.Graph.Cities[res[len(res)-1][0]]
		cameFrom := prevTree.CameFrom
		cameFromTransp := ks.H.Graph.Cities[cameFrom].Transp
		res = append(res, [2]int{prevTree.CameFrom, cameFromTransp})
	}

	return res

}

func (ks KstarGen) getTetaSeq(node *Node) [][3]int { // {uId, vId, transp}

	if node.v == R {
		return [][3]int{}
	}

	var res [][3]int
	res = append(res, [3]int{node.u, node.v, node.transp})

	lastNode := node

	for {
		parent := ks.H.Graph.Parents[lastNode.getHtOrHin()][lastNode.u][lastNode.v][lastNode.transp][lastNode.indParent]
		if parent.v == R {
			return res
		}
		if isCross(&parent, lastNode) {
			res = append(res, [3]int{parent.u, parent.v, parent.transp})
		}
		lastNode = &parent
	}

	return res
}

func (h *UsualCombinationsHandler) initVars(dep, arr int) {

	(*h).Graph.Cities = make(map[int]City)
	(*h).Graph.Hins = make(map[int]Hin)
	(*h).Graph.Hts = make(map[int]Ht)
	(*h).Graph.Closed = make(map[int]map[int]map[int]map[int]bool)
	(*h).Graph.Explored = make(map[int]map[int]map[int]map[int]bool)
	(*h).Graph.Parents = make(map[int]map[int]map[int]map[int][]Node)

	(*h).Graph.Cities[dep] = City{Id: dep, Transp: 0, CameFrom: dep}
	(*h).Graph.DepId, (*h).Graph.ArrId = dep, arr

	(*h).Astar.pq = make([]AstarNode, 1)
	(*h).Astar.pq[0] = AstarNode{Id: dep, Transfers: -1, Prevs: make(map[int]bool)}
	heap.Init(&(*h).Astar)

	(*h).Astar.closedVertices = make(map[int]bool)
	(*h).Astar.openVertices = make(map[int]bool)

	(*h).Astar.gScore = make(map[int]float64)
	(*h).Astar.fScore = make(map[int]float64)

	(*h).Astar.incomingEdges = make(map[int][]EdgeGen)

	(*h).Astar.gScore[dep] = 0
	(*h).Astar.fScore[dep] = (*h).Astar.getHeuristicValue(dep, arr)
	(*h).Astar.openVertices[dep] = true

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
