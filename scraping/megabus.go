package scraping

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/jcasado94/tfg/common"
)

/*
	pre: year, month, day represent a future date.
		orig and dest are valid ids for Megabus website.
		adults > 0
	post: the trips of that day are returned

*/

func GetDayOffersMegabus(year, month, day, adults, orig, dest string) []common.Trip {

	adultsInt, _ := strconv.Atoi(adults)
	var result []common.Trip

	client := initializeClient()
	myUrl := "https://us.megabus.com/JourneyResults.aspx?originCode=" + orig + "&destinationCode=" + dest +
		"&outboundDepartureDate=" + month + "%2f" + day + "%2f" + year +
		"&inboundDepartureDate=&passengerCount=" + adults +
		"&transportType=0&concessionCount=0&nusCount=0&outboundWheelchairSeated=0&outboundOtherDisabilityCount=0&inboundWheelchairSeated=0&inboundOtherDisabilityCount=0&outboundPcaCount=0&inboundPcaCount=0&promotionCode=&withReturn=0"

	req, err := http.NewRequest("POST", myUrl, nil)
	if err != nil {
		return []common.Trip{}
	}

	resp, err := client.Do(req)
	if err != nil {
		return []common.Trip{}
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	for {

		tt := z.Next()
		t := z.Token()

		if tt == html.TextToken && strings.TrimSpace(t.Data) == "Departs" {

			/* departure time */
			var trip common.Trip
			trip.Url = myUrl

			z.Next()
			z.Next()
			t = z.Token()
			trimmed := strings.TrimSpace(t.Data)
			x := strings.Split(trimmed, ":") // weird stuff with splitting with " " (utf-8 related...)
			depHour := x[0]
			depHourInt, _ := strconv.Atoi(depHour)
			depMin := x[1][0:2] // take the two numbers after the ':'
			depMinInt, _ := strconv.Atoi(depMin)

			if strings.ContainsRune(trimmed, 'P') { // this means there is PM, so afternoon time
				depHourInt += 12
			}

			trip.DepHour = depHourInt
			trip.DepMin = depMinInt
			trip.DepDay, _ = strconv.Atoi(day)
			trip.DepMonth, _ = strconv.Atoi(month)
			trip.DepYear, _ = strconv.Atoi(year)

			for {

				tt = z.Next()
				t = z.Token()

				/* arrival time */
				if tt == html.StartTagToken && len(t.Attr) == 1 && t.Attr[0].Val == "three" { // trip duration tag

					duration := 0 // minutes
					z.Next()
					z.Next()
					z.Next()
					t = z.Token()
					trimmed = strings.TrimSpace(t.Data)
					index := strings.Index(trimmed, "hrs")
					if index != -1 {
						hoursStr := trimmed[0:index]
						hours, _ := strconv.Atoi(hoursStr)
						duration += hours * 60
					}

					index = strings.Index(trimmed, "mins")
					if index != -1 {
						minsStr := trimmed[index-2 : index]
						minsStr = strings.TrimSpace(minsStr)
						mins, _ := strconv.Atoi(minsStr)
						duration += mins
					}

					newYorkLocation, _ := time.LoadLocation("America/New_York")
					depDatePlusDuration := time.Date(trip.DepYear, time.Month(trip.DepMonth), trip.DepDay, trip.DepHour, trip.DepMin+duration, 0, 0, newYorkLocation)
					// depDatePlusDuration.Add(time.Duration(duration) * time.Minute)
					trip.ArrMin = depDatePlusDuration.Minute()
					trip.ArrHour = depDatePlusDuration.Hour()
					trip.ArrDay = depDatePlusDuration.Day()
					trip.ArrMonth = int(depDatePlusDuration.Month())
					trip.ArrYear = depDatePlusDuration.Year()

					/* price */
					for {

						tt = z.Next()
						t = z.Token()

						/* arrival time */
						if tt == html.StartTagToken && len(t.Attr) == 1 && t.Attr[0].Val == "five" { // trip duration tag

							z.Next()
							z.Next()
							z.Next()
							z.Next()
							z.Next()
							z.Next()
							z.Next() // price should be here
							t = z.Token()
							priceStr := strings.TrimSpace(t.Data)
							priceStr = priceStr[1:]
							totalPrice, _ := strconv.ParseFloat(priceStr, 32)
							trip.TotalPrice = totalPrice
							trip.PricePerAdult = trip.TotalPrice / float64(adultsInt)

							result = append(result, trip)

							break

						}

					}

					break

				}

			}

		} else if tt == html.StartTagToken && len(t.Attr) == 1 && t.Attr[0].Val == "footer" { // trip duration tag
			// end
			break
		}
	}

	return result

}
