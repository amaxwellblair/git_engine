package mitgine

import (
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

// Handler serves as a global context
type Handler struct {
	templates *template.Template
}

// NewHandler creates a new handler
func NewHandler() *Handler {
	return &Handler{
		templates: templates(),
	}
}

// NewRouter creates a new router
func (h *Handler) NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", h.rootHandler)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))
	return r
}

func (h *Handler) rootHandler(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "index.html", nil)
}

func templates() *template.Template {
	return template.Must(template.ParseFiles("static/index.html", "static/_header.html", "static/_nav.html"))
}
