package handlers

import (
	"net/http"
	"fmt"
	"strings"
	
	"github.com/kouzant/go-short/storage"
)

type GoShortHandler struct {
	StateStore storage.StateStore
}

func (h *GoShortHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokens := strings.SplitAfterN(r.URL.Path, "/", 2)
	value, error := h.StateStore.Load(storage.StorageKey(tokens[1]))
	if error != nil {
		fmt.Fprintf(w, "Error: %v", error)
	} else {
		http.Redirect(w, r, value.(string), http.StatusTemporaryRedirect)
	}	
}
