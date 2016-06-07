package main

import (
	"bytes"
	"fmt"
	// "io/ioutil"
	"encoding/gob"
	"encoding/json"
	"github.com/jcasado94/tfg/common"
	"github.com/jmcvetta/neoism"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	// "time"
)

func main() {

	db, _ := neoism.Connect(common.TRANSACTION_URL)

	//USH->SLA, IGR->USH, JUJ->FTE, RGL->MDQ, IGR->FTE
	pairs := [][2]int{{33, 23}, {10, 33}, {11, 7}, {19, 13}, {33, 7}}

	var heuristic = make(map[int]map[int]float64)
	for i := 0; i < 40; i++ {
		heuristic[i] = make(map[int]float64)
	}
	dataFile, _ := os.Create("heuristicGen.gob")
	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(heuristic)
	dataFile.Close()
	dataFile, _ = os.Create("heuristicSpec.gob")
	dataEncoder = gob.NewEncoder(dataFile)
	dataEncoder.Encode(heuristic)
	dataFile.Close()
	var prices = make(map[int]map[int][]float64)
	var combinations = make(map[int]map[int][]int)
	for _, p := range pairs {
		prices[p[0]] = make(map[int][]float64)
		combinations[p[0]] = make(map[int][]int)
	}
	dataFile, _ = os.Create("prices.gob")
	dataEncoder = gob.NewEncoder(dataFile)
	dataEncoder.Encode(prices)
	dataFile.Close()
	dataFile, _ = os.Create("combinations.gob")
	dataEncoder = gob.NewEncoder(dataFile)
	dataEncoder.Encode(combinations)
	dataFile.Close()
	var a []float64
	dataFile, _ = os.Create("specTimes.gob")
	dataEncoder = gob.NewEncoder(dataFile)
	dataEncoder.Encode(a)
	dataFile.Close()
	dataFile, _ = os.Create("genTimes.gob")
	dataEncoder = gob.NewEncoder(dataFile)
	dataEncoder.Encode(a)
	dataFile.Close()
	b := 0
	dataFile, _ = os.Create("specInconsistent.gob")
	dataEncoder = gob.NewEncoder(dataFile)
	dataEncoder.Encode(b)
	dataFile.Close()
	dataFile, _ = os.Create("genInconsistent.gob")
	dataEncoder = gob.NewEncoder(dataFile)
	dataEncoder.Encode(b)
	dataFile.Close()
	var c []int
	dataFile, _ = os.Create("relationshipsSpec.gob")
	dataEncoder = gob.NewEncoder(dataFile)
	dataEncoder.Encode(c)
	dataFile.Close()
	dataFile, _ = os.Create("relationshipsGen.gob")
	dataEncoder = gob.NewEncoder(dataFile)
	dataEncoder.Encode(c)
	dataFile.Close()

	client := &http.Client{}
	for i := 0; i < 1000; i++ {
		// execute 100 directTrips
		var wg sync.WaitGroup
		for j := 0; j < 50; j++ {
			wg.Add(1)
			go func() {
				depID := rand.Intn(35)
				arrID := rand.Intn(35)
				for depID == arrID {
					depID = rand.Intn(35)
				}
				var jsonStr = []byte(`{"Year":"2016","Month":"6","Day":"` + strconv.Itoa(rand.Intn(4)+10) +
					`","DepID":"` + strconv.Itoa(depID) +
					`","ArrID":"` + strconv.Itoa(arrID) +
					`","Adults":"1","Children":"0","Infants":"0"}`)

				var wg2 sync.WaitGroup
				wg2.Add(3)
				go func() {
					r, _ := http.NewRequest("POST", "http://localhost:8080/directTripsAerolineas", bytes.NewBuffer(jsonStr))
					r.Header.Add("Content-Type", "application/json")
					_, err := client.Do(r)
					if err != nil {
						fmt.Println(err)
					}
					wg2.Done()
				}()
				go func() {
					r, _ := http.NewRequest("POST", "http://localhost:8080/directTripsLAN", bytes.NewBuffer(jsonStr))
					r.Header.Add("Content-Type", "application/json")
					_, err := client.Do(r)
					if err != nil {
						fmt.Println(err)
					}
					wg2.Done()
				}()
				go func() {
					r, _ := http.NewRequest("POST", "http://localhost:8080/directTripsPlat10", bytes.NewBuffer(jsonStr))
					r.Header.Add("Content-Type", "application/json")
					_, err := client.Do(r)
					if err != nil {
						fmt.Println(err)
					}
					wg2.Done()
				}()
				wg2.Wait()
				wg.Done()
			}()
		}
		wg.Wait()
		// direct trips done
		// get number of rels in db
		res0 := []struct {
			R int `json:"R"`
		}{}
		cq0 := neoism.CypherQuery{
			Statement: "MATCH ()-[r:SPEC]->() RETURN count(r) AS R",
			Result:    &res0,
		}
		db.Cypher(&cq0)
		rels := res0[0].R

		dataFile, _ := os.Open("relationshipsSpec.gob")
		dataDecoder := gob.NewDecoder(dataFile)
		_ = dataDecoder.Decode(&c)
		dataFile.Close()
		c = append(c, rels)
		dataFile, _ = os.Create("relationshipsSpec.gob")
		dataEncoder := gob.NewEncoder(dataFile)
		dataEncoder.Encode(c)
		dataFile.Close()

		res0 = []struct {
			R int `json:"R"`
		}{}
		cq0 = neoism.CypherQuery{
			Statement: "MATCH ()-[r:GEN]->() RETURN count(r) AS R",
			Result:    &res0,
		}
		db.Cypher(&cq0)
		rels = res0[0].R

		dataFile, _ = os.Open("relationshipsGen.gob")
		dataDecoder = gob.NewDecoder(dataFile)
		_ = dataDecoder.Decode(&c)
		dataFile.Close()
		c = append(c, rels)
		dataFile, _ = os.Create("relationshipsGen.gob")
		dataEncoder = gob.NewEncoder(dataFile)
		dataEncoder.Encode(c)
		dataFile.Close()

		for j := 0; j < len(pairs); j++ {
			dep, arr := pairs[j][0], pairs[j][1]
			var jsonStr = []byte(`{"Year":"2016","Month":"6","Day":"10","DepID":"` + strconv.Itoa(dep) +
				`","ArrID":"` + strconv.Itoa(arr) +
				`","Adults":"1","Children":"0","Infants":"0"}`)

			c := make(chan [][]map[string]interface{})
			// t1 := time.Now()
			go func() {
				r, _ := http.NewRequest("POST", "http://localhost:8080/usualCombinations", bytes.NewBuffer(jsonStr))
				r.Header.Add("Content-Type", "application/json")
				resp, err := client.Do(r)
				if err != nil {
					fmt.Println(err)
				}
				decoder := json.NewDecoder(resp.Body)
				var x [][]map[string]interface{}
				err = decoder.Decode(&x)
				if err != nil {
					fmt.Println(err)
				}
				c <- x
			}()
			go func() {
				r, _ := http.NewRequest("POST", "http://localhost:8080/sameDayCombinations", bytes.NewBuffer(jsonStr))
				r.Header.Add("Content-Type", "application/json")
				resp, err := client.Do(r)
				if err != nil {
					fmt.Println(err)
				}
				decoder := json.NewDecoder(resp.Body)
				var x [][]map[string]interface{}
				err = decoder.Decode(&x)
				if err != nil {
					fmt.Println(err)
				}
				c <- x
			}()

			combination := 0
			minPrice := 0.0
			for k := 0; k < 2; k++ {
				x := <-c
				// if k == 1 {
				// 	fmt.Println(time.Now().Sub(t1))
				// }
				combination += len(x)
				for _, trips := range x {
					price := 0.0
					for _, trip := range trips {
						price += trip["TotalPrice"].(float64)
					}
					if price < minPrice || minPrice == 0.0 {
						minPrice = price
					}
				}
			}

			dataFile, _ = os.Open("prices.gob")
			dataDecoder := gob.NewDecoder(dataFile)
			_ = dataDecoder.Decode(&prices)
			dataFile.Close()
			prices[dep][arr] = append(prices[dep][arr], minPrice)
			dataFile, _ := os.Create("prices.gob")
			dataEncoder = gob.NewEncoder(dataFile)
			dataEncoder.Encode(prices)
			dataFile.Close()

			dataFile, _ = os.Open("combinations.gob")
			dataDecoder = gob.NewDecoder(dataFile)
			_ = dataDecoder.Decode(&combinations)
			dataFile.Close()
			combinations[dep][arr] = append(combinations[dep][arr], combination)
			dataFile, _ = os.Create("combinations.gob")
			dataEncoder = gob.NewEncoder(dataFile)
			dataEncoder.Encode(combinations)
			dataFile.Close()

		}

	}

}
