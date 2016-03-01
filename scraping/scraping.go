package scraping

import (
	"bytes"
	// "fmt"
	"github.com/jcasado94/tfg/types"
	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

/*
	pre: year, month, day represent a future date. orig and dest are valid places for the specific webpage database.
		option might be:
			0 - Aerolineas Argentinas
			1 - LAN
			2 - Plataforma10
	post: the trips of that day are returned
*/
func GetDayOffers(year, month, day, orig, dest string, option int) []types.Trip {
	switch {
	case option == 0:
		return getDayOffersAerolineas(year, month, day, orig, dest)
	case option == 1:
		return getDayOffersLAN(year, month, day, orig, dest)
	case option == 2:
		return getDayOffersP10(year, month, day, orig, dest)
	default:
		return []types.Trip{}
	}
}

/*
	pre: year, month, day represent a future date. orig and dest are valid places for Aerolineas website.
	post: the trips of that day are returned
*/
func getDayOffersAerolineas(year, month, day, orig, dest string) []types.Trip {

	client := initializeClient()

	//connect to first website after query
	myUrl := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html"
	form := url.Values{}
	form.Set("name", "ADVSForm")
	form.Set("id", "ADVSForm")
	form.Add("pointOfSale", "AR")
	form.Add("searchType", "CALENDAR")
	form.Add("currency", "ARS")
	form.Add("alternativeLandingPage", "true")
	form.Add("journeySpan", "OW")
	form.Add("origin", "BRC")
	form.Add("destination", "BUE")
	form.Add("departureDate", "2016-03-09")
	form.Add("numAdults", "1")
	form.Add("numChildren", "0")
	form.Add("numInfants", "0")
	form.Add("cabin", "ALL")
	form.Add("lang", "es_ES")

	r, _ := http.NewRequest("POST", myUrl, bytes.NewBufferString(form.Encode()))
	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Do(r)
	// body, _ := ioutil.ReadAll(resp.Body)
	// ioutil.WriteFile("output1.html", body, 0677)

	//connect to second website
	myUrl = "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html?execution=e1s1"
	form = url.Values{}
	form.Set("_eventId_next", "")
	r, _ = http.NewRequest("POST", myUrl, bytes.NewBufferString(form.Encode()))
	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ = client.Do(r)
	defer resp.Body.Close()
	// body, _ := ioutil.ReadAll(resp.Body)
	// ioutil.WriteFile("output2.html", body, 0677)

	return getDatePriceAerolineas(&resp.Body)
}

//TODO
func getDayOffersLAN(year, month, day, orig, dest string) []types.Trip {
	return []types.Trip{}
}

//TODO
func getDayOffersP10(year, month, day, orig, dest string) []types.Trip {
	return []types.Trip{}
}

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
func getDatePriceAerolineas(body *io.ReadCloser) []types.Trip {

	z := html.NewTokenizer(*body)
	result := []types.Trip{}
	end := false

	for end == false {
		tt := z.Next()
		switch {

		case tt == html.StartTagToken:

			t := z.Token()
			if t.Data == "span" && len(t.Attr) > 0 && t.Attr[0].Val == "translate time wasTranslated" {

				var trip types.Trip
				time := strings.Split(strings.Split(t.Attr[1].Val, ",")[0], " ")[1]
				divided := strings.Split(time, ":")
				trip.DepHour, _ = strconv.Atoi(divided[0])
				trip.DepMin, _ = strconv.Atoi(divided[1])
				found := false
				for found == false {
					tt = z.Next() // move till the departure time
					switch {
					case tt == html.StartTagToken:
						t = z.Token()
						if t.Data == "span" && len(t.Attr) > 0 && t.Attr[0].Val == "translate time wasTranslated" {
							time = strings.Split(strings.Split(t.Attr[1].Val, ",")[0], " ")[1]
							divided = strings.Split(time, ":")
							trip.ArrHour, _ = strconv.Atoi(divided[0])
							trip.ArrMin, _ = strconv.Atoi(divided[1])
							found = true
						}
					default:
						continue
					}
				}
				// check if it's a direct flight
				found = false
				for found == false {
					tt = z.Next()
					switch {
					case tt == html.StartTagToken:
						t = z.Token()
						if t.Data == "span" && len(t.Attr) > 0 && t.Attr[0].Val == "translate stops wasTranslated" {
							z.Next() //should be TextNode with info
							if string(z.Raw()) == "Directo:" {
								//look for flight's cheapest price
								found2 := false
								for found2 == false {
									tt = z.Next()
									switch {
									case tt == html.StartTagToken:
										t = z.Token()
										if t.Data == "span" && len(t.Attr) > 0 && t.Attr[0].Val == "prices-amount" {
											z.Next() //should be the price
											float, _ := strconv.ParseFloat(string(z.Raw()), 64)
											trip.Price = float
											result = append(result, trip)
											found = true
											found2 = true
										}

									default:
										continue
									}
								}
							} else {
								end = true
								found = true
							}
						}

					default:
						continue
					}
				}
			}

		case tt == html.ErrorToken:
			end = true

		default:
			continue

		}
	}
	return result
}

/*func getDayPrice(year, month, day string, body *io.ReadCloser) float64 {

	z := html.NewTokenizer(*body)

	date := year + "/" + month + "/" + day

	for {
		tt := z.Next()
		switch {
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == "td" {
				for _, a1 := range t.Attr {
					if a1.Key == "data-tdc-date" && strings.Split(a1.Val, " ")[0] == date { //date divided into day and hour
						for {
							tt = z.Next()
							switch {
							case tt == html.StartTagToken:
								t = z.Token()
								if t.Data == "span" {
									for _, a2 := range t.Attr {
										if a2.Key == "class" && a2.Val == "prices-amount" {
											tt = z.Next() //should be TextNode with price
											float, _ := strconv.ParseFloat(string(z.Raw()), 64)
											return float
										}
									}
								}
							default:
								continue
							}
						}
					}
				}
			}

		case tt == html.ErrorToken:
			return 0.0

		default:
			continue
		}

	}

}*/
