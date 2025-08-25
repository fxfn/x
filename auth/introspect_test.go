package auth

import (
	"os"
	"testing"
)

func TestIntrospect(t *testing.T) {
	t.Run("should return an error if no endpoint is set", func(t *testing.T) {
		auth := Default()
		introspectResponse, err := auth.Introspect(IntrospectOpts{
			Token: "test",
		})

		if err == nil {
			t.Errorf("expected error, got nil")
		}

		if introspectResponse != nil {
			t.Errorf("expected nil, got %v", introspectResponse)
		}
	})

	t.Run("should return an introspection token", func(t *testing.T) {

		clientId := os.Getenv("CLIENT_ID")
		clientSecret := os.Getenv("CLIENT_SECRET")
		token := os.Getenv("TOKEN")
		authEndpoint := os.Getenv("AUTH_ENDPOINT")

		if clientId == "" || clientSecret == "" || token == "" || authEndpoint == "" {
			t.Skip("CLIENT_ID, CLIENT_SECRET, TOKEN, and AUTH_ENDPOINT must be set")
		}

		auth, err := Discovery(authEndpoint)
		if err != nil {
			t.Fatalf("failed to discover auth: %v", err)
		}

		introspectResponse, err := auth.Introspect(IntrospectOpts{
			Token:        token,
			ClientId:     clientId,
			ClientSecret: clientSecret,
		})

		if err != nil {
			t.Fatalf("failed to introspect: %v", err)
		}

		if introspectResponse == nil {
			t.Fatalf("introspect response is nil")
		}

		if introspectResponse.Active == false {
			t.Fatalf("introspect response is not active")
		}
	})
}
