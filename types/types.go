package types

type Trip struct {
	PricePerAdult    float64 // price per person
	TotalPrice       float64 // total price
	DepAirp, ArrAirp string  // departure and arrival airports. nil if bus trip.
	DepHour, DepMin  int
	ArrHour, ArrMin  int
	FlightNumber     string              // "LA4240", "AR2200", etc.
	Url              string              // url to the buying site
	UrlParams        map[string][]string // post parameters if url method=POST. if method=GET, equals nil.
}
