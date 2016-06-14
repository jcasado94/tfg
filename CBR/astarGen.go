package CBR

import (
	"container/heap"
	// "github.com/jcasado94/tfg/common"
	// "sort"
	// "fmt"
)

type EdgeGen struct {
	idNode, transp int
	weight         float64
}

type AstarGen struct {
	Astar

	incomingEdges map[int][]EdgeGen

	H *UsualCombinationsHandler
}

// if start = true, we would look for arr vertex.
// if start = false, we iterate until we double the current found vertices.
func (as *AstarGen) GoAStar(arr int, start bool) bool {

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
		if as.closedVertices[currentId] {
			continue
		}
		currentTransfers := current.Transfers
		currentCity := as.H.Graph.Cities[currentId]
		var hin Hin
		heap.Init(&hin)
		for _, edge := range as.incomingEdges[currentId] {
			if currentCity.CameFrom != edge.idNode || currentCity.Transp != edge.transp {
				// not search tree node
				// create Hin
				var node Node
				node.u = edge.idNode
				node.v = currentId
				node.hin = currentId
				node.dValue = as.gScore[edge.idNode] + edge.weight - as.gScore[currentId] // g(u) + c(u,v) - g(v)
				node.transp = edge.transp
				node.ht = NULL_HT
				heap.Push(&hin, node)
			}
			as.H.Graph.Hins[currentId] = hin
		}

		as.openVertices[currentId] = false
		as.closedVertices[currentId] = true
		i++

		if currentId == arr {
			if start {
				return true
			} else {
				continue
			}
		}

		// take every GEN relationship departing from the node
		dbNode, _ := as.H.Db.Node(currentId)
		dbNode.Db = as.H.Db
		relsDb, _ := dbNode.Outgoing("GEN")

		for i := 0; i < len(relsDb); i++ {

			nextCityNode, _ := relsDb[i].End()
			nextCityId := nextCityNode.Id()
			var nextCity City
			nextCity, exists := as.H.Graph.Cities[nextCityId]
			if !exists {
				nextCity = City{Id: nextCityId}
				as.H.Graph.Cities[nextCityId] = nextCity
			}

			_, cycling := current.Prevs[nextCityId]
			if cycling /*|| currentTransfers == common.MAX_TRANSFERS*/ {
				continue
			}

			relsDb[i].Db = as.H.Db
			props, _ := relsDb[i].Properties()
			transpProp := props["transp"].(float64)
			transp := int(transpProp)
			weight, _ := props["price"].(float64)

			as.incomingEdges[nextCityId] = append(as.incomingEdges[nextCityId], EdgeGen{idNode: currentId, transp: transp, weight: weight})

			if as.closedVertices[nextCityId] {

				// still we add the new sidetrack edge
				var hin Hin
				for _, node := range as.H.Graph.Hins[nextCityId] {
					hin = append(hin, node)
				}
				if len(hin) == 0 {
					heap.Init(&hin)
				}
				node := Node{u: currentId, v: nextCityId, dValue: as.gScore[currentId] + weight - as.gScore[nextCityId], hin: nextCityId, transp: transp, ht: NULL_HT}
				pair := len(hin)%2 == 0
				heap.Push(&hin, node)

				tentativeGScore := as.gScore[currentId] + weight
				if tentativeGScore < as.gScore[nextCityId] {
					as.consistent = false
					// fmt.Println("not consistent")
				}

				// // check if node addition is interfering with the dijkstra execution
				add := true
				if !start {
					for k := range hin {
						node := hin[k]
						if node.u == currentId && node.v == nextCityId && node.transp == transp {
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
					as.H.Graph.Hins[nextCityId] = hin
				} else if start {
					as.H.Graph.Hins[nextCityId] = hin
				}

				continue
			}

			tentativeGScore := as.gScore[currentId] + weight

			if _, exists := as.gScore[nextCityId]; exists && tentativeGScore >= as.gScore[nextCityId] {
				continue
			}

			// update scores
			as.gScore[nextCityId] = tentativeGScore
			as.fScore[nextCityId] = tentativeGScore + as.getHeuristicValue(nextCityId, arr)

			nextCity.CameFrom = currentId
			nextCity.Transp = transp
			as.H.Graph.Cities[nextCityId] = nextCity

			newPrevs := make(map[int]bool)
			for k := range current.Prevs {
				newPrevs[k] = true
			}
			newPrevs[currentId] = true

			if !as.openVertices[nextCityId] {
				heap.Push(as, AstarNode{Id: nextCityId, Transfers: currentTransfers + 1, Prevs: newPrevs})
				as.openVertices[nextCityId] = true
			} else {
				// fix the heap
				for i := range as.pq {
					if as.pq[i].Id == nextCityId {
						as.pq[i] = AstarNode{Id: nextCityId, Transfers: currentTransfers + 1, Prevs: newPrevs}
						heap.Fix(as, i)
					}
				}
			}

		}

	}

	return false

}
