package main

import (
	// "fmt"
	"github.com/jcasado94/tfg/CBR"
	"github.com/jcasado94/tfg/common"
	"github.com/jcasado94/tfg/db"
	"github.com/jcasado94/tfg/web"
	"net/http"
	// "github.com/jcasado94/tfg/scraping"
)

func main() {

	http.Handle("/usualCombinations", CBR.UsualCombinationsHandler{})
	http.Handle("/directTripsAerolineas", CBR.DirectTripsAerolineasHandler{})
	http.Handle("/directTripsLAN", CBR.DirectTripsLANHandler{})
	http.Handle("/directTripsPlat10", CBR.DirectTripsPlat10Handler{})
	http.Handle("/sameDayCombinations", CBR.SameDayCombinationsHandler{})
	http.Handle("/index", web.IndexHandler{})
	http.Handle("/flights", web.FlightsHandler{})
	http.Handle("/getDbCities", db.GetDbCitiesHandler{})

	fs := http.FileServer(http.Dir(common.WEB_STATIC_PATH))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(":8080", nil)
	// t1 := time.Now()
	// fmt.Println(scraping.GetDayOffers("2016", "05", "11", "5", "5", "0", "0", "10", "521", 2))
	// t2 := time.Now()
	// fmt.Println(t2.Sub(t1))
}
