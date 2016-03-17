package scraping

import (
	"bytes"
	"encoding/json"
	// "fmt"
	"github.com/jcasado94/tfg/types"
	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
	// "io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	// "time"
	// "typesScraping"
)

/*
	pre:
	year, month, day represent a future date.
	orig and dest are valid places for the specific webpage database.
	0 < adults+children < 8
	0 <= children < 7 (2-11 years)
	0 <= babues < 2 (<2 years)
		option might be:
			0 - Aerolineas Argentinas
			1 - LAN
			2 - Plataforma10
	post: the trips of that day are returned
*/
func GetDayOffers(year, month, day, adults, children, babies, orig, dest string, option int) []types.Trip {
	switch {
	case option == 0:
		return getDayOffersAerolineas(year, month, day, adults, children, babies, orig, dest)
	case option == 1:
		return getDayOffersLAN(year, month, day, adults, children, babies, orig, dest)
	case option == 2:
		return getDayOffersP10(year, month, day, adults, children, babies, orig, dest)
	default:
		return []types.Trip{}
	}
}

/*
	pre: year, month, day represent a future date. orig and dest are valid places for Aerolineas website.
	post: the trips of that day are returned
*/
func getDayOffersAerolineas(year, month, day, adults, children, babies, orig, dest string) []types.Trip {

	var result []types.Trip

	c := make(chan types.Trip)
	numFlightsChan := make(chan int)
	go getFlightsAerolineas(100, initializeClient(), nil, year, month, day, adults, children, babies, orig, dest, numFlightsChan) // "tricky" call for knowing the number of flights in the web
	numFlights := <-numFlightsChan
	for i := 0; i < numFlights; i++ {
		go getFlightsAerolineas(i, initializeClient(), c, year, month, day, adults, children, babies, orig, dest, nil)
	}

	for i := 0; i < numFlights; i++ {
		result = append(result, <-c)
	}

	return result
}

//TODO
func getDayOffersLAN(year, month, day, adults, children, babies, orig, dest string) []types.Trip {

	// client := initializeClient()

	// myUrl := "http://booking.lan.com/ws/booking/quoting/fares_availability/5.0/rest/get_availability"
	// form := url.Values{}
	// form.Add("adults", adults)
	// form.Add("application", "compra_normal")
	// form.Add("cabin", "Y")
	// form.Add("children", children)
	// form.Add("country", "AR")
	// form.Add("departureDate", year+"-"+month+"-"+day)
	// form.Add("destination", dest)
	// form.Add("infants", babies)
	// form.Add("language", "ES")
	// form.Add("origin", orig)
	// form.Add("portal", "personas")
	// form.Add("roundTrip", "false")
	// form.Add("section", "step2")

	// var jsonStr = []byte(`{"language":"ES","country":"AR","portal":"personas","application":"compra_normal","section":"step2","cabin":"Y","adults":` + adults + `,"children":` + children + `,"infants":` + babies + `,"roundTrip":false,"departureDate":"` + year + "-" + month + "-" + day + `","origin":"` + orig + `","destination":"` + dest + `"}`)
	// req, err := http.NewRequest("POST", myUrl, bytes.NewBuffer(jsonStr))
	// req.Header.Set("Content-Type", "application/json")
	// resp, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer resp.Body.Close()
	// body, _ := ioutil.ReadAll(resp.Body)
	// var JSON map[string]interface{}
	// json.Unmarshal(body, &JSON)

	// //get routes from JSON
	// routes := JSON["data"].(map[string]interface{})["itinerary"].(map[string]interface{})["routesMap"].(map[string]interface{})

	// //itarate the routes
	// for k, v := range routes {
	// 	var trip types.Trip
	// 	trip.FlightNumber = k
	// 	trip.Url = `http://booking.lan.com/es_ar/apps/personas/compra?fecha1_dia=` + day + `&fecha1_anomes=` + year + "-" + month + `&auAvailability=1&ida_vuelta=ida&from_city1=` + orig + `&to_city1=` + dest + `&flex=1&cabina=Y&nadults=` + adults + `&nchildren=` + children + `&ninfants=` + babies
	// 	trip.UrlParams = nil
	// 	flight := v.(map[string]interface{})
	// 	travel := flight["travel"].(map[string]interface{})
	// 	origin := travel["origin"].(map[string]interface{})
	// 	destination := travel["destination"].(map[string]interface{})

	// 	//departure info
	// 	trip.DepAirp = origin["airport"].(map[string]interface{})["code"].(string)
	// 	date := origin["date"].(string)
	// 	hour := strings.Split(date, "T")[1]
	// 	hourMin := strings.Split(hour, ":")
	// 	trip.DepHour = strconv.Atoi(hourMin[0])
	// 	trip.DepMin = strconv.Atoi(hourMin[1])

	// 	//arrival info
	// 	trip.ArrAirp = destination["airport"].(map[string]interface{})["code"].(string)
	// 	date = destination["date"].(string)
	// 	hour = strings.Split(date, "T")[1]
	// 	hourMin = strings.Split(hour, ":")
	// 	trip.ArrHour = strconv.Atoi(hourMin[0])
	// 	trip.ArrMin = strconv.Atoi(hourMin[1])

	// 	//price
	// 	fares := flight["fareFamilyMap"].(map[string]interface{})
	// 	LE := fares["LE"].(map[string]interface{}) //cheapest
	// 	if LE["availability"].(int) == 0 {
	// 		FX := fares["FX"].(map[string]interface{})
	// 		if FX["availability"].(int) == 0 {
	// 			FF := fares["FF"].(map[string]interface{})
	// 			trip.PricePerAdult = FF["fare"].(map[string]interface{})["passengerMap"].(map[string]interface{})["adult"]
	// 		}
	// 	}

	// }
	return []types.Trip{}

}

