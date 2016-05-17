package averages

import (
	"encoding/gob"
	// "fmt"
	"github.com/jcasado94/tfg/common"
	"os"
	"strconv"
	"sync"
	// "time"
)

// mutex for controlling the access to the file
var mu sync.Mutex

// given a sum of prices and a n (number of prices), refreshes the average price in the database
func RetainAverage(depId, arrId string, sols [][]common.Trip) {

	// t1 := time.Now()

	if len(sols) == 0 {
		return
	}

	id1, _ := strconv.Atoi(depId)
	id2, _ := strconv.Atoi(arrId)

	mu.Lock()

	// read averages
	var averages map[int]map[int]common.AveragePrice
	dataFile, _ := os.Open("averages.gob")
	dataDecoder := gob.NewDecoder(dataFile)
	_ = dataDecoder.Decode(&averages)

	dataFile.Close()

	if _, exists := averages[id1]; !exists {
		averages[id1] = make(map[int]common.AveragePrice)
	}

	wg := &sync.WaitGroup{}
	var mu2 sync.Mutex

	for _, sol := range sols {

		solFunc := sol
		wg.Add(1)

		go func() {

			sumPrices := 0.0
			for _, trip := range solFunc {
				sumPrices += trip.PricePerAdult
			}

			// write averages
			mu2.Lock()
			x := averages[id1][id2]
			averages[id1][id2] = common.AveragePrice{Price: (float64(x.N)*x.Price + sumPrices) / float64(x.N+1), N: x.N + 1}
			mu2.Unlock()

			wg.Done()

		}()

	}

	wg.Wait()

	// write file
	dataFile, err := os.Create("averages.gob")
	common.PanicErr(err)

	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(averages)

	dataFile.Close()

	mu.Unlock()

	// t2 := time.Now()
	// fmt.Println(t2.Sub(t1))

}
