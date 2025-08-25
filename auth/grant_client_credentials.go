package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ClientCredentials struct {
	ClientID     string
	ClientSecret string
	Scope        string
}

type GrantClientCredentialsOpts struct {
	ClientID     string
	ClientSecret string
	Scope        string
}

func (a *Auth) GrantClientCredentials(opts GrantClientCredentialsOpts) (*Token, error) {
	if a.server == nil {
		return nil, &InvalidRequest{
			message: "use auth.SetServer() or auth.Discovery() to set the server",
		}
	}

	tokenEndpoint := a.server.TokenEndpoint

	form := url.Values{
		"grant_type":    {"client_credentials"},
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
	err = json.Unmarshal(body, &token)
	if err != nil {
		return nil, err
	}

	if len(token.Error) > 0 {
		if token.Error == "invalid_client" {
			return nil, &InvalidClientError{
				message: token.ErrorDescription,
			}
		}

		return nil, fmt.Errorf("failed to grant client credentials: %v", token.Error)
	}

	return &token, nil
}
