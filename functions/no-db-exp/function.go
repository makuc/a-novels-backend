package nodbexp

import (
	"net/http"
)

// BrezBaze is the exported function for executing
func BrezBaze(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Hello World!\n"))
}
