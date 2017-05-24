package scraping

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/jcasado94/tfg/common"
	// "time"
)

/*
	pre: year, month, day represent a future date.
		orig and dest are valid places for LAN website.
		adults > 0, babies <= adults
	post: the trips of that day are returned

*/
func getDayOffersLAN(year, month, day, adults, children11, children5, babies, orig, dest string) []common.Trip {

	monthN, _ := strconv.Atoi(month)
	dayN, _ := strconv.Atoi(day)
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

	if monthN < 10 && string(month[0]) != "0" {
		month = "0" + month
	}
	if dayN < 10 && string(day[0]) != "0" {
		day = "0" + day
	}
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
