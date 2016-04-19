package mitgine

import (
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Handler serves as a global context
type Handler struct {
	client    *Client
	templates *template.Template
	secrets   map[string]string
	domain    string
}

// NewHandler creates a new handler
func NewHandler() *Handler {
	return &Handler{
		client:    NewClient(secrets()),
		templates: templates(),
		secrets:   secrets(),
		domain:    "http://localhost:9000",
	}
}

// NewRouter creates a new router
func (h *Handler) NewRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", h.rootHandler).
		Methods("GET")
	r.HandleFunc("/dashboard", h.dashboardHandler).
		Methods("GET")
	r.HandleFunc("/login", h.loginHandler).
		Methods("GET")
	r.HandleFunc("/logout", h.logoutHandler).
		Methods("DELETE")
	r.HandleFunc("/login/callback", h.loginCallbackHandler).
		Methods("GET")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))
	return handlers.HTTPMethodOverrideHandler(r)
}

func (h *Handler) rootHandler(w http.ResponseWriter, r *http.Request) {
	if isCurrentUser(r) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	h.templates.ExecuteTemplate(w, "index.html", nil)
}

func (h *Handler) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "dashboard.html", nil)
}

func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	// Create url
	u := new(url.URL)
	u.Scheme = "https"
	u.Host = "github.com"
	u.Path = "/login/oauth/authorize"
	params := u.Query()
	params.Add("client_id", h.secrets["clientID"])
	params.Add("redirect_uri", h.domain+"/login/callback")
	params.Add("scope", "public_repo")
	params.Add("state", h.secrets["githubState"])
	u.RawQuery = params.Encode()

	// Send a successful response
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (h *Handler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if isCurrentUser(r) {
		// Delete cookie
		cookie := http.Cookie{
			Name:     "token",
			Value:    "deleted",
			Path:     "/",
			Expires:  time.Now(),
			HttpOnly: true,
			MaxAge:   -1,
		}
		http.SetCookie(w, &cookie)
	}

	// Redirect to root
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) loginCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Parse parameters
	code := r.URL.Query().Get("code")

	// Request token from github
	resp, err := h.client.postAccessToken(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create domain wide cookie
	expire := time.Now().Add(time.Hour * 24 * 7)
	cookie := http.Cookie{
		Name:     "token",
		Value:    resp.AccessToken,
		Path:     "/",
		Expires:  expire,
		HttpOnly: true,
		MaxAge:   int(expire.Sub(time.Now()).Seconds()),
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func isCurrentUser(r *http.Request) bool {
	_, err := r.Cookie("token")
	return err != http.ErrNoCookie
}

func baseURL(r *http.Request) string {
	return r.URL.Scheme + r.URL.Host
}

func templates() *template.Template {
	return template.Must(template.ParseFiles(
		"static/index.html",
		"static/dashboard.html",
		"static/_header.html",
		"static/_nav.html",
		"static/_footer.html",
	))
}
