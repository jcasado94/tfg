package main

import (
	"fmt"
	"github.com/jcasado94/tfg/scraping"
	// "github.com/jcasado94/tfg/types"
)

func main() {
	year, month, day := "2016", "03", "05"
	arr, dest := "SLA", "MDZ"
	trips := scraping.GetDayOffers(year, month, day, arr, dest, 0)
	fmt.Println(trips)
}
