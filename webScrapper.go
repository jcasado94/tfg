package main

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
	// "golang.org/x/net/html"
)

func getDayPrice(year, month, day string, body *io.ReadCloser) float64 {

	//start parsing
	z := html.NewTokenizer(*body)
	date := year + "/" + month + "/" + day
	// contents, _ := ioutil.ReadAll(*body)
	// fmt.Println(string(contents))

	for {
		tt := z.Next()
		switch {
		case tt == html.StartTagToken:
			t := z.Token()
			// fmt.Println(t.Data)
			if t.Data == "td" {
				for _, a1 := range t.Attr {
					if a1.Key == "data-tdc-date" && a1.Val == date {
						fmt.Println("IM HERE")
						for {
							tt = z.Next()
							switch {
							case tt == html.StartTagToken:
								t = z.Token()
								if t.Data == "span" {
									for _, a2 := range t.Attr {
										if a2.Key == "class" && a2.Val == "prices-amount" {
											fmt.Println(string(z.Text()))
											return 0.1
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
			return 0.2

		default:
			continue
		}

	}

	return 0.2

}

func main() {
	year, month, day := "2016", "02", "28"
	myUrl := "https://vuelos.aerolineas.com.ar/SSW2010/ARAR/webqtrip.html"
	//req, err := http.NewRequest("GET", myUrl, strings.NewReader(form.Encode()))
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
	form.Add("departureDate", "2016-02-28")
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

	myUrl2 := "http://ww1.aerolineas.com.ar/arg/bodies/reservas/procesarParametrosSABRE.asp?idSitio=AR&idIdioma=es&idavuelta=onewaytravel&idIATAOrigen=SLA&idIATAOrigenValue=19&idIATADestino=MDZ&idIATADestinoValue=7&calendarioIda=28%2f02%2f2016&calendarioHoras=anytimeFromHost&calendarioVuelta=29%2f02%2f2016&classService=ECONOMY&cantidadAdultos=1&cantidadChicos=0&cantidadBebes=0&claseEnCabina=E&pais=&currency=ARS&isMobile=False"
	client := &http.Client{Jar: jar}
	r, _ := http.NewRequest("GET", myUrl, bytes.NewBufferString(form.Encode()))
	// r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
	// r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Get(myUrl2)
	time.Sleep(4000 * time.Millisecond)
	resp, _ = client.Do(r)
	// time.Sleep(10000*time.Millisecond)
	defer resp.Body.Close()
	// z := html.NewTokenizer(resp.Body)
	// for {
	// 	z.Next()
	// 	fmt.Println(z.Token().Type)
	// }
	contents, err := ioutil.ReadAll(resp.Body)
	ioutil.WriteFile("output.txt", contents, 0644)
	// fmt.Println(string(contents))
	fmt.Println(getDayPrice(year, month, day, &resp.Body))
	/*contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
	            fmt.Printf("%s", err)
	            os.Exit(1)
	        }
	    fmt.Println(string(contents))*/
}
