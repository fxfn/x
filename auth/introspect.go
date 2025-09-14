package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type IntrospectOpts struct {
	Token        string
	ClientId     string
	ClientSecret string
}

type IntrospectResponse struct {
	Active    bool   `json:"active"`
	ClientID  string `json:"client_id"`
	Username  string `json:"username"`
	Scope     string `json:"scope"`
	Subject   string `json:"sub"`
	Audience  string `json:"aud"`
	Issuer    string `json:"iss"`
	ExpiresAt int    `json:"exp"`
	IssuedAt  int    `json:"iat"`
	TokenType string `json:"token_type"`
	NotBefore int    `json:"nbf"`
	TokenID   string `json:"jti"`
}

func (a *Auth) Introspect(opts IntrospectOpts) (*IntrospectResponse, error) {
	return IntrospectGeneric[IntrospectResponse](a, opts)
}

func IntrospectGeneric[T any](a *Auth, opts IntrospectOpts) (*T, error) {

	if a.server == nil {
		return nil, errors.New("no server set")
	}

	if a.server.IntrospectionEndpoint == "" {
		return nil, errors.New("no introspection endpoint set")
	}

	u, err := url.Parse(a.server.IntrospectionEndpoint)
	if err != nil {
		return nil, err
	}

	values := url.Values{
		"token": {opts.Token},
	}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", opts.ClientId, opts.ClientSecret)))))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var introspectResponse T
	json.Unmarshal(body, &introspectResponse)

	return &introspectResponse, nil
}
