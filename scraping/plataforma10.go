package scraping

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jcasado94/tfg/common"
	"github.com/jmcvetta/neoism"
	"golang.org/x/net/html"
	// "time"
)

/*
	pre: year, month, day represent a future date.
		orig and dest are valid ids for plat10 website.
		adults > 0, babies <= adults
	post: the trips of that day are returned. if children5+babies > adults, returns empty.

*/
func getDayOffersP10(year, month, day, adults, children11, children5, babies, orig, dest string) []common.Trip {

	children5N, _ := strconv.Atoi(children5)
	children11N, _ := strconv.Atoi(children11)
	adultsN, _ := strconv.Atoi(adults)
	babiesN, _ := strconv.Atoi(babies)

	if children5N+babiesN > adultsN {
		return []common.Trip{}
	}

	var ret []common.Trip

	db, err := neoism.Connect(common.GetDBTransactionUrl())
	common.PanicErr(err)

	// get the names of orig and dest
	res := []struct {
		OrigName string `json:"a.plat10name"`
		DestName string `json:"b.plat10name"`
	}{}

	cq1 := neoism.CypherQuery{
		Statement: `
			MATCH (a:City), (b:City)
			WHERE a.plat10id = {orig} AND b.plat10id = {dest}
			RETURN a.plat10name, b.plat10name
		`,
		Parameters: neoism.Props{"orig": orig, "dest": dest},
		Result:     &res,
	}
	db.Cypher(&cq1)

	tripUrl := "http://www.plataforma10.com/ar/Servicios#buscar/" + orig + "/" + dest + "/" + day + "-" + month + "-" + year

	fecha := day + "/" + month + "/" + year
	Url, err := url.Parse("http://www.plataforma10.com/ar/ServiciosApi/Buscar?")
	params := url.Values{}
	params.Add("FechaIda", fecha)
	params.Add("IdPadDestino", dest)
	params.Add("NombrePadDestino", res[0].DestName)
	params.Add("IdPadOrigen", orig)
	params.Add("NombrePadOrigen", res[0].OrigName)
	Url.RawQuery = params.Encode()

	myUrl := Url.String()

	req, _ := http.NewRequest("GET", myUrl, nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", "www.plataforma10")

	client := initializeClient()

	resp, err := client.Do(req)
	if resp == nil || err != nil {
		return []common.Trip{}
	}
	defer resp.Body.Close()

	var foundTrips = make(map[string]map[string]map[string]common.Trip) // [company][depDate][arrDate]price

	z := html.NewTokenizer(resp.Body)

	for {

		tt := z.Next()

		if tt == html.StartTagToken {

			t := z.Token()

			if len(t.Attr) > 3 {

				if t.Attr[3].Key == "data-orden-precio" {

					disp, _ := strconv.Atoi(t.Attr[4].Val)
					if adultsN+children11N > disp {
						continue
					}

					newPrice, _ := strconv.ParseFloat(t.Attr[3].Val, 64)

					// check if it's the cheapest
					cheapest := true
					if comp, exists := foundTrips[t.Attr[7].Val]; !exists {
						foundTrips[t.Attr[7].Val] = make(map[string]map[string]common.Trip)
						foundTrips[t.Attr[7].Val][t.Attr[8].Val] = make(map[string]common.Trip)
					} else {
						if depDate, exists := comp[t.Attr[8].Val]; !exists {
							foundTrips[t.Attr[7].Val][t.Attr[8].Val] = make(map[string]common.Trip)
						} else {
							if trip, exists := depDate[t.Attr[9].Val]; exists {
								price := trip.PricePerAdult
								if newPrice >= price {
									cheapest = false
								}
							}
						}
					}

					if !cheapest {
						continue
					}

					var trip common.Trip
					trip.PricePerAdult, _ = strconv.ParseFloat(t.Attr[3].Val, 64)
					trip.TotalPrice = trip.PricePerAdult * float64(adultsN+children11N)
					trip.DepAirp, trip.ArrAirp = res[0].OrigName, res[0].DestName

					dateHour := strings.Split(t.Attr[8].Val, " ")
					yearMonthDay := strings.Split(dateHour[0], "/")
					trip.DepYear, _ = strconv.Atoi(yearMonthDay[0])
					trip.DepMonth, _ = strconv.Atoi(yearMonthDay[1])
					trip.DepDay, _ = strconv.Atoi(yearMonthDay[2])
					hourMin := strings.Split(dateHour[1], ":")
					trip.DepHour, _ = strconv.Atoi(hourMin[0])
					trip.DepMin, _ = strconv.Atoi(hourMin[1])

					dateHour = strings.Split(t.Attr[9].Val, " ")
					yearMonthDay = strings.Split(dateHour[0], "/")
					trip.ArrYear, _ = strconv.Atoi(yearMonthDay[0])
					trip.ArrMonth, _ = strconv.Atoi(yearMonthDay[1])
					trip.ArrDay, _ = strconv.Atoi(yearMonthDay[2])
					hourMin = strings.Split(dateHour[1], ":")
					trip.ArrHour, _ = strconv.Atoi(hourMin[0])
					trip.ArrMin, _ = strconv.Atoi(hourMin[1])

					trip.FlightNumber = t.Attr[7].Val + " - " + t.Attr[11].Val

					trip.Url = tripUrl

					// ret = append(ret, trip)
					foundTrips[t.Attr[7].Val][t.Attr[8].Val][t.Attr[9].Val] = trip
				}
			}
		} else if tt == html.ErrorToken {
			break
		}
	}

	for _, comp := range foundTrips {
		for _, depDate := range comp {
			for _, trip := range depDate {
				ret = append(ret, trip)
			}
		}
	}

	return ret
}
