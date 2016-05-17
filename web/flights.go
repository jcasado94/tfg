package web

import (
	// "fmt"
	"github.com/jcasado94/tfg/common"
	"html/template"
	"net/http"
)

type GetParameters struct {
	Adults     string
	Children11 string
	Children5  string
	Infants    string
	DepId      string
	DepName    string
	ArrId      string
	ArrName    string
	Year       string
	Month      string
	Day        string
}

type FlightsHandler struct {
}

func (h FlightsHandler) renderTemplate(w http.ResponseWriter, getParams *GetParameters) {
	t, _ := template.ParseFiles(common.WEB_HTML_PATH + `flights.html`)
	t.Execute(w, *getParams)
}

func (h FlightsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()

	getParams := GetParameters{
		Adults:     params.Get("adults"),
		Children11: params.Get("children11"),
		Children5:  params.Get("children5"),
		Infants:    params.Get("infants"),
		DepId:      params.Get("dep"),
		DepName:    params.Get("depName"),
		ArrId:      params.Get("arr"),
		ArrName:    params.Get("arrName"),
		Year:       params.Get("year"),
		Month:      params.Get("month"),
		Day:        params.Get("day"),
	}

	h.renderTemplate(w, &getParams)

}
