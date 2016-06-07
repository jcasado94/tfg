package CBR

import (
	"fmt"
	// "github.com/Professorq/dijkstra"
	// "github.com/jcasado94/tfg/averages"
	"github.com/jcasado94/tfg/common"
	"github.com/jcasado94/tfg/scraping"
	"github.com/jmcvetta/neoism"
	"net/http"
	// "runtime/debug"
	// "encoding/gob"
	// "os"
	"sort"
	"strconv"
	"sync"
	"time"
)

/*
	TYPES
*/

// sorting interfaces
type ByArrTime struct {
	trips    []common.Trip
	location *time.Location
}

type ByPriceAndArrTime struct {
	trips    []common.Trip
	location *time.Location
}

func (a ByArrTime) Len() int      { return len(a.trips) }
func (a ByArrTime) Swap(i, j int) { a.trips[i], a.trips[j] = a.trips[j], a.trips[i] }

// arrival time
func (a ByArrTime) Less(i, j int) bool {
	iDate := time.Date(a.trips[i].ArrYear, time.Month(a.trips[i].ArrMonth), a.trips[i].ArrDay, a.trips[i].ArrHour, a.trips[i].ArrMin, 0, 0, a.location)
	jDate := time.Date(a.trips[j].ArrYear, time.Month(a.trips[j].ArrMonth), a.trips[j].ArrDay, a.trips[j].ArrHour, a.trips[j].ArrMin, 0, 0, a.location)
	return jDate.Sub(iDate) > 0
}

func (a ByPriceAndArrTime) Len() int      { return len(a.trips) }
func (a ByPriceAndArrTime) Swap(i, j int) { a.trips[i], a.trips[j] = a.trips[j], a.trips[i] }

// first price, then arr time
func (a ByPriceAndArrTime) Less(i, j int) bool {
	if a.trips[i].TotalPrice != a.trips[j].TotalPrice {
		return a.trips[i].TotalPrice < a.trips[j].TotalPrice
	} else {
		iDate := time.Date(a.trips[i].ArrYear, time.Month(a.trips[i].ArrMonth), a.trips[i].ArrDay, a.trips[i].ArrHour, a.trips[i].ArrMin, 0, 0, a.location)
		jDate := time.Date(a.trips[j].ArrYear, time.Month(a.trips[j].ArrMonth), a.trips[j].ArrDay, a.trips[j].ArrHour, a.trips[j].ArrMin, 0, 0, a.location)
		return jDate.Sub(iDate) > 0
	}
}

type rel struct {
	transp                    int // dep, arr: the valid codes of the cities corresponding to the transp transportation method.
	dep, arr                  string
	depDay, depMonth, depyear int // only if specific relation
}

type tripsAndPos struct {
	Trips []common.Trip
	Pos   int
}

/*
	HANDLERS
*/

type SameDayCombinationsHandler struct {
	Graph    GraphSpec
	Kstar    KstarSpec
	Astar    AstarSpec
	Location *time.Location
	Db       *neoism.Database
}

func a(h SameDayCombinationsHandler, n int) {
	c := h.Graph.Rels[n]
	fmt.Print("(" + strconv.Itoa(c.Id) + ") ")
	if h.Graph.Rels[c.CameFrom].Id != START_ID {
		a(h, c.CameFrom)
	}
}

func (h SameDayCombinationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// t1 := time.Now()

	q := common.DecodeRequest(r)
	json := common.EncodeAnswer((&h).RetrieveSpecificSolutions(q.Year, q.Month, q.Day, q.DepID, q.ArrID, q.Adults, q.Children11, q.Children5, q.Infants))
	fmt.Fprintf(w, json)

	// debug.PrintStack()

	// t2 := time.Now()
	// fmt.Println(t2.Sub(t1))

	// debug.PrintStack()
}

type UsualCombinationsHandler struct {
	Graph GraphGen
	Kstar KstarGen
	Astar AstarGen
	Db    *neoism.Database
}

func (h UsualCombinationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// t1 := time.Now()

	q := common.DecodeRequest(r)
	json := common.EncodeAnswer((&h).RetrieveGeneralSolutions(q.Year, q.Month, q.Day, q.DepID, q.ArrID, q.Adults, q.Children11, q.Children5, q.Infants))
	fmt.Fprintf(w, json)

	// t2 := time.Now()
	// fmt.Println(t2.Sub(t1))

	// debug.PrintStack()

}

type DirectTripsAerolineasHandler struct{}

