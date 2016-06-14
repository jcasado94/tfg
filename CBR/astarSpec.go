package CBR

import (
	"container/heap"
	// "fmt"
	"github.com/jcasado94/tfg/common"
	"github.com/jmcvetta/neoism"
	"strconv"
)

type AstarNode struct {
	Id        int
	Transfers int          // transfers
	Prevs     map[int]bool // previous cities
}

type Astar struct {
	pq []AstarNode

	// which vertices are currently open
	openVertices map[int]bool

	// which vertices are closed
	closedVertices map[int]bool

	// cost from start to node
	gScore map[int]float64

	// cost from start to end, passing by node i (g+h)
	fScore map[int]float64

	heuristic map[int]map[int]float64

	admissible, consistent bool
	seconds                float64
}

type AstarSpec struct {
	Astar

	// incoming edges to the relationship, indexed by relationship id. Their weight is equal to the relationship price, except for end node, with price 0.
	incomingEdges map[int][]int

	H *SameDayCombinationsHandler
}

func (as Astar) Len() int { return len(as.pq) }

func (as Astar) Empty() bool { return len(as.pq) == 0 }

func (as Astar) Less(i, j int) bool {
	return as.fScore[as.pq[i].Id] < as.fScore[as.pq[j].Id]
}

func (as Astar) Swap(i, j int) {
	as.pq[i], as.pq[j] = as.pq[j], as.pq[i]
}

func (as *Astar) Push(x interface{}) {
	(*as).pq = append((*as).pq, x.(AstarNode))
}

func (as *Astar) Pop() interface{} {
	old := (*as).pq
	n := len(old)
	rel := old[n-1]
	(*as).pq = old[0 : n-1]
	return rel
}

func (as Astar) Top() interface{} {
	return as.pq[0]
}

// if start = true, we would look for arr vertex.
// if start = false, we iterate until we double the current found vertices.
func (as *AstarSpec) GoAStar(arr int, start bool) bool {

	i := 0
	n := len(as.closedVertices) * 2

	for !as.Empty() {

		if !start {
			if i == n {
				return true
			}
		}

		current := heap.Pop(as).(AstarNode)
		currentId := current.Id
		currentTransfers := current.Transfers
		currentRel := as.H.Graph.Rels[currentId]
		var hin Hin
		heap.Init(&hin)
		for _, edge := range as.incomingEdges[currentId] {
			if currentRel.CameFrom != edge {
				// not search tree node
				// create Hin
				var node Node
				node.u = edge
				node.v = currentId
				node.hin = currentId
				node.dValue = as.gScore[edge] + as.H.Graph.Rels[currentId].Weight - as.gScore[currentId] // g(u) + c(u,v) - g(v)
				node.ht = NULL_HT
				heap.Push(&hin, node)
			}
			as.H.Graph.Hins[currentId] = hin
		}

		as.openVertices[currentId] = false
		if !as.closedVertices[currentId] {
			i++
		}
		as.closedVertices[currentId] = true

		if start && currentId == END_ID {
			return true
		}

		// every relationship that departs from the end node is a potential new vertex
		dbNode, _ := as.H.Db.Node(currentRel.ArrNode)
		dbNode.Db = as.H.Db
		// relsDb, _ := dbNode.Outgoing("SPEC")
		relsDb := as.getRels(currentRel.ArrNode)

		lastNode := currentRel.ArrNode == arr
		firstNode := currentId == START_ID

		for i := 0; i < len(relsDb) || lastNode; i++ {

			var nextRelId int
			if lastNode {
				nextRelId = END_ID
			} else {
				nextRelId = relsDb[i].Id()
			}
			var nextRel Rel
			nextRel = as.H.Graph.Rels[nextRelId]
			if nextRel.Id == 0 {
				if lastNode {
					nextRel = Rel{Id: nextRelId, Weight: 0.0, DepNode: arr}
					as.H.Graph.Rels[nextRelId] = nextRel
				} else {
					nextRel = as.H.Graph.GetRel(&relsDb[i])
					as.H.Graph.Rels[nextRelId] = nextRel
				}
			}

			if firstNode && (nextRel.DepTime.Day() != as.H.Kstar.DepartureTime.Day() || nextRel.DepTime.Month() != as.H.Kstar.DepartureTime.Month() || nextRel.DepTime.Year() != as.H.Kstar.DepartureTime.Year()) {
				// only trips starting the requested day
				continue
			}

			_, cycling := current.Prevs[nextRel.DepNode]

			if !lastNode && !firstNode && (cycling /*|| currentTransfers == common.MAX_TRANSFERS */ || nextRel.DepTime.Sub(currentRel.ArrTime).Minutes() < common.TRANSFER_TIME || nextRel.DepTime.Sub(currentRel.ArrTime).Hours() > common.MAX_TRANSFER_HOURS) {
				// not a good transfer
				continue
			}

			as.incomingEdges[nextRelId] = append(as.incomingEdges[nextRelId], currentId)

			if as.closedVertices[nextRelId] {

				// still we add the new sidetrack edge
				var hin Hin
				for _, node := range as.H.Graph.Hins[nextRelId] {
					hin = append(hin, node)
				}
				if len(hin) == 0 {
					heap.Init(&hin)
				}
				node := Node{u: currentId, v: nextRelId, dValue: as.gScore[currentId] + as.H.Graph.Rels[nextRelId].Weight - as.gScore[nextRelId], hin: nextRelId, ht: NULL_HT}
				pair := len(hin)%2 == 0
				heap.Push(&hin, node)

				tentativeGScore := as.gScore[currentId] + nextRel.Weight
				if tentativeGScore < as.gScore[nextRelId] {
					as.consistent = false
					// fmt.Println("not consistent")
				}

				// // check if node addition is interfering with the dijkstra execution
				add := true
				if !start {
					for k := range hin {
						node := hin[k]
						if node.u == currentId && node.v == nextRelId {
							if pair && k < len(hin)-2 {
								for k1 := k + 1; k1 < len(hin); k1++ {
									if hin[k1].usedInDijkstra {
										add = false
									}
								}
							} else if k < len(hin)-1 {
								for k1 := k + 1; k1 < len(hin); k1++ {
									if hin[k1].usedInDijkstra {
										add = false
									}
								}
							}
						}
					}
				}
				if !start && add {
					as.H.Graph.Hins[nextRelId] = hin
				} else if start {
					as.H.Graph.Hins[nextRelId] = hin
				}

				if lastNode {
					break
				}

				continue
			}

			tentativeGScore := as.gScore[currentId] + nextRel.Weight

			if _, exists := as.gScore[nextRelId]; exists && tentativeGScore >= as.gScore[nextRelId] {
				if lastNode {
					break
				}
				continue
			}

			// update scores
			as.gScore[nextRelId] = tentativeGScore
			if lastNode {
				as.fScore[nextRelId] = tentativeGScore
			} else {
				as.fScore[nextRelId] = tentativeGScore + as.getHeuristicValue(nextRel.ArrNode, arr)
			}

			nextRel.CameFrom = currentId
			as.H.Graph.Rels[nextRelId] = nextRel

			newPrevs := make(map[int]bool)
			for k := range current.Prevs {
				newPrevs[k] = true
			}
			newPrevs[currentRel.DepNode] = true

			if !as.openVertices[nextRelId] {
				heap.Push(as, AstarNode{Id: nextRelId, Transfers: currentTransfers + 1, Prevs: newPrevs})
				as.openVertices[nextRelId] = true
			} else {
				// fix the heap
				for i := range as.pq {
					if as.pq[i].Id == nextRelId {
						as.pq[i] = AstarNode{Id: nextRelId, Transfers: currentTransfers + 1, Prevs: newPrevs}
						heap.Fix(as, i)
					}
				}
			}

			if lastNode {
				break
			}

		}

	}

	return false

}

