package main

import (
	"bytes"
	"fmt"
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

func getDayPrice(year, month, day string, body *io.ReadCloser) float64 {

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

}

func main() {
	year, month, day := "2016", "03", "05"
	myUrl := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html"
	form := url.Values{}
	form.Set("name", "ADVSForm")
	form.Set("id", "ADVSForm")
	form.Add("pointOfSale", "AR")
	form.Add("searchType", "CALENDAR")
	form.Add("currency", "ARS")
	form.Add("alternativeLandingPage", "true")
	form.Add("journeySpan", "OW")
	form.Add("origin", "SLA")
	form.Add("destination", "MDZ")
	form.Add("departureDate", "2016-03-05")
	form.Add("numAdults", "1")
	form.Add("numChildren", "0")
	form.Add("numInfants", "0")
	form.Add("cabin", "ALL")
	form.Add("lang", "es_ES")

	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}

	jar, err := cookiejar.New(&options)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{Jar: jar}
	r, _ := http.NewRequest("POST", myUrl, bytes.NewBufferString(form.Encode()))
	r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Do(r)
	defer resp.Body.Close()
	fmt.Println(getDayPrice(year, month, day, &resp.Body))
}