func (h DirectTripsAerolineasHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// t1 := time.Now()

	q := common.DecodeRequest(r)
	var res [][]common.Trip
	trips := DirectTrips(q.Year, q.Month, q.Day, q.Adults, q.Children11, q.Children5, q.Infants, q.DepID, q.ArrID, common.TRANSP_AEROL)
	for _, trip := range trips {
		res = append(res, []common.Trip{trip})
	}
	json := common.EncodeAnswer(res)
	fmt.Fprintf(w, json)

	// t2 := time.Now()
	// fmt.Println(t2.Sub(t1))

	// debug.PrintStack()

}

type DirectTripsLANHandler struct{}

func (h DirectTripsLANHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// t1 := time.Now()

	q := common.DecodeRequest(r)
	var res [][]common.Trip
	trips := DirectTrips(q.Year, q.Month, q.Day, q.Adults, q.Children11, q.Children5, q.Infants, q.DepID, q.ArrID, common.TRANSP_LAN)
	for _, trip := range trips {
		res = append(res, []common.Trip{trip})
	}
	json := common.EncodeAnswer(res)
	fmt.Fprintf(w, json)

	// t2 := time.Now()
	// fmt.Println(t2.Sub(t1))

	// debug.PrintStack()

}

type DirectTripsPlat10Handler struct{}

func (h DirectTripsPlat10Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// t1 := time.Now()

	q := common.DecodeRequest(r)
	var res [][]common.Trip
	trips := DirectTrips(q.Year, q.Month, q.Day, q.Adults, q.Children11, q.Children5, q.Infants, q.DepID, q.ArrID, common.TRANSP_BUS)
	for _, trip := range trips {
		res = append(res, []common.Trip{trip})
	}
	json := common.EncodeAnswer(res)
	fmt.Fprintf(w, json)

	// t2 := time.Now()
	// fmt.Println(t2.Sub(t1))

	// debug.PrintStack()

}

/*
	FUNCTIONS
*/

/*
	PUBLIC
*/

/*
	Not really a CBR function. Added here for simplicity.
	Gets the set of solutions directly taking them from the specified (which) scraper
*/
func DirectTrips(year, month, day, adults, children11, children5, infants, depId, arrId string, which int) []common.Trip {

	db, err := neoism.Connect(common.TRANSACTION_URL)
	common.PanicErr(err)

	intDepId, _ := strconv.Atoi(depId)
	intArrId, _ := strconv.Atoi(arrId)

	nodeA, _ := db.Node(intDepId)
	nodeA.Db = db
	propsA, _ := nodeA.Properties()
	nodeB, _ := db.Node(intArrId)
	nodeB.Db = db
	propsB, _ := nodeB.Properties()

	var dep, arr string
	if which == common.TRANSP_AEROL || which == common.TRANSP_LAN {
		dep = propsA["airpCode"].(string)
		arr = propsB["airpCode"].(string)
	} else if which == common.TRANSP_BUS {
		dep = propsA["plat10id"].(string)
		arr = propsB["plat10id"].(string)
	}

	var res = scraping.GetDayOffersAndRetain(year, month, day, adults, children11, children5, infants, dep, arr, which)

	var solsAverages [][]common.Trip
	for _, trip := range res {
		solsAverages = append(solsAverages, []common.Trip{trip})
	}
	// averages.RetainAverage(depId, arrId, solsAverages)

	return res

}

