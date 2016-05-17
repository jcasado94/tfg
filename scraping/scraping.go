package scraping

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jcasado94/tfg/common"
	"github.com/jmcvetta/neoism"
	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	// "time"
)

const MAX_FLIGHTS_AEROLINEAS = 15

var mutex = &sync.Mutex{}
var mutex2 = &sync.Mutex{}

func GetDayOffersAndRetain(year, month, day, adults, children11, children5, babies, orig, dest string, option int) []common.Trip {
	res := GetDayOffers(year, month, day, adults, children11, children5, babies, orig, dest, option)
	go retainSpecificTrips(res, orig, dest, option)
	go retainGeneralTrips(res, orig, dest, option)
	return res
}

/*
	pre:
	year, month, day represent a future date.
	orig and dest are valid places for the specific webpage database.
	0 < adults
	babies (<= 2 years) <= adults
		option might be:
			0 - Aerolineas Argentinas
			1 - LAN
			2 - Plataforma10
	post: the trips of that day are returned
*/
func GetDayOffers(year, month, day, adults, children11, children5, babies, orig, dest string, option int) []common.Trip {
	switch {
	case option == common.TRANSP_AEROL:
		return getDayOffersAerolineas(year, month, day, adults, children11, children5, babies, orig, dest)
	case option == common.TRANSP_LAN:
		return getDayOffersLAN(year, month, day, adults, children11, children5, babies, orig, dest)
	case option == common.TRANSP_BUS:
		return getDayOffersP10(year, month, day, adults, children11, children5, babies, orig, dest)
	default:
		return []common.Trip{}
	}
}

//initializes an http client with a cookiejar (cookie holder)
func initializeClient() *http.Client {

	//create client with cookies
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}

	jar, err := cookiejar.New(&options)
	if err != nil {
		log.Fatal(err)
	}

	return &http.Client{Jar: jar}
}