//TODO
func getDayOffersP10(year, month, day, adults, children, babies, orig, dest string) []types.Trip {
	return []types.Trip{}
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
	scrapes for the departure and arrival times and the cheapest price for all the direct flights
*/
func getFlightsAerolineas(i int, client *http.Client, c chan types.Trip, year, month, day, adults, children, babies, orig, dest string, c1 chan int) {

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
	form1.Add("numChildren", children)
	form1.Add("numInfants", babies)
	form1.Add("cabin", "ALL")
	form1.Add("lang", "es_ES")

	r, _ := http.NewRequest("POST", myUrl1, bytes.NewBufferString(form1.Encode()))
	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Do(r)

	//connect to second website
	myUrl2 := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html?execution=e1s1"
	form2 := url.Values{}
	form2.Set("_eventId_next", "")
	r, _ = http.NewRequest("POST", myUrl2, bytes.NewBufferString(form2.Encode()))
	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ = client.Do(r)

	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)

	end := false

	it := 0

	for !end {
		//take all the ids
		tt := z.Next()

		switch {

		case tt == html.TextToken:

			if string(z.Raw()) == "Directo:" {

				//look for flight id, should be in next "input" tag
				tripFound := i != it //only if it's the wanted trip
				for !tripFound {

					tt = z.Next()

					switch {

					case tt == html.SelfClosingTagToken:

						t := z.Token()

						if t.Data == "input" {
							tripFound = true
							end = true
							id := strings.Split(t.Attr[0].Val, "_")[2]
							c <- getFlightAerolinias(id, myUrl1, &form1, client)
						}

					}
				}

				it++

			} else if string(z.Raw()) == "1 parada:" {
				//only first one should get here
				end = true
				c1 <- it
			}

		case tt == html.ErrorToken:
			//only first one should get here
			end = true
			c1 <- it
		}

	}

}

func getFlightAerolinias(id string, postUrl string, urlParams *url.Values, client *http.Client) types.Trip {

	var newTrip types.Trip
	newTrip.Url = postUrl
	newTrip.UrlParams = *urlParams

	//get json with info about flight
	myUrl := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html"
	form := url.Values{}
	form.Set("_eventId_ajax", "")
	form.Set("ajaxSource", "true")
	form.Set("contextObject", `{"transferObjects":[{"componentType":"cart","actionCode":"checkPrice","queryData":{"componentId":"cart_1","componentType":"cart","actionCode":"checkPrice","queryData":null,"requestPartials":["initialized"],"selectedBasketRefs":[`+id+`]}}]}`)
	form.Set("execution", "e1s2")

	r, _ := http.NewRequest("POST", myUrl, bytes.NewBufferString(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, _ := client.Do(r)

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var JSON map[string]interface{}
	json.Unmarshal(body, &JSON)

	model := JSON["content"].([]interface{})[0].(map[string]interface{})["model"].(map[string]interface{})

	//total price
	totalPrice := model["amountDuePrices"].(map[string]interface{})["priceAlternatives"].([]interface{})[0].(map[string]interface{})["pricesPerCurrency"].(map[string]interface{})["ARS"].(map[string]interface{})["amount"].(string)
	newTrip.TotalPrice, _ = strconv.ParseFloat(totalPrice, 64)

	//other info
	itineraryParts := model["itineraryParts"].([]interface{})[0].(map[string]interface{})
	segments := itineraryParts["segments"].([]interface{})[0].(map[string]interface{})
	flightNumber := strconv.FormatFloat(segments["flightNumber"].([]interface{})[0].(float64), 'f', 0, 64)
	newTrip.FlightNumber = "AR" + flightNumber
	newTrip.DepAirp = segments["departureCode"].(string)
	newTrip.ArrAirp = segments["arrivalCode"].(string)
	date := strings.Split(segments["departureDate"].(string), " ")[1]
	hour := strings.Split(date, ":")
	newTrip.DepHour, _ = strconv.Atoi(hour[0])
	newTrip.DepMin, _ = strconv.Atoi(hour[1])
	date = strings.Split(segments["arrivalDate"].(string), " ")[1]
	hour = strings.Split(date, ":")
	newTrip.ArrHour, _ = strconv.Atoi(hour[0])
	newTrip.ArrMin, _ = strconv.Atoi(hour[1])

	//adult price
	newTrip.PricePerAdult, _ = strconv.ParseFloat(itineraryParts["prices"].(map[string]interface{})["priceAlternatives"].([]interface{})[0].(map[string]interface{})["pricesPerCurrency"].(map[string]interface{})["ARS"].(map[string]interface{})["amount"].(string), 64)

	return newTrip
}

// /*
// 	scrapes for the departure and arrival times and the cheapest price for all the direct flights
// */
// func getDatePriceAerolineasOld(body *io.ReadCloser, postUrl string, urlParams *url.Values) []types.Trip {

// 	z := html.NewTokenizer(*body)
// 	result := []types.Trip{}
// 	end := false

// 	for end == false {
// 		tt := z.Next()
// 		switch {

// 		case tt == html.StartTagToken:

// 			t := z.Token()
// 			if t.Data == "span" && len(t.Attr) > 0 && t.Attr[0].Val == "airport_code" {
// 				var trip types.Trip

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
// 							z.Next() //should be TextNode with info
// 							if string(z.Raw()) == "Directo:" {
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
// 											trip.Url = postUrl
// 											trip.UrlParams = *urlParams
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
