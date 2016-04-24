package search

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Client holds relevant
type Client struct {
	baseURL *url.URL
	secrets map[string]string
}

// NewClient creates a new instance of Client
func NewClient(secrets map[string]string) *Client {
	url := new(url.URL)
	url.Host = "api.github.com"
	url.Scheme = "https"
	return &Client{
		baseURL: url,
		secrets: secrets,
	}
}

func (c *Client) getCommits(token, name, owner string) ([]*GitCommit, error) {
	// Create URL
	u := c.baseURL
	u.Path = fmt.Sprintf("/repos/%s/%s/commits", owner, name)

	// Create request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+token)

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Parse response
	var commits []*GitCommit
	err = json.NewDecoder(resp.Body).Decode(&commits)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func (c *Client) getRepositories(token string) ([]*Repository, error) {
	// Create URL
	u := c.baseURL
	u.Path = "/user/repos"

	// Create request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+token)

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Parse response
	var repos []*Repository
	if err = json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return repos, nil
}

func (c *Client) getUsername(token string) (string, error) {
	// Create URL
	u := c.baseURL
	u.Path = "/user"

	// Create request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "token "+token)

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	// Parse response
	var user User
	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}

	return user.Username, nil
}

func (c *Client) postAccessToken(code string) (*accessTokenResponse, error) {
	// Create URL
	u := new(url.URL)
	u.Scheme = "https"
	u.Host = "github.com"
	u.Path = "/login/oauth/access_token"
	params := u.Query()
	params.Add("client_id", c.secrets["clientID"])
	params.Add("client_secret", c.secrets["clientSecret"])
	params.Add("code", code)
	params.Add("redirect_uri", "http://localhost:9000/login/callback")
	params.Add("state", c.secrets["githubState"])
	u.RawQuery = params.Encode()

	// Create request
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Send request
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil && resp.StatusCode != 302 {
		return nil, err
	}

	// Parse response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	params, err = url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	// Return response
	return &accessTokenResponse{
		AccessToken: params.Get("access_token"),
		Scope:       params.Get("scope"),
		TokenType:   params.Get("token_type"),
	}, nil
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

// Repository holds information for a github repository
type Repository struct {
	ID     int    `json:"id,omitempty"`
	Name   string `json:"name"`
	Active bool   `json:"active,omitempty"`
}

// User holds information for a github user
type User struct {
	Username string `json:"login"`
}

// GitCommit holds Github commits from a specific repository
type GitCommit struct {
	HTML   string  `json:"html_url"`
	Commit *Commit `json:"commit"`
}

// Commit holds the commit message
type Commit struct {
	Message string `json:"message"`
}
