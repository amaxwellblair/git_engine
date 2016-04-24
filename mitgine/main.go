package main

import (
	"net/http"

	"github.com/amaxwellblair/git_engine"
)

func main() {
	h := search.NewHandler()
	r := h.NewRouter()
	http.ListenAndServe(":9000", r)
}
