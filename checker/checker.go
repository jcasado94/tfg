package main

import (
	"encoding/gob"
	"fmt"
	"os"
)

func main() {
	var prices = make(map[int]map[int][]float64)
	var combinations = make(map[int]map[int][]int)
	var a, b []float64
	c, d := 0, 0
	var e, f []int

	dataFile, _ := os.Open("prices.gob")
	dataDecoder := gob.NewDecoder(dataFile)
	_ = dataDecoder.Decode(&prices)
	dataFile.Close()
	dataFile, _ = os.Open("combinations.gob")
	dataDecoder = gob.NewDecoder(dataFile)
	_ = dataDecoder.Decode(&combinations)
	dataFile.Close()
	dataFile, _ = os.Open("specTimes.gob")
	dataDecoder = gob.NewDecoder(dataFile)
	_ = dataDecoder.Decode(&a)
	dataFile.Close()
	dataFile, _ = os.Open("genTimes.gob")
	dataDecoder = gob.NewDecoder(dataFile)
	_ = dataDecoder.Decode(&b)
	dataFile.Close()
	dataFile, _ = os.Open("specInconsistent.gob")
	dataDecoder = gob.NewDecoder(dataFile)
	_ = dataDecoder.Decode(&c)
	dataFile.Close()
	dataFile, _ = os.Open("genInconsistent.gob")
	dataDecoder = gob.NewDecoder(dataFile)
	_ = dataDecoder.Decode(&d)
	dataFile.Close()
	dataFile, _ = os.Open("relationshipsSpec.gob")
	dataDecoder = gob.NewDecoder(dataFile)
	_ = dataDecoder.Decode(&e)
	dataFile.Close()
	dataFile, _ = os.Open("relationshipsGen.gob")
	dataDecoder = gob.NewDecoder(dataFile)
	_ = dataDecoder.Decode(&f)
	dataFile.Close()

	fmt.Println(prices)
	fmt.Println(combinations)
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	fmt.Println(d)
	fmt.Println(e)
	fmt.Println(f)
	fmt.Println(len(e))
	fmt.Println(len(f))
	fmt.Println(len(a))
	fmt.Println(len(b))

}
