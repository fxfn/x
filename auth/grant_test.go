package auth

import (
	"os"
	"slices"
	"testing"
)

func TestGrantClientCredentials(t *testing.T) {
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	authEndpoint := os.Getenv("AUTH_ENDPOINT")

	if clientId == "" || clientSecret == "" || authEndpoint == "" {
		t.Skip("CLIENT_ID, CLIENT_SECRET, and AUTH_ENDPOINT must be set")
	}

	auth, err := Discovery(authEndpoint)

	if err != nil {
		t.Fatalf("failed to discover auth: %v", err)
	}

	token, err := auth.GrantClientCredentials(GrantClientCredentialsOpts{
		ClientID:     clientId,
		ClientSecret: clientSecret,
	})

	if err != nil {
		t.Fatalf("failed to grant client credentials: %v", err)
	}

	if token == nil {
		t.Fatalf("token is nil")
	}

	if token.AccessToken == "" {
		t.Fatalf("access token is empty")
	}

	if token.TokenType == "" {
		t.Fatalf("token type is empty")
	}
}

func TestGrantPassword(t *testing.T) {
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	authEndpoint := os.Getenv("AUTH_ENDPOINT")
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")

	if clientId == "" || clientSecret == "" || authEndpoint == "" || username == "" || password == "" {
		t.Skip("CLIENT_ID, CLIENT_SECRET, AUTH_ENDPOINT, USERNAME, and PASSWORD must be set")
	}

	auth, err := Discovery(authEndpoint)

	if !slices.Contains(auth.server.GrantTypesSupported, "password") {
		t.Skip("password grant type is not supported by server")
	}

	if err != nil {
		t.Fatalf("failed to discover auth: %v", err)
	}

	token, err := auth.GrantPassword(GrantPasswordOpts{
		Username:     username,
		Password:     password,
		ClientID:     clientId,
		ClientSecret: clientSecret,
	})

	if err != nil {
		t.Fatalf("failed to grant password: %v", err)
	}

	if token == nil {
		t.Fatalf("token is nil")
	}

	if token.AccessToken == "" {
		t.Fatalf("access token is empty")
	}

	if token.TokenType == "" {
		t.Fatalf("token type is empty")
	}
}