func (as Astar) getHeuristicValue(a, b int) float64 {
	if a == b {
		return 0.0
	}
	x := as.heuristic[a][b]
	if x == 0.0 {
		// theoretically, with 2 combinations (A->Buenos Aires->B), it's highly probable that we could get till b. So the average price of 2 trips -50% is returned.
		return 400
	} else {
		return x * 0.5
	}
	return 0.0
}

func (as AstarSpec) getRels(depId int) []neoism.Relationship {
	props := neoism.Props{}
	var WHERE = "WHERE id(a)={id} AND "
	props["id"] = depId
	if depId == as.H.Kstar.depId {
		WHERE += "id(b)<>{idArr} AND "
		props["idArr"] = as.H.Kstar.arrId
	}
	if as.H.Kstar.difMonth {
		WHERE += "( "
	}
	WHERE += "r.depYear={Y1} AND r.depMonth={M1} AND ("
	props["Y1"] = as.H.Kstar.yearsLookup[0]
	props["M1"] = as.H.Kstar.monthsLookup[0]
	var lastDay, i int
	for _, day := range as.H.Kstar.daysLookup {
		if day < lastDay {
			continue
		}
		dayKey := "D" + strconv.Itoa(day)
		props[dayKey] = day
		if i != 0 {
			WHERE += "OR "
		}
		WHERE += "r.depDay={" + dayKey + "} "
		lastDay = day
		i++
	}
	WHERE += ") "
	if as.H.Kstar.difMonth { //another month
		WHERE += " OR "
		if len(as.H.Kstar.yearsLookup) > 1 {
			WHERE += "r.depYear={Y2} AND "
			props["Y2"] = as.H.Kstar.yearsLookup[1]
		}
		WHERE += "r.depMonth={M2} AND ("
		props["M2"] = as.H.Kstar.monthsLookup[1]
		ii := i
		for i < len(as.H.Kstar.daysLookup) {
			dayKey := "D" + strconv.Itoa(as.H.Kstar.daysLookup[i])
			props[dayKey] = as.H.Kstar.daysLookup[i]
			if i != ii {
				WHERE += "OR "
			}
			WHERE += "r.depDay={" + dayKey + "} "
			i++
		}
		WHERE += ") )"
	}

	type answer struct {
		Rel neoism.Relationship `json:"r"`
	}
	var rels []answer
	cq := neoism.CypherQuery{
		Statement: `
			MATCH (a:City)-[r:SPEC]->(b:City) 
			` + WHERE +
			` RETURN r LIMIT 100
		`,
		Parameters: props,
		Result:     &rels,
	}
	err := as.H.Db.Cypher(&cq)
	common.PanicErr(err)

	var ans []neoism.Relationship
	for _, rel := range rels {
		ans = append(ans, rel.Rel)
	}
	return ans

}
