package mitgine

import (
	"net/http"
	"net/url"
	"text/template"

	"github.com/gorilla/mux"
)

// Handler serves as a global context
type Handler struct {
	templates *template.Template
	secrets   map[string]string
}

// NewHandler creates a new handler
func NewHandler() *Handler {
	return &Handler{
		templates: templates(),
		secrets:   secrets(),
	}
}

// NewRouter creates a new router
func (h *Handler) NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", h.rootHandler)
	r.HandleFunc("/login", h.loginHandler)
	r.HandleFunc("/login/callback", h.loginCallbackHandler)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))
	return r
}

func (h *Handler) rootHandler(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "index.html", nil)
}

func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	// Create url
	u := new(url.URL)
	params := u.Query()
	params.Add("client_id", h.secrets["client_id"])
	params.Add("redirect_uri", baseURL(r)+"/login/callback")
	params.Add("scope", "public_repo")

	// Create request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send a successful response
	http.Redirect(w, req, "https://github.com/login/oauth/authorize", http.StatusFound)
}

func (h *Handler) loginCallbackHandler(w http.ResponseWriter, r *http.Request) {

}

func baseURL(r *http.Request) string {
	return r.URL.Scheme + r.URL.Host
}

func templates() *template.Template {
	return template.Must(template.ParseFiles("static/index.html", "static/_header.html", "static/_nav.html"))
}
