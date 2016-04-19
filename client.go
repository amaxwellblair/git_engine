package mitgine

import (
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

func (c *Client) postAccessToken(code string) (*accessTokenResponse, error) {
	// Create request
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

type accessTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	State        string `json:"state"`
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}
