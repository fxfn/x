package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type GrantPasswordOpts struct {
	Username     string
	Password     string
	Scope        string
	ClientID     string
	ClientSecret string
}

func (a *Auth) GrantPassword(opts GrantPasswordOpts) (*Token, error) {
	if a.server == nil {
		return nil, &InvalidRequest{
			message: "use auth.SetServer() or auth.Discovery() to set the server",
		}
	}

	tokenEndpoint := a.server.TokenEndpoint

	form := url.Values{
		"grant_type":    {"password"},
		"scope":         {opts.Scope},
		"client_id":     {opts.ClientID},
		"client_secret": {opts.ClientSecret},
	}

	res, err := http.PostForm(tokenEndpoint, form)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var token Token
	json.Unmarshal(body, &token)

	if token.Error == "unsupported_grant_type" {
		return nil, &InvalidRequest{
			message: token.ErrorDescription,
		}
	}

	return &token, nil
}