/*
	Gets the set of solutions checking the GEN (general) relationships between the departure and the arrival nodes.
*/
func (h *UsualCombinationsHandler) RetrieveGeneralSolutions(year, month, day, depId, arrId, adults, children11, children5, infants string) (sols [][]common.Trip) {

	/*
		FETCH PATHS
	*/

	intDepId, _ := strconv.Atoi(depId)
	intArrId, _ := strconv.Atoi(arrId)

	db, err := neoism.Connect(common.TRANSACTION_URL)
	common.PanicErr(err)

	h.Db = db
	h.Kstar.H = h

	paths := h.Kstar.GoKStar(intDepId, intArrId)

	// var a []float64
	// dataFile, _ := os.Open("genTimes.gob")
	// dataDecoder := gob.NewDecoder(dataFile)
	// _ = dataDecoder.Decode(&a)
	// dataFile.Close()
	// a = append(a, h.Astar.seconds)
	// dataFile, err = os.Create("genTimes.gob")
	// common.PanicErr(err)
	// dataEncoder := gob.NewEncoder(dataFile)
	// dataEncoder.Encode(a)
	// dataFile.Close()
	// var b int
	// dataFile, _ = os.Open("genInconsistent.gob")
	// dataDecoder = gob.NewDecoder(dataFile)
	// _ = dataDecoder.Decode(&b)
	// dataFile.Close()
	// if !h.Astar.consistent {
	// 	b++
	// }
	// dataFile, err = os.Create("genInconsistent.gob")
	// common.PanicErr(err)
	// dataEncoder = gob.NewEncoder(dataFile)
	// dataEncoder.Encode(b)
	// dataFile.Close()

	var relPaths [][]rel

	for _, path := range paths {

		var newRels []rel
		dep := intDepId // same as path[len-1]
		depNode, _ := db.Node(dep)
		depNode.Db = db
		depProps, _ := depNode.Properties()
		for i := len(path) - 2; i >= 0; i-- {
			arr := path[i][0]
			arrNode, _ := db.Node(arr)
			arrNode.Db = db
			arrProps, _ := arrNode.Properties()
			transp := path[i][1]
			var depCode, arrCode string
			switch {
			case transp == common.TRANSP_AEROL || transp == common.TRANSP_LAN:
				depCode = depProps["airpCode"].(string)
				arrCode = arrProps["airpCode"].(string)
			case transp == common.TRANSP_BUS:
				depCode = depProps["plat10id"].(string)
				arrCode = arrProps["plat10id"].(string)
			}
			newRel := rel{dep: depCode, arr: arrCode, transp: transp}
			newRels = append(newRels, newRel)
			dep = arr
			depProps = arrProps
		}
		relPaths = append(relPaths, newRels)

	}

	sols = checkPaths(year, month, day, adults, children11, children5, infants, relPaths, db, false)

	// averages.RetainAverage(depId, arrId, sols)

	// return sols
	return sols

}

func (h *SameDayCombinationsHandler) RetrieveSpecificSolutions(year, month, day, depId, arrId, adults, children11, children5, infants string) (sols [][]common.Trip) {

	h.Location, _ = time.LoadLocation("America/Buenos_Aires")

	/*
		FETCH ALL SPECS RELS WITH SAME OR OLDER DATE
	*/

	db, err := neoism.Connect(common.TRANSACTION_URL)
	common.PanicErr(err)

	intDepYear, _ := strconv.Atoi(year)
	intDepMonth, _ := strconv.Atoi(month)
	intDepDay, _ := strconv.Atoi(day)
	h.Kstar.DepartureTime = time.Date(intDepYear, time.Month(intDepMonth), intDepDay, 0, 0, 0, 0, h.Location)
	h.Kstar.yearsLookup = []int{intDepYear}
	h.Kstar.monthsLookup = []int{intDepMonth}
	h.Kstar.daysLookup = []int{intDepDay}
	lastDay := intDepDay

	for plusDays := 1; plusDays <= common.MAX_DAYS_SPEC; plusDays++ {
		nextTime := time.Date(intDepYear, time.Month(intDepMonth), intDepDay+plusDays, 0, 0, 0, 0, h.Location)
		h.Kstar.daysLookup = append(h.Kstar.daysLookup, nextTime.Day())
		if nextTime.Day() < lastDay {
			h.Kstar.difMonth = true
		}
		if int(nextTime.Month()) != intDepMonth {
			h.Kstar.monthsLookup = append(h.Kstar.monthsLookup, int(nextTime.Month()))
			if int(nextTime.Year()) != intDepYear {
				h.Kstar.yearsLookup = append(h.Kstar.yearsLookup, nextTime.Year())
			}
		}
	}

	intDepId, _ := strconv.Atoi(depId)
	intArrId, _ := strconv.Atoi(arrId)

	h.Db = db
	h.Kstar.H = h

	// t1 := time.Now()
	paths := h.Kstar.GoKStar(intDepId, intArrId)
	// fmt.Println(time.Now().Sub(t1))

	// var a []float64
	// dataFile, _ := os.Open("specTimes.gob")
	// dataDecoder := gob.NewDecoder(dataFile)
	// _ = dataDecoder.Decode(&a)
	// dataFile.Close()
	// a = append(a, h.Astar.seconds)
	// dataFile, err = os.Create("specTimes.gob")
	// common.PanicErr(err)
	// dataEncoder := gob.NewEncoder(dataFile)
	// dataEncoder.Encode(a)
	// dataFile.Close()
	// var b int
	// dataFile, _ = os.Open("specInconsistent.gob")
	// dataDecoder = gob.NewDecoder(dataFile)
	// _ = dataDecoder.Decode(&b)
	// dataFile.Close()
	// if !h.Astar.consistent {
	// 	b++
	// }
	// dataFile, err = os.Create("specInconsistent.gob")
	// common.PanicErr(err)
	// dataEncoder = gob.NewEncoder(dataFile)
	// dataEncoder.Encode(b)
	// dataFile.Close()

	var relPaths [][]rel

	for _, path := range paths {

		path := path[1 : len(path)-1]
		var relPath []rel

		for i := len(path) - 1; i >= 0; i-- {
			relId := path[i]
			nodeDep, _ := db.Node(h.Graph.Rels[relId].DepNode)
			nodeDep.Db = db
			nodeArr, _ := db.Node(h.Graph.Rels[relId].ArrNode)
			nodeArr.Db = db
			propsDep, _ := nodeDep.Properties()
			propsArr, _ := nodeArr.Properties()
			var dep, arr string
			transp := h.Graph.Rels[relId].Transp
			switch {
			case transp == common.TRANSP_AEROL || transp == common.TRANSP_LAN:
				dep = propsDep["airpCode"].(string)
				arr = propsArr["airpCode"].(string)

			case transp == common.TRANSP_BUS:
				dep = propsDep["plat10id"].(string)
				arr = propsArr["plat10id"].(string)
			}
			newRel := rel{dep: dep, arr: arr, transp: transp, depDay: h.Graph.Rels[relId].DepTime.Day(), depMonth: int(h.Graph.Rels[relId].DepTime.Month()), depyear: h.Graph.Rels[relId].DepTime.Year()}
			relPath = append(relPath, newRel)
		}

		relPaths = append(relPaths, relPath)

	}

	sols = checkPaths(year, month, day, adults, children11, children5, infants, relPaths, db, true)

	// averages.RetainAverage(depId, arrId, sols)

	return sols
}

