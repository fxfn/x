package auth

import (
	"testing"
)

func TestSetEndpoint(t *testing.T) {
	auth := Default()

	auth.SetServer(&Server{
		TokenEndpoint: "https://auth.shipeedo.com/token",
	})

	if auth.server == nil {
		t.Fatalf("server is nil")
	}

	if auth.server.TokenEndpoint != "https://auth.shipeedo.com/token" {
		t.Fatalf("token endpoint is not set")
	}
}

func TestDiscovery(t *testing.T) {
	t.Run("With .well-known/openid-configuration", func(t *testing.T) {
		auth, err := Discovery("https://auth.shipeedo.com/.well-known/openid-configuration")

		if err != nil {
			t.Fatalf("failed to discover auth: %v", err)
		}

		if auth.server == nil {
			t.Fatalf("server metadata is nil")
		}
	})

	t.Run("Without .well-known/openid-configuration", func(t *testing.T) {
		auth, err := Discovery("https://auth.shipeedo.com")

		if err != nil {
			t.Fatalf("failed to discover auth: %v", err)
		}

		if auth.server == nil {
			t.Fatalf("server metadata is nil")
		}
	})
}

func TestFetchServerMetadata(t *testing.T) {
	metadata, err := fetchServerMetadata("https://auth.shipeedo.com/.well-known/openid-configuration")
	if err != nil {
		t.Fatalf("failed to fetch server metadata: %v", err)
	}

	if metadata == nil {
		t.Fatalf("server metadata is nil")
	}
}
