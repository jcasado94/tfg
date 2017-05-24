package scraping

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/jcasado94/tfg/common"
	"github.com/jmcvetta/neoism"
	// "time"
)

var mutex = &sync.Mutex{}
var mutex2 = &sync.Mutex{}

/*
	trips is a slice of common.Trip from a dep city to an arr city, strings corresponding to valid airport or city codes for the node properties depending on the type of transportation specified in transp.
	fullfils the database with the corresponding trips in the form of SPEC relationships .
*/
func retainSpecificTrips(trips []common.Trip, dep, arr string, transp int) {

	// t1 := time.Now()

	db, err := neoism.Connect(common.GetDBTransactionUrl())
	common.PanicErr(err)

	if len(trips) == 0 {
		return
	}

	var code string
	switch {

	case transp == common.TRANSP_AEROL || transp == common.TRANSP_LAN:

		code = "airpCode"

	case transp == common.TRANSP_BUS:

		code = "plat10id"

	}

	var query string
	MATCH := "MATCH (a:City {" + code + ":{dep}}), (b:City {" + code + ":{arr}})\n"

	var props = make(map[string]interface{})
	props["dep"] = dep
	props["arr"] = arr

	var writtenQueries = make(map[string]bool) // tells us whether a query (concatenation of the strings that populate it) has been written or not

	MERGE := ""

	for i, trip := range trips {

		depYear := fmt.Sprintf("depYear%d", i)
		depMonth := fmt.Sprintf("depMonth%d", i)
		depDay := fmt.Sprintf("depDay%d", i)
		depHour := fmt.Sprintf("depHour%d", i)
		depMin := fmt.Sprintf("depMin%d", i)

		arrYear := fmt.Sprintf("arrYear%d", i)
		arrMonth := fmt.Sprintf("arrMonth%d", i)
		arrDay := fmt.Sprintf("arrDay%d", i)
		arrHour := fmt.Sprintf("arrHour%d", i)
		arrMin := fmt.Sprintf("arrMin%d", i)

		price := fmt.Sprintf("price%d", i)
		transpStr := fmt.Sprintf("transp%d", i)

		queryString := strconv.Itoa(trip.DepYear) + strconv.Itoa(trip.DepMonth) + strconv.Itoa(trip.DepDay) + strconv.Itoa(trip.DepHour) + strconv.Itoa(trip.DepMin) +
			strconv.Itoa(trip.ArrYear) + strconv.Itoa(trip.ArrMonth) + strconv.Itoa(trip.ArrDay) + strconv.Itoa(trip.ArrHour) + strconv.Itoa(trip.ArrMin) + strconv.FormatFloat(trip.PricePerAdult, 'f', 0, 64) + strconv.Itoa(transp)

		if _, exists := writtenQueries[queryString]; exists {
			continue
		}

		// we can proceed with the pattern
		writtenQueries[queryString] = true

		MERGE = MERGE + " MERGE (a)-[:SPEC {depYear:{" + depYear + "}, depMonth:{" + depMonth + "}, depDay:{" + depDay + "}, depHour:{" + depHour + "}, depMin:{" + depMin +
			"}, arrYear:{" + arrYear + "}, arrMonth:{" + arrMonth + "}, arrDay:{" + arrDay + "}, arrHour:{" + arrHour + "}, arrMin:{" + arrMin +
			"}, price:{" + price + "}, transp:{" + transpStr + "} }]->(b) "

		props[depYear] = trip.DepYear
		props[depMonth] = trip.DepMonth
		props[depDay] = trip.DepDay
		props[depHour] = trip.DepHour
		props[depMin] = trip.DepMin

		props[arrYear] = trip.ArrYear
		props[arrMonth] = trip.ArrMonth
		props[arrDay] = trip.ArrDay
		props[arrHour] = trip.ArrHour
		props[arrMin] = trip.ArrMin

		props[price] = trip.PricePerAdult
		props[transpStr] = transp

	}

	query = MATCH + MERGE

	cq := neoism.CypherQuery{
		Statement:  query,
		Parameters: props,
	}

	mutex.Lock()
	err = db.Cypher(&cq)
	mutex.Unlock()
	common.PanicErr(err)

	// t2 := time.Now()
	// fmt.Println(t2.Sub(t1))

}

func retainGeneralTrips(trips []common.Trip, dep, arr string, transp int) {

	db, err := neoism.Connect(common.GetDBTransactionUrl())
	common.PanicErr(err)

	if len(trips) == 0 {
		return
	}

	sumPrices := 0.0
	n := 0

	for _, trip := range trips {

		n++
		sumPrices += trip.PricePerAdult

	}

	var code string
	switch {

	case transp == common.TRANSP_AEROL || transp == common.TRANSP_LAN:

		code = "airpCode"

	case transp == common.TRANSP_BUS:

		code = "plat10id"

	}

	cq := neoism.CypherQuery{
		Statement: `
			MATCH (a:City {` + code + `:{depCode}}), (b:City {` + code + `:{arrCode}})
			MERGE (a)-[r:GEN {transp:{transpOption}}]->(b) 
			ON MATCH SET r.price = (coalesce(r.price, 0)*coalesce(r.n, 0) + {totalPrice})/(coalesce(r.n, 0)+{trips}),
			r.n = coalesce(r.n, 0)+{trips} 
			ON CREATE SET r.price = {totalPrice}/{trips}, r.n = {trips}
			`,
		Parameters: neoism.Props{"depCode": dep, "arrCode": arr, "transpOption": transp, "totalPrice": sumPrices, "trips": n},
	}

	mutex2.Lock()
	err = db.Cypher(&cq)
	mutex2.Unlock()
	common.PanicErr(err)

}
