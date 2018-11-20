package noDBExp

import (
	"net/http"
	""
)

func BrezBaze(w http.ResponseWriter, r *http.Request) {
	w.Write(message())
}
