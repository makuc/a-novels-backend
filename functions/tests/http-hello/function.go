package httphello

import (
	"net/http"
)

// Function is the exported function for executing
func Function(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Hello World!\n"))
}
