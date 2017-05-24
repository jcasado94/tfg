package common

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

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

// given the relative path from the project root, returns the system's absolute path.
func GetAbsPath(relPath string) string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, relPath)
}

// given the username and password, returns the url to connect to the db
func GetDBTransactionUrl() string {
	user := os.Getenv("NU")
	pw := os.Getenv("NP")
	return "http://" + user + ":" + pw + "@localhost:7474/db/data"
}

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}
