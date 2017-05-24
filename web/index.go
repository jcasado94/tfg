package web

import (
	"github.com/jcasado94/tfg/common"
	"html/template"
	"net/http"
)

type IndexHandler struct{}

func (IndexHandler) renderTemplate(w http.ResponseWriter) {
	t, _ := template.ParseFiles(common.GetAbsPath(common.WEB_HTML_PATH + `index.html`))
	emptyMap := make(map[string]interface{})
	t.Execute(w, emptyMap)
}

func (h IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	h.renderTemplate(w)

}
