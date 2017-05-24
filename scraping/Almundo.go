package scraping

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jcasado94/tfg/common"
	// "time"
)

func getDayOffersAlmundo(year, month, day, adults, children11, children5, babies, orig, dest string) []common.Trip {
	// intYear, _ := strconv.Atoi(year)
	intMonth, _ := strconv.Atoi(month)
	intDay, _ := strconv.Atoi(day)
	intYear, _ := strconv.Atoi(year)
	intChildren11, _ := strconv.Atoi(children11)
	intChildren5, _ := strconv.Atoi(children5)
	intAdults, _ := strconv.Atoi(adults)
	floatAdults := float64(intAdults)
	children := strconv.Itoa(intChildren11 + intChildren5)

	//URL & URL PARAMS

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
	form1.Add("numChildren", children)
	form1.Add("numInfants", babies)
	form1.Add("cabin", "ALL")
	form1.Add("lang", "es_ES")

	url := "https://almundo.com.ar/flights/async/itineraries?adults=" + adults + "&children=" + children + "&date=" + year + "-" + month + "-" + day + "&from=" + orig + "&infants=" + babies + "&stops=0&to=" + dest

	req, _ := http.NewRequest("GET", url, nil)
	client := initializeClient()
	resp, err := client.Do(req)
	if err != nil {
		return []common.Trip{}
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var JSON map[string]interface{}
	json.Unmarshal(body, &JSON)

	if JSON["results"] == nil {
		return []common.Trip{}
	}
	clusters := JSON["results"].(map[string]interface{})["clusters"].([]interface{})
	var trips []common.Trip

	for _, c := range clusters {

		var newTrip common.Trip
		newTrip.DepYear, newTrip.DepMonth, newTrip.DepDay = intYear, intMonth, intDay
		newTrip.Url, newTrip.UrlParams = myUrl1, form1

		cluster := c.(map[string]interface{})
		trip := cluster["segments"].([]interface{})[0].(map[string]interface{})["choices"].([]interface{})[0].(map[string]interface{})
		price := cluster["price"].(map[string]interface{})

		depTime := trip["departure_time"].(string)
		times := strings.Split(depTime, ":")
		newTrip.DepHour, _ = strconv.Atoi(times[0])
		newTrip.DepMin, _ = strconv.Atoi(times[1])

		arrDate := trip["arrival_date"].(map[string]interface{})
		date := strings.Split(arrDate["plain"].(string), "-")
		newTrip.ArrYear, _ = strconv.Atoi(date[0])
		newTrip.ArrMonth, _ = strconv.Atoi(date[1])
		newTrip.ArrDay, _ = strconv.Atoi(date[2])
		arrTime := trip["arrival_time"].(string)
		times = strings.Split(arrTime, ":")
		newTrip.ArrHour, _ = strconv.Atoi(times[0])
		newTrip.ArrMin, _ = strconv.Atoi(times[1])

		leg := trip["legs"].([]interface{})[0].(map[string]interface{})
		carrier := leg["marketing_carrier"].(map[string]interface{})["code"].(string)
		if carrier != "AR" {
			continue
		}
		number := int(leg["number"].(float64))
		flight := carrier + strconv.Itoa(number)
		newTrip.FlightNumber = flight

		newTrip.DepAirp = leg["origin"].(map[string]interface{})["code"].(string)
		newTrip.ArrAirp = leg["destination"].(map[string]interface{})["code"].(string)

		total := price["total"].(float64)
		detail := price["detail"].(map[string]interface{})
		adults := detail["adults"].(float64)
		taxes := detail["taxes"].(float64)
		tax := taxes / floatAdults
		newTrip.PricePerAdult = adults + tax
		fee := detail["fee"].(float64)
		newTrip.TotalPrice = total - fee

		trips = append(trips, newTrip)

	}

	return trips

}
