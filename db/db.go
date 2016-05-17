package db

import (
	"encoding/json"
	"fmt"
	"github.com/jcasado94/tfg/common"
	"github.com/jmcvetta/neoism"
	"net/http"
)

type GetDbCitiesHandler struct {
	Db *neoism.Database
}

func (h GetDbCitiesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	db, err := neoism.Connect(common.TRANSACTION_URL)
	common.PanicErr(err)

	res0 := []struct {
		Id   int    `json: "id"`
		Name string `json: "name"`
	}{}
	cq0 := neoism.CypherQuery{
		Statement: "MATCH (n:City) RETURN id(n) AS id, n.name AS name",
		Result:    &res0,
	}
	db.Cypher(&cq0)

	var bytes []byte
	bytes, _ = json.Marshal(res0)

	fmt.Fprintf(w, string(bytes))

}