/*
	PRIVATE
*/

// construct a set of solutions from the set of paths found in the database.
func checkPaths(year, month, day, adults, children11, children5, infants string, paths [][]rel, db *neoism.Database, specific bool) (sols [][]common.Trip) {

	/*
		CHECK PATHS WITH SCRAPERS
	*/

	solsChan := make(chan []common.Trip)
	wg := &sync.WaitGroup{}

	for i := range paths {

		path := paths[i]

		cPrice := make(chan tripsAndPos)
		cTime := make(chan tripsAndPos)
		cFirstRel := make(chan []common.Trip)
		nSorts := 0
		var tripsSortedByPrice = make([][]common.Trip, len(path))
		var tripsSortedByTime = make([][]common.Trip, len(path))

		idPath := i * 1000 // 1000 trip ids for every path

		wg.Add(1)
		go func() {

			defer wg.Done()

			for j := range path {

				rel := path[j]

				idRel := idPath + j*1000/len(path) // share ids equally between rels

				// how many trip slices will be passed
				if j != 0 {
					nSorts += 2
				} else {
					nSorts++
				}

				jj := j

				go func() {

					var trips []common.Trip

					if specific {
						trips = scraping.GetDayOffersAndRetain(strconv.Itoa(rel.depyear), strconv.Itoa(rel.depMonth), strconv.Itoa(rel.depDay), adults, children11, children5, infants, rel.dep, rel.arr, rel.transp)
					} else {
						if jj == 0 {
							// departing trips just for the departure day
							trips = scraping.GetDayOffersAndRetain(year, month, day, adults, children11, children5, infants, rel.dep, rel.arr, rel.transp)
						} else {
							// other trips departure day or one more
							yearInt, _ := strconv.Atoi(year)
							monthInt, _ := strconv.Atoi(month)
							dayInt, _ := strconv.Atoi(day)
							location, _ := time.LoadLocation("America/Buenos_Aires")

							daySols := make(chan []common.Trip)
							for i := 0; i < common.GEN_DAYS_SCOPE+1; i++ {
								ii := i
								go func() {
									date := time.Date(yearInt, time.Month(monthInt), dayInt+ii, 0, 0, 0, 0, location)
									daySols <- scraping.GetDayOffersAndRetain(strconv.Itoa(date.Year()), strconv.Itoa(int(date.Month())), strconv.Itoa(date.Day()), adults, children11, children5, infants, rel.dep, rel.arr, rel.transp)
								}()
							}

							for i := 0; i < common.GEN_DAYS_SCOPE+1; i++ {
								nextTrips := <-daySols
								trips = append(trips, nextTrips...)
							}
						}
					}

					for t := range trips {
						trips[t].Id = idRel
						idRel++
					}

					if jj == 0 {
						cFirstRel <- trips
					} else {
						// h.sortTrips(trips, cPrice, cTime, jj)
						sortTrips(trips, cPrice, cTime, jj)
					}

				}()
			}

			// wait for the sorted trips to arrive
			for i := 0; i < nSorts; i++ {

				select {

				case priceTrips := <-cPrice:

					tripsSortedByPrice[priceTrips.Pos] = priceTrips.Trips

				case timeTrips := <-cTime:

					tripsSortedByTime[timeTrips.Pos] = timeTrips.Trips

				case firstTrips := <-cFirstRel:

					tripsSortedByPrice[0] = firstTrips
					tripsSortedByTime[0] = firstTrips

				}

			}

			// create the solutions
			wg2 := &sync.WaitGroup{}
			startingTrips := tripsSortedByPrice[0]

			for _, stTrip := range startingTrips {

				startingPath := []common.Trip{stTrip}
				wg2.Add(2)
				// go h.findPath(startingPath, tripsSortedByPrice[1:], solsChan, wg2)
				go findPath(startingPath, tripsSortedByPrice[1:], solsChan, wg2)
				// go h.findPath(startingPath, tripsSortedByTime[1:], solsChan, wg2)
				go findPath(startingPath, tripsSortedByTime[1:], solsChan, wg2)

			}

			wg2.Wait()
		}()
	}

	// wait for the solutions to arrive
	go func() {
		wg.Wait()
		close(solsChan)
	}()

	//receive solutions
	for i := range solsChan {
		sols = append(sols, i)
	}

	// sols = h.removeEqualSols(sols)
	sols = removeEqualSols(sols)

	return sols

}

