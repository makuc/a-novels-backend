package noDBExp

import (
	"github.com/makuc/diploma/pkg/mytest"
	"net/http"
)

func BrezBaze(w http.ResponseWriter, r *http.Request) {
	msg := mytest.Message()
	w.Write(msg)
}
