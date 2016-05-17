package main

import (
	"bytes"
	"fmt"
	// "io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
)

func main() {

	client := &http.Client{}
	var myUrl string

	for {

		depID := rand.Intn(35)
		arrID := rand.Intn(35)
		for depID == arrID {
			arrID = rand.Intn(35)
		}

		// switch rand.Intn(2) {
		// case 0:
		myUrl = "http://localhost:8080/directTrips"

		// case 1:
		// 	myUrl = "http://localhost:8080/usualCombinations"

		// case 2:
		// 	myUrl = "http://localhost:8080/sameDayCombinations"

		var jsonStr = []byte(`{"Year":"2016","Month":"` + strconv.Itoa(rand.Intn(8)+5) +
			`","Day":"` + strconv.Itoa(rand.Intn(31)+1) +
			`","DepID":"` + strconv.Itoa(depID) +
			`","ArrID":"` + strconv.Itoa(arrID) +
			`","Adults":"` + strconv.Itoa(rand.Intn(4)+1) +
			`","Children":"` + strconv.Itoa(rand.Intn(1)+1) + `","Infants":"0"}`)
		// fmt.Println(string(jsonStr))
		r, _ := http.NewRequest("POST", myUrl, bytes.NewBuffer(jsonStr))
		r.Header.Add("Content-Type", "application/json")
		_, err := client.Do(r)
		if err != nil {
			fmt.Println(err)
		}

		// body, _ := ioutil.ReadAll(resp.Body)
		// fmt.Println(string(body))

	}

}
