package main

import (
	"fmt"
	"github.com/jcasado94/tfg/scraping"
	"runtime"
	// "github.com/jcasado94/tfg/types"
)

func main() {
	runtime.GOMAXPROCS(100000)
	year, month, day := "2016", "03", "19"
	orig, dest := "BUE", "COR"
	adults, children, babies := "1", "3", "0"
	trips := scraping.GetDayOffers(year, month, day, adults, children, babies, orig, dest, 0)
	fmt.Println(trips)
}