/*
	pre: year, month, day represent a future date.
		orig and dest are valid places for Aerolineas website.
		adults > 0, babies <= adults
	post: the trips of that day are returned

*/
func getDayOffersAerolineas(year, month, day, adults, children11, children5, babies, orig, dest string) []common.Trip {

	adultsN, _ := strconv.Atoi(adults)
	children5N, _ := strconv.Atoi(children5)
	children11N, _ := strconv.Atoi(children11)
	childrenN := children5N + children11N
	babiesN, _ := strconv.Atoi(babies)
	if adultsN+childrenN+babiesN > 8 {
		return []common.Trip{}
	}

	client := initializeClient()

	//connect to first website after query
	myUrl1 := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html"
	form1 := url.Values{}
	form1.Set("name", "ADVSForm")
	form1.Set("id", "ADVSForm")
	form1.Add("pointOfSale", "AR")
	form1.Add("searchType", "CALENDAR")
	form1.Add("currency", "ARS")
	form1.Add("alternativeLandingPage", "true")
	form1.Add("journeySpan", "OW")
	form1.Add("origin", orig)
	form1.Add("destination", dest)
	form1.Add("departureDate", year+"-"+month+"-"+day)
	form1.Add("numAdults", adults)
	form1.Add("numChildren", strconv.Itoa(childrenN))
	form1.Add("numInfants", babies)
	form1.Add("cabin", "ALL")
	form1.Add("lang", "es_ES")

	r, _ := http.NewRequest("POST", myUrl1, bytes.NewBufferString(form1.Encode()))
	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(r)
	if err != nil {
		return []common.Trip{}
	}

	//connect to second website
	myUrl2 := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html?execution=e1s1"
	form2 := url.Values{}
	form2.Set("_eventId_next", "")
	r, _ = http.NewRequest("POST", myUrl2, bytes.NewBufferString(form2.Encode()))
	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(r)
	if err != nil {
		return []common.Trip{}
	}

	//get the results sorted
	form3 := url.Values{}
	form3.Add("_eventId_ajax", "")
	form3.Add("execution", "e1s2")
	form3.Add("ajaxSource", "true")
	form3.Add("contextObject", `{"transferObjects":[{"componentType":"flc","actionCode":"sortFlights","queryData":{"componentId":"flc_1","componentType":"flc","actionCode":"sortFlights","queryData":null,"direction":"outbounds","flightIndex":0,"sortOption":"lowestprice","requestPartials":["__oneway"],"basketHashRefs":null}}]}`)
	r, _ = http.NewRequest("POST", myUrl1, bytes.NewBufferString(form3.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, err = client.Do(r)
	if err != nil {
		return []common.Trip{}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []common.Trip{}
	}
	// fmt.Println(string(body))

	// defer resp.Body.Close()
	var JSON interface{}
	json.Unmarshal(body, &JSON)

	//check if there are results
	if JSON == nil {
		return []common.Trip{}
	}

	JSONmap := JSON.(map[string]interface{})

	var result []common.Trip

	if JSONmap["content"] == nil {
		return []common.Trip{}
	}

	content := JSONmap["content"].([]interface{})

	if content[0] == nil {
		return []common.Trip{}
	}

	content0 := content[0].(map[string]interface{})

	if content0["model"] == nil {
		return []common.Trip{}
	}

	outbounds := content0["model"].(map[string]interface{})["outbounds"].([]interface{})

	for _, v := range outbounds {

		t := v.(map[string]interface{})

		segments := t["segments"].([]interface{})

		if len(segments) == 1 { // direct trip

			var trip common.Trip

			trip.Url = myUrl1
			trip.UrlParams = form1

			segment := segments[0].(map[string]interface{})

			trip.FlightNumber = "AR" + strconv.FormatInt(int64(segment["flightNumber"].([]interface{})[0].(float64)), 10)

			trip.DepAirp = segment["departureCode"].(string)
			trip.ArrAirp = segment["arrivalCode"].(string)

			date := strings.Split(segment["departureDate"].(string), " ")
			yearmonthday := strings.Split(date[0], "/")
			trip.DepYear, _ = strconv.Atoi(yearmonthday[0])
			trip.DepMonth, _ = strconv.Atoi(yearmonthday[1])
			trip.DepDay, _ = strconv.Atoi(yearmonthday[2])
			hour := strings.Split(date[1], ":")
			trip.DepHour, _ = strconv.Atoi(hour[0])
			trip.DepMin, _ = strconv.Atoi(hour[1])

			date = strings.Split(segment["arrivalDate"].(string), " ")
			yearmonthday = strings.Split(date[0], "/")
			trip.ArrYear, _ = strconv.Atoi(yearmonthday[0])
			trip.ArrMonth, _ = strconv.Atoi(yearmonthday[1])
			trip.ArrDay, _ = strconv.Atoi(yearmonthday[2])
			hour = strings.Split(date[1], ":")
			trip.ArrHour, _ = strconv.Atoi(hour[0])
			trip.ArrMin, _ = strconv.Atoi(hour[1])

			fares := t["basketsRef"].(map[string]interface{})

			fareInterface, hasFare := fares["PO"]
			fareName := "PO"
			if !hasFare {
				fareInterface, hasFare = fares["EC"]
				fareName = "EC"
				if !hasFare {
					fareInterface, hasFare = fares["FX"]
					fareName = "FX"
					if !hasFare {
						fareInterface = fares["CE"]
						fareName = "CE"
					}
				}
			}

			fare := fareInterface.(map[string]interface{})
			prices := fare["prices"].(map[string]interface{})

			trip.PricePerAdult, _ = strconv.ParseFloat(prices["priceAlternatives"].([]interface{})[0].(map[string]interface{})["pricesPerCurrency"].(map[string]interface{})["ARS"].(map[string]interface{})["amount"].(string), 64)

			if childrenN == 0 {
				trip.TotalPrice = trip.PricePerAdult * float64(adultsN)
			} else {
				moneyElements := prices["moneyElements"].([]interface{})
				var otherTaxes []float64
				var adultPrice float64
				XR, _ := strconv.ParseFloat(moneyElements[0].(map[string]interface{})["moneyTO"].(map[string]interface{})["amount"].(string), 64)
				TQ, _ := strconv.ParseFloat(moneyElements[1].(map[string]interface{})["moneyTO"].(map[string]interface{})["amount"].(string), 64)
				for i := 2; i < len(moneyElements); i++ {
					if i == len(moneyElements)-1 {
						adultPrice, _ = strconv.ParseFloat(moneyElements[i].(map[string]interface{})["moneyTO"].(map[string]interface{})["amount"].(string), 64)
					} else {
						tax, _ := strconv.ParseFloat(moneyElements[i].(map[string]interface{})["moneyTO"].(map[string]interface{})["amount"].(string), 64)
						otherTaxes = append(otherTaxes, tax)
					}
				}

				var factor float64

				if fareName == "PO" || fareName == "EC" {
					factor = float64(adultsN) + float64(childrenN)*0.8
				} else {
					factor = float64(adultsN) + float64(childrenN)*0.67
				}

				sumOtherTaxes := 0.0
				for _, tax := range otherTaxes {
					sumOtherTaxes += tax
				}
				trip.TotalPrice = (adultPrice+sumOtherTaxes)*factor + (XR+TQ)*float64(adultsN+childrenN)

			}

			result = append(result, trip)

		}

	}

	// t2 := time.Now()
	// fmt.Println(t2.Sub(t1))

	return result
}

/*
	pre: year, month, day represent a future date.
		orig and dest are valid places for Aerolineas website.
		adults > 0, babies <= adults
	post: the trips of that day are returned

*/
func getDayOffersLAN(year, month, day, adults, children11, children5, babies, orig, dest string) []common.Trip {

	adultsN, _ := strconv.Atoi(adults)
	children5N, _ := strconv.Atoi(children5)
	children11N, _ := strconv.Atoi(children11)
	childrenN := children11N + children5N
	if adultsN+childrenN > 7 {
		return []common.Trip{}
	}

	var result []common.Trip

	client := initializeClient()

	myUrl := "http://booking.lan.com/ws/booking/quoting/fares_availability/5.0/rest/get_availability"
	form := url.Values{}
	form.Add("adults", adults)
	form.Add("application", "compra_normal")
	form.Add("cabin", "Y")
	form.Add("children", strconv.Itoa(childrenN))
	form.Add("country", "AR")
	form.Add("departureDate", year+"-"+month+"-"+day)
	form.Add("destination", dest)
	form.Add("infants", babies)
	form.Add("language", "ES")
	form.Add("origin", orig)
	form.Add("portal", "personas")
	form.Add("roundTrip", "false")
	form.Add("section", "step2")

	var jsonStr = []byte(`{"language":"ES","country":"AR","portal":"personas","application":"compra_normal","section":"step2","cabin":"Y","adults":` + adults + `,"children":` + strconv.Itoa(childrenN) + `,"infants":` + babies + `,"roundTrip":false,"departureDate":"` + year + "-" + month + "-" + day + `","origin":"` + orig + `","destination":"` + dest + `"}`)
	req, _ := http.NewRequest("POST", myUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return []common.Trip{}
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var JSON map[string]interface{}
	json.Unmarshal(body, &JSON)
	//check if there are results
	if JSON["data"] == nil {
		return []common.Trip{}
	}

	//get routes from JSON
	routes := JSON["data"].(map[string]interface{})["itinerary"].(map[string]interface{})["routesMap"].(map[string]interface{})

	//itarate the routes
	c := make(chan common.Trip)
	nRoutes := 0
	for k, v := range routes {
		flightNumber := k
		flight := v.(map[string]interface{})
		segments := flight["segments"].([]interface{})
		// ignore if there is a transfer
		if len(segments) > 1 {
			continue
		}
		nRoutes++
		go func() {
			var trip common.Trip
			trip.FlightNumber = flightNumber
			trip.Url = `http://booking.lan.com/es_ar/apps/personas/compra?fecha1_dia=` + day + `&fecha1_anomes=` + year + "-" + month + `&auAvailability=1&ida_vuelta=ida&from_city1=` + orig + `&to_city1=` + dest + `&flex=1&cabina=Y&nadults=` + adults + `&nchildren=` + strconv.Itoa(childrenN) + `&ninfants=` + babies
			trip.UrlParams = nil
			travel := flight["travel"].(map[string]interface{})
			origin := travel["origin"].(map[string]interface{})
			destination := travel["destination"].(map[string]interface{})

			//departure info
			trip.DepAirp = origin["airport"].(map[string]interface{})["code"].(string)
			date := origin["date"].(string)
			hour := strings.Split(date, "T")
			yearmonthday := strings.Split(hour[0], "-")
			trip.DepYear, _ = strconv.Atoi(yearmonthday[0])
			trip.DepMonth, _ = strconv.Atoi(yearmonthday[1])
			trip.DepDay, _ = strconv.Atoi(yearmonthday[2])
			hourMin := strings.Split(hour[1], ":")
			trip.DepHour, _ = strconv.Atoi(hourMin[0])
			trip.DepMin, _ = strconv.Atoi(hourMin[1])

			//arrival info
			trip.ArrAirp = destination["airport"].(map[string]interface{})["code"].(string)
			date = destination["date"].(string)
			hour = strings.Split(date, "T")
			yearmonthday = strings.Split(hour[0], "-")
			trip.ArrYear, _ = strconv.Atoi(yearmonthday[0])
			trip.ArrMonth, _ = strconv.Atoi(yearmonthday[1])
			trip.ArrDay, _ = strconv.Atoi(yearmonthday[2])
			hourMin = strings.Split(hour[1], ":")
			trip.ArrHour, _ = strconv.Atoi(hourMin[0])
			trip.ArrMin, _ = strconv.Atoi(hourMin[1])

			//price
			var fare map[string]interface{}
			fares := flight["fareFamilyMap"].(map[string]interface{})
			SP, SPok := fares["SP"] //cheapest
			if SPok && SP.(map[string]interface{})["availability"].(float64) > 0.0 {
				fare = SP.(map[string]interface{})
			} else {
				LE, LEok := fares["LE"] //second cheapest
				if LEok && LE.(map[string]interface{})["availability"].(float64) > 0.0 {
					fare = LE.(map[string]interface{})
				} else {
					FX, FXok := fares["FX"] //third cheapest
					if FXok && FX.(map[string]interface{})["availability"].(float64) > 0.0 {
						fare = FX.(map[string]interface{})
					} else {
						fare = fares["FF"].(map[string]interface{}) // there for sure
					}
				}
			}

			passengerMap := fare["fare"].(map[string]interface{})["passengerMap"].(map[string]interface{})
			adultPrices := passengerMap["adult"].(map[string]interface{})
			adultPrice := adultPrices["amount"].(float64) + adultPrices["fee"].(float64) + adultPrices["tax"].(float64)
			var childrenPrice float64
			if childrenN > 0 {
				childrenPrices := passengerMap["child"].(map[string]interface{})
				childrenPrice = childrenPrices["amount"].(float64) + childrenPrices["fee"].(float64) + childrenPrices["tax"].(float64)
			}

			trip.PricePerAdult = adultPrice
			trip.TotalPrice = float64(adultsN)*adultPrice + float64(childrenN)*childrenPrice

			c <- trip

		}()

	}

	//wait for the routes to finish
	for i := 0; i < nRoutes; i++ {
		result = append(result, <-c)
	}

	return result

}

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

	db, err := neoism.Connect(common.TRANSACTION_URL)
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

	resp, _ := client.Do(req)
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

/*
	trips is a slice of common.Trip from a dep city to an arr city, strings corresponding to valid airport or city codes for the node properties depending on the type of transportation specified in transp.
	fullfils the database with the corresponding trips in the form of SPEC relationships .
*/
func retainSpecificTrips(trips []common.Trip, dep, arr string, transp int) {

	// t1 := time.Now()

	db, err := neoism.Connect(common.TRANSACTION_URL)
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

	db, err := neoism.Connect(common.TRANSACTION_URL)
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
			`,
		Parameters: neoism.Props{"depCode": dep, "arrCode": arr, "transpOption": transp, "totalPrice": sumPrices, "trips": n},
	}

	mutex2.Lock()
	err = db.Cypher(&cq)
	mutex2.Unlock()
	common.PanicErr(err)

}

/*****
OLD
*/

/*
	gets all the available flights with its cheapest available price.
*/
// func getFlightsAerolineasChildren(i int, client *http.Client, c chan common.Trip, year, month, day, adults, children, babies, orig, dest string) {

// 	//connect to first website after query
// 	myUrl1 := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html"
// 	form1 := url.Values{}
// 	form1.Set("name", "ADVSForm")
// 	form1.Set("id", "ADVSForm")
// 	form1.Add("pointOfSale", "AR")
// 	form1.Add("searchType", "CALENDAR")
// 	form1.Add("currency", "ARS")
// 	form1.Add("alternativeLandingPage", "true")
// 	form1.Add("journeySpan", "OW")
// 	form1.Add("origin", orig)
// 	form1.Add("destination", dest)
// 	form1.Add("departureDate", year+"-"+month+"-"+day)
// 	form1.Add("numAdults", adults)
// 	form1.Add("numChildren", children)
// 	form1.Add("numInfants", babies)
// 	form1.Add("cabin", "ALL")
// 	form1.Add("lang", "es_ES")

// 	r, _ := http.NewRequest("POST", myUrl1, bytes.NewBufferString(form1.Encode()))
// 	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
// 	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	resp, _ := client.Do(r)

// 	//connect to second website
// 	myUrl2 := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html?execution=e1s1"
// 	form2 := url.Values{}
// 	form2.Set("_eventId_next", "")
// 	r, _ = http.NewRequest("POST", myUrl2, bytes.NewBufferString(form2.Encode()))
// 	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
// 	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	resp, _ = client.Do(r)

// 	defer resp.Body.Close()
// 	z := html.NewTokenizer(resp.Body)

// 	it := 0

// 	for {
// 		//take all the ids
// 		tt := z.Next()

// 		switch {

// 		case tt == html.TextToken:

// 			direct := string(z.Raw()) == orig

// 			if direct {
// 				z.Next()
// 				z.Next()
// 				z.Next()
// 				z.Next()
// 				z.Next()
// 				z.Next()
// 				direct = string(z.Raw()) == dest
// 			}

// 			if direct {

// 				//look for flight id, should be in next "input" tag

// 				if i == it {

// 					tripFound := false

// 					for !tripFound {

// 						tt = z.Next()

// 						switch {

// 						case tt == html.SelfClosingTagToken:

// 							t := z.Token()

// 							if t.Data == "input" {
// 								tripFound = true
// 								id := strings.Split(t.Attr[0].Val, "_")[2]
// 								c <- getFlightAerolinias(id, myUrl1, &form1, client)
// 								return
// 							}

// 						}
// 					}
// 				}

// 				it++

// 			} else {
// 				c <- common.Trip{TotalPrice: -1.0}
// 				return
// 			}

// 		case tt == html.ErrorToken:
// 			c <- common.Trip{TotalPrice: -1.0}
// 			return
// 		}

// 	}

// }

// func getFlightAerolinias(id string, postUrl string, urlParams *url.Values, client *http.Client) common.Trip {

// 	var newTrip common.Trip
// 	newTrip.Url = postUrl
// 	newTrip.UrlParams = *urlParams

// 	//get json with info about flight
// 	myUrl := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html"
// 	form := url.Values{}
// 	form.Set("_eventId_ajax", "")
// 	form.Set("ajaxSource", "true")
// 	form.Set("contextObject", `{"transferObjects":[{"componentType":"cart","actionCode":"checkPrice","queryData":{"componentId":"cart_1","componentType":"cart","actionCode":"checkPrice","queryData":null,"requestPartials":["initialized"],"selectedBasketRefs":[`+id+`]}}]}`)
// 	form.Set("execution", "e1s2")

// 	r, _ := http.NewRequest("POST", myUrl, bytes.NewBufferString(form.Encode()))
// 	r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

// 	resp, _ := client.Do(r)

// 	body, _ := ioutil.ReadAll(resp.Body)
// 	// ioutil.WriteFile("megajson.html", body, 0644)
// 	resp.Body.Close()
// 	var JSON map[string]interface{}
// 	json.Unmarshal(body, &JSON)

// 	model := JSON["content"].([]interface{})[0].(map[string]interface{})["model"].(map[string]interface{})

// 	//total price
// 	totalPrice := model["amountDuePrices"].(map[string]interface{})["priceAlternatives"].([]interface{})[0].(map[string]interface{})["pricesPerCurrency"].(map[string]interface{})["ARS"].(map[string]interface{})["amount"].(string)
// 	newTrip.TotalPrice, _ = strconv.ParseFloat(totalPrice, 64)

// 	//other info
// 	itineraryParts := model["itineraryParts"].([]interface{})[0].(map[string]interface{})
// 	segments := itineraryParts["segments"].([]interface{})[0].(map[string]interface{})
// 	flightNumber := strconv.FormatFloat(segments["flightNumber"].([]interface{})[0].(float64), 'f', 0, 64)
// 	newTrip.FlightNumber = "AR" + flightNumber
// 	newTrip.DepAirp = segments["departureCode"].(string)
// 	newTrip.ArrAirp = segments["arrivalCode"].(string)
// 	date := strings.Split(segments["departureDate"].(string), " ")
// 	yearmonthday := strings.Split(date[0], "/")
// 	newTrip.DepYear, _ = strconv.Atoi(yearmonthday[0])
// 	newTrip.DepMonth, _ = strconv.Atoi(yearmonthday[1])
// 	newTrip.DepDay, _ = strconv.Atoi(yearmonthday[2])
// 	hour := strings.Split(date[1], ":")
// 	newTrip.DepHour, _ = strconv.Atoi(hour[0])
// 	newTrip.DepMin, _ = strconv.Atoi(hour[1])
// 	date = strings.Split(segments["arrivalDate"].(string), " ")
// 	yearmonthday = strings.Split(date[0], "/")
// 	newTrip.ArrYear, _ = strconv.Atoi(yearmonthday[0])
// 	newTrip.ArrMonth, _ = strconv.Atoi(yearmonthday[1])
// 	newTrip.ArrDay, _ = strconv.Atoi(yearmonthday[2])
// 	hour = strings.Split(date[1], ":")
// 	newTrip.ArrHour, _ = strconv.Atoi(hour[0])
// 	newTrip.ArrMin, _ = strconv.Atoi(hour[1])

// 	//adult price
// 	newTrip.PricePerAdult, _ = strconv.ParseFloat(itineraryParts["prices"].(map[string]interface{})["priceAlternatives"].([]interface{})[0].(map[string]interface{})["pricesPerCurrency"].(map[string]interface{})["ARS"].(map[string]interface{})["amount"].(string), 64)

// 	return newTrip
// }

// /*
//  	pre: Children = 0.
// 	gets all the available flights with its cheapest available price.
// */
// func getFlightsAerolineasAdults(year, month, day, adults, children, babies, orig, dest string) []common.Trip {

// 	client := initializeClient()
// 	adultsN, _ := strconv.Atoi(adults)
// 	nextDay := 0

// 	//connect to first website after query
// 	myUrl1 := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html"
// 	form1 := url.Values{}
// 	form1.Set("name", "ADVSForm")
// 	form1.Set("id", "ADVSForm")
// 	form1.Add("pointOfSale", "AR")
// 	form1.Add("searchType", "CALENDAR")
// 	form1.Add("currency", "ARS")
// 	form1.Add("alternativeLandingPage", "true")
// 	form1.Add("journeySpan", "OW")
// 	form1.Add("origin", orig)
// 	form1.Add("destination", dest)
// 	form1.Add("departureDate", year+"-"+month+"-"+day)
// 	form1.Add("numAdults", adults)
// 	form1.Add("numChildren", children)
// 	form1.Add("numInfants", babies)
// 	form1.Add("cabin", "ALL")
// 	form1.Add("lang", "es_ES")

// 	r, _ := http.NewRequest("POST", myUrl1, bytes.NewBufferString(form1.Encode()))
// 	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
// 	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	resp, _ := client.Do(r)

// 	//connect to second website
// 	myUrl2 := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html?execution=e1s1"
// 	form2 := url.Values{}
// 	form2.Set("_eventId_next", "")
// 	r, _ = http.NewRequest("POST", myUrl2, bytes.NewBufferString(form2.Encode()))
// 	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
// 	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	resp, _ = client.Do(r)

// 	defer resp.Body.Close()

// 	z := html.NewTokenizer(resp.Body)
// 	result := []common.Trip{}
// 	end := false

// 	for end == false {
// 		tt := z.Next()
// 		switch {

// 		case tt == html.StartTagToken:

// 			t := z.Token()
// 			if t.Data == "span" && len(t.Attr) > 0 && t.Attr[0].Val == "airport_code" {
// 				var trip common.Trip

// 				//departure airport
// 				z.Next() //should be airport code
// 				trip.DepAirp = string(z.Raw())

// 				//departure time
// 				z.Next()
// 				z.Next()
// 				z.Next() //should be dep time
// 				divided := strings.Split(string(z.Raw()), ":")
// 				trip.DepHour, _ = strconv.Atoi(divided[0])
// 				trip.DepMin, _ = strconv.Atoi(divided[1])
// 				trip.DepYear, _ = strconv.Atoi(year)
// 				trip.DepMonth, _ = strconv.Atoi(month)
// 				trip.DepDay, _ = strconv.Atoi(day)

// 				found := false
// 				for found == false {
// 					tt = z.Next() // move till the arrival time
// 					switch {
// 					case tt == html.StartTagToken:
// 						t = z.Token()
// 						if t.Data == "span" && len(t.Attr) > 0 && t.Attr[0].Val == "airport_code" {

// 							//arrival airport
// 							z.Next()
// 							trip.ArrAirp = string(z.Raw())

// 							//arrival time
// 							z.Next()
// 							z.Next()
// 							z.Next()
// 							divided = strings.Split(string(z.Raw()), ":")
// 							trip.ArrHour, _ = strconv.Atoi(divided[0])
// 							trip.ArrMin, _ = strconv.Atoi(divided[1])
// 							if trip.ArrHour < trip.DepHour { // arrives the next day
// 								nextDay++
// 							}
// 							trip.ArrYear = trip.DepYear
// 							trip.ArrMonth = trip.DepMonth
// 							trip.ArrDay = trip.DepDay + nextDay // if year or month changes, golang will correct it in the Date creation, when needed

// 							found = true
// 						}
// 					default:
// 						continue
// 					}
// 				}
// 				// check the flight's number
// 				found = false
// 				for found == false {
// 					tt = z.Next()
// 					switch {
// 					case tt == html.StartTagToken:
// 						t = z.Token()
// 						if t.Data == "a" { //it's the next link in the html
// 							z.Next() //should be the number
// 							trip.FlightNumber = strings.Replace(strings.TrimSpace(string(z.Raw())), " ", "", 1)
// 							found = true
// 						}
// 					default:
// 						continue
// 					}
// 				}

// 				// check if it's a direct flight
// 				found = false
// 				for found == false {
// 					tt = z.Next()
// 					switch {
// 					case tt == html.StartTagToken:
// 						t = z.Token()
// 						if t.Data == "span" && len(t.Attr) > 0 && t.Attr[0].Val == "translate stops wasTranslated" {

// 							z.Next()

// 							direct := string(z.Raw()) == orig

// 							if direct {
// 								z.Next()
// 								z.Next()
// 								z.Next()
// 								z.Next()
// 								z.Next()
// 								z.Next()
// 								direct = string(z.Raw()) == dest
// 							}

// 							if direct {
// 								//look for flight's cheapest price
// 								found2 := false
// 								for found2 == false {
// 									tt = z.Next()
// 									switch {
// 									case tt == html.StartTagToken:
// 										t = z.Token()
// 										if t.Data == "span" && len(t.Attr) > 0 && t.Attr[0].Val == "prices-amount" {
// 											z.Next() //should be the price
// 											float, _ := strconv.ParseFloat(string(z.Raw()), 64)
// 											trip.PricePerAdult = float
// 											trip.TotalPrice = float * float64(adultsN)
// 											trip.Url = myUrl1
// 											trip.UrlParams = form1
// 											result = append(result, trip)
// 											found = true
// 											found2 = true

// 											//check the total price
// 											tt = z.Next()

// 										}

// 									default:
// 										continue
// 									}
// 								}
// 							} else {
// 								end = true
// 								found = true
// 							}
// 						}

// 					default:
// 						continue
// 					}
// 				}
// 			}

// 		case tt == html.ErrorToken:
// 			end = true

// 		default:
// 			continue

// 		}
// 	}
// 	return result
// }
