package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func Default() *Auth {
	return &Auth{}
}

func (a *Auth) SetServer(server *Server) {
	a.server = server
}

type SetEndpointOpts struct {
	TokenEndpoint               string
	UserinfoEndpoint            string
	IntrospectionEndpoint       string
	RevocationEndpoint          string
	EndSessionEndpoint          string
	AuthorizationEndpoint       string
	DeviceAuthorizationEndpoint string
}

func (a *Auth) SetEndpoint(opts *SetEndpointOpts) {
	if opts.TokenEndpoint != "" {
		a.server.TokenEndpoint = opts.TokenEndpoint
	}

	if opts.UserinfoEndpoint != "" {
		a.server.UserinfoEndpoint = opts.UserinfoEndpoint
	}

	if opts.IntrospectionEndpoint != "" {
		a.server.IntrospectionEndpoint = opts.IntrospectionEndpoint
	}

	if opts.RevocationEndpoint != "" {
		a.server.RevocationEndpoint = opts.RevocationEndpoint
	}

	if opts.EndSessionEndpoint != "" {
		a.server.EndSessionEndpoint = opts.EndSessionEndpoint
	}

	if opts.AuthorizationEndpoint != "" {
		a.server.AuthorizationEndpoint = opts.AuthorizationEndpoint
	}

	if opts.DeviceAuthorizationEndpoint != "" {
		a.server.DeviceAuthorizationEndpoint = opts.DeviceAuthorizationEndpoint
	}
}

func Discovery(endpoint string) (*Auth, error) {
	if !strings.HasSuffix(endpoint, ".well-known/openid-configuration") {
		endpoint = fmt.Sprintf("%s/.well-known/openid-configuration", endpoint)
	}

	serverMetadata, err := fetchServerMetadata(endpoint)
	if err != nil {
		return nil, err
	}

	return &Auth{
		endpoint: endpoint,
		server:   serverMetadata,
	}, nil
}

func fetchServerMetadata(endpoint string) (*Server, error) {

	res, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// unmarshal json
	var serverMetadata Server
	err = json.Unmarshal(body, &serverMetadata)
	if err != nil {
		return nil, err
	}

	return &serverMetadata, nil
}
