package main

import (
	"net/http"

	"github.com/amaxwellblair/git_engine"
)

func main() {
	h := mitgine.NewHandler()
	r := h.NewRouter()
	http.ListenAndServe(":9000", r)
}
