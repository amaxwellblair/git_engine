package search

import (
	"encoding/json"
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
	store     *Store
	templates *template.Template
	secrets   map[string]string
	domain    string
}

// NewHandler creates a new handler
func NewHandler() *Handler {
	return &Handler{
		client:    NewClient(secrets()),
		store:     NewStore(),
		templates: templates(),
		secrets:   secrets(),
		domain:    "http://localhost:9000",
	}
}

// NewRouter creates a new router
func (h *Handler) NewRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", h.getRootHandler).
		Methods("GET")
	r.HandleFunc("/dashboard", h.getDashboardHandler).
		Methods("GET")
	r.HandleFunc("/repositories", h.getRepositoriesHandler).
		Methods("GET")
	r.HandleFunc("/repositories/activate", h.postActivateRepositoryHandler).
		Methods("POST")
	r.HandleFunc("/login", h.getLoginHandler).
		Methods("GET")
	r.HandleFunc("/logout", h.deleteLogoutHandler).
		Methods("DELETE")
	r.HandleFunc("/login/callback", h.getLoginCallbackHandler).
		Methods("GET")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))
	return handlers.HTTPMethodOverrideHandler(r)
}

func (h *Handler) getRootHandler(w http.ResponseWriter, r *http.Request) {
	if token := currentUser(r); token != "" {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	h.templates.ExecuteTemplate(w, "index.html", nil)
}

func (h *Handler) getDashboardHandler(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "dashboard.html", nil)
}

func (h *Handler) postActivateRepositoryHandler(w http.ResponseWriter, r *http.Request) {
	token := currentUser(r)
	if token == "" {
		http.Error(w, "unauthorized user", http.StatusForbidden)
		return
	}

	// Parse request
	if err := r.ParseForm(); err != nil {
		http.Error(w, "unauthorized user", http.StatusForbidden)
		return
	}
	name := r.FormValue("name")

	// Check if a repository already exists
	if !h.store.RepoExists(token, name) {

		// Create and populate the repository with commits
		un, err := h.client.getUsername(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if commits, err := h.client.getCommits(token, name, un); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if err = h.store.CreateRepository(name, un, token, commits); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Update repositorylist with active status
	if err := h.store.ActivateRepository(token, name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *Handler) getRepositoriesHandler(w http.ResponseWriter, r *http.Request) {
	token := currentUser(r)
	if token == "" {
		http.Error(w, "unauthorized user", http.StatusForbidden)
		return
	}

	// Retrieve repositores from elastic search
	search := r.URL.Query().Get("term")
	repos, err := h.store.GetRepositories(token, search)
	if err != nil && (err.Error() == "no repository type exists for this token" || err.Error() == "no user exists for this token") {

		// Create a user if no user exists
		if err.Error() == "no user exists for this token" {

			if err := h.store.CreateUserIndex(token); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Retrieve repositories from Github
		repos, err = h.client.getRepositories(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Place respositories in elastic search
		for i := 0; i < len(repos); i++ {
			if err := h.store.CreateRepositoryList(token, repos[i]); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Format data for autocomplete
	var results []string
	for _, r := range repos {
		results = append(results, r.Name)
	}

	// Send successful response
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getLoginHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) deleteLogoutHandler(w http.ResponseWriter, r *http.Request) {
	if token := currentUser(r); token != "" {
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

func (h *Handler) getLoginCallbackHandler(w http.ResponseWriter, r *http.Request) {
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

func currentUser(r *http.Request) string {
	token, err := r.Cookie("token")
	if err == http.ErrNoCookie {
		return ""
	}
	return token.Value
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
