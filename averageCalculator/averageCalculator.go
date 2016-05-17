package main

import (
	"encoding/gob"
	"fmt"
	"github.com/jcasado94/tfg/common"
	// "github.com/jcasado94/tfg/scraping"
	// "github.com/jmcvetta/neoism"
	// "math/rand"
	"os"
	// "strconv"
)

const TRANSACTION_URL = "http://neo4j:k1llm3plz@localhost:7474/db/data"

const NUM_QUERIES = 100

func main() {

	// db, err := neoism.Connect(TRANSACTION_URL)
	// common.PanicErr(err)

	// type neoismAnswer struct {
	// 	Id   int    `json:"id"`
	// 	Code string `json:"airpCode"`
	// }

	// var ans []neoismAnswer
	// cq := neoism.CypherQuery{
	// 	Statement: `
	// 		MATCH (a:City)
	// 		RETURN id(a) AS id, a.airpCode AS airpCode
	// 	`,
	// 	Result: &ans,
	// }
	// err = db.Cypher(&cq)
	// common.PanicErr(err)

	var averages map[int]map[int]common.AveragePrice
	dataFile, _ := os.Open("averages.gob")
	dataDecoder := gob.NewDecoder(dataFile)
	_ = dataDecoder.Decode(&averages)

	dataFile.Close()

	fmt.Println(averages[0])

	// for _, city1 := range ans {

	// 	// averages[city1.Id] = make(map[int]common.AveragePrice)

	// 	for _, city2 := range ans {

	// 		averages[city1.Id][city2.Id] = common.AveragePrice{N: 0, Price: 0.0}

	// 		orig := city1.Code
	// 		dest := city2.Code

	// 		if orig == dest {
	// 			continue
	// 		}

	// 		c := make(chan float64)
	// 		c1 := make(chan int)

	// 		for i := 0; i < NUM_QUERIES; i++ {

	// 			go func() {

	// 				day := strconv.Itoa(rand.Intn(30) + 1)
	// 				month := strconv.Itoa(rand.Intn(8) + 5)
	// 				year := "2016"

	// 				trips := scraping.GetDayOffers(year, month, day, "1", "0", "0", orig, dest, 1)

	// 				n := len(trips)
	// 				sum := 0.0

	// 				for _, trip := range trips {
	// 					sum += trip.PricePerAdult
	// 				}

	// 				if n == 0 {
	// 					c <- 0.0
	// 				} else {
	// 					c <- sum
	// 					c1 <- n
	// 				}

	// 			}()

	// 		}

	// 		for i := 0; i < NUM_QUERIES; i++ {
	// 			price := <-c
	// 			if price != 0.0 {
	// 				n := <-c1
	// 				x := averages[city1.Id][city2.Id]
	// 				averages[city1.Id][city2.Id] = common.AveragePrice{Price: (float64(x.N)*x.Price + price) / float64(x.N+n), N: x.N + n}
	// 			}
	// 		}

	// 		fmt.Println("Average " + city1.Code + "->" + city2.Code)
	// 		fmt.Println(averages[city1.Id][city2.Id])

	// 	}

	// }

	// dataFile, err = os.Create("averages.gob")
	// common.PanicErr(err)

	// dataEncoder := gob.NewEncoder(dataFile)
	// dataEncoder.Encode(averages)

	// dataFile.Close()

}
