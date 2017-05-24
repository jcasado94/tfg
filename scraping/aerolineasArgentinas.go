package scraping

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jcasado94/tfg/common"
	// "time"
)

const MAX_FLIGHTS_AEROLINEAS = 15

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

	monthN, _ := strconv.Atoi(month)
	dayN, _ := strconv.Atoi(day)
	if monthN < 10 && string(month[0]) != "0" {
		month = "0" + month
	}
	if dayN < 10 && string(day[0]) != "0" {
		day = "0" + day
	}

	if childrenN == 0 {
		return getDayOffersAlmundo(year, month, day, adults, children11, children5, babies, orig, dest)
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