func sortTrips(trips []common.Trip, cPrice chan tripsAndPos, cTime chan tripsAndPos, jj int) {
	location, _ := time.LoadLocation("America/Buenos_Aires")
	sort.Sort(ByPriceAndArrTime{location: location, trips: trips})
	cPrice <- tripsAndPos{Trips: trips, Pos: jj}
	var trips2 = make([]common.Trip, len(trips))
	copy(trips2, trips)
	sort.Sort(ByArrTime{location: location, trips: trips2})
	cTime <- tripsAndPos{Trips: trips2, Pos: jj}
}

// given the sorted trips and the solution populated by one starting trips, returns through the channel the first valid path found.
func findPath(sol []common.Trip, sortedTrips [][]common.Trip, solsChan chan []common.Trip, wg *sync.WaitGroup) {
	location, _ := time.LoadLocation("America/Buenos_Aires")
	if len(sortedTrips) == 0 {
		defer wg.Done()
		solsChan <- sol
	} else {
		lastTrip := sol[len(sol)-1]
		arrYear := lastTrip.ArrYear
		arrMonth := lastTrip.ArrMonth
		arrDay := lastTrip.ArrDay
		arrHour := lastTrip.ArrHour
		arrMin := lastTrip.ArrMin
		nextTrips := sortedTrips[0]
		for _, nextTrip := range nextTrips {
			tArr := time.Date(arrYear, time.Month(arrMonth), arrDay, arrHour, arrMin, 0, 0, location)
			tDep := time.Date(nextTrip.DepYear, time.Month(nextTrip.DepMonth), nextTrip.DepDay, nextTrip.DepHour, nextTrip.DepMin, 0, 0, location)
			if tDep.Sub(tArr).Minutes() >= common.TRANSFER_TIME && tDep.Sub(tArr).Hours() <= common.MAX_TRANSFER_HOURS {
				sol = append(sol, nextTrip)
				//h.findPath(sol, sortedTrips[1:], solsChan, wg)
				findPath(sol, sortedTrips[1:], solsChan, wg)
				return
			}
		}
		// no possible combination
		defer wg.Done()
		return
	}
}

// given a set of solutions, removes the duplicated ones.
func /*(h *UsualCombinationsHandler)*/ removeEqualSols(sols [][]common.Trip) (res [][]common.Trip) {
	type Something struct {
		mapa     map[int]*Something
		happened bool
	}

	var st = Something{mapa: make(map[int]*Something), happened: false}
	for _, v := range sols {
		found := false
		st1 := st
		for i, w := range v {
			id := w.Id
			_, ok := st1.mapa[id]
			if ok {
				st1 = *st1.mapa[id]
				if i == len(v)-1 && st1.happened {
					found = true
				}
			} else {
				st1.mapa[id] = &Something{mapa: make(map[int]*Something), happened: i == len(v)-1}
				st1 = *(st1.mapa[id])
			}
		}
		if !found {
			res = append(res, v)
		}
	}

	return res
}
