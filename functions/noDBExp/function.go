package nodbexp

import (
	"github.com/makuc/diploma/modules/mytest"
	"net/http"
)

// BrezBaze is the exported function for executing
func BrezBaze(w http.ResponseWriter, r *http.Request) {
	w.Write(mytest.Message())
}
