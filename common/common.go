package common

import (
	"encoding/json"
	"net/http"
)

// DATABASE GLOBALS
const TRANSP_AEROL = 0
const TRANSP_LAN = 1
const TRANSP_BUS = 2

// DATABASE CONNECTION
const TRANSACTION_URL = "http://neo4j:k1llm3plz@localhost:7474/db/data"

// TRANSFER TIME (mins)
const TRANSFER_TIME = 120

// MAX HOURS OF TRANSFER
const MAX_TRANSFER_HOURS = 12

// MAX TRANSFERS
const MAX_TRANSFERS = 5

// NUMBER OF DAYS AWAY FROM DEP DAY FOR SPECIFIC TRIP LOOKUP
const MAX_DAYS_SPEC = 4

// // PATH TO HTML FILES
const WEB_HTML_PATH = `D:\Users\Documentos\GOlang\src\github.com\jcasado94\tfg\web\html\`

// PATH TO WEB STATIC ELEMENTS
const WEB_STATIC_PATH = `D:\Users\Documentos\GOlang\src\github.com\jcasado94\tfg\web\static`

// NUMBER OF DAYS AWAY FROM THE DEPARTURE DAY TO LOOK FOR IN GENERAL PATHS
const GEN_DAYS_SCOPE = 1

type Trip struct {
	Id                                         int     // handled by the scraper caller
	PricePerAdult                              float64 // price per person
	TotalPrice                                 float64 // total price
	DepAirp, ArrAirp                           string  // departure and arrival airports. plat10name if bus.
	DepYear, DepMonth, DepDay, DepHour, DepMin int
	ArrYear, ArrMonth, ArrDay, ArrHour, ArrMin int
	FlightNumber                               string              // "LA4240", "AR2200", etc. In bus trips, company name and quality ("El Rapido - Semicama", "El Turista - Cama Plus"...)
	Url                                        string              // url to the buying site
	UrlParams                                  map[string][]string // post parameters if url method=POST. if method=GET, equals nil.
}

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

const FILE_SPEC_HEURISTIC = "heuristicSpec.gob"
const FILE_GEN_HEURISTIC = "heuristicGen.gob"

type AveragePrice struct {
	N     int
	Price float64
}

type Query struct {
	Year       string `json:"Year"`
	Month      string `json:"Month"`
	Day        string `json:"Day"`
	DepID      string `json:"DepID"`
	ArrID      string `json:"ArrID"`
	Adults     string `json:"Adults"`
	Children11 string `json:"Children11"`
	Children5  string `json:"Children5"`
	Infants    string `json:"Infants"`
}

//decodes the request in the form of a query
func DecodeRequest(r *http.Request) Query {
	decoder := json.NewDecoder(r.Body)
	var x Query
	err := decoder.Decode(&x)
	PanicErr(err)
	return x
}

//encodes the solutions found into a json
func EncodeAnswer(sols [][]Trip) string {
	var bytes []byte
	bytes, _ = json.Marshal(sols)
	return string(bytes)
}
