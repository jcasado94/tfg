package common

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
