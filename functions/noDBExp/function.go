package noDBExp

import (
	"diploma/pkg/mytest"
	"net/http"
)

func BrezBaze(w http.ResponseWriter, r *http.Request) {
	w.Write(mytest.Message())
}
