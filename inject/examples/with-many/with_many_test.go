package main

import (
	"testing"

	"github.com/fxfn/x/inject"
)

func TestWithMany(t *testing.T) {
	container := inject.NewContainer()
	inject.RegisterNamed[AuthProvider](container, authProvider{}, NewGoogleAuthProvider)
	inject.RegisterNamed[AuthProvider](container, authProvider{}, NewGithubAuthProvider)

	allProviders := inject.GetAllNamed[AuthProvider](container, authProvider{})
	t.Run("should have 2 providers", func(t *testing.T) {
		if len(allProviders) != 2 {
			t.Errorf("expected 2 providers, got %d", len(allProviders))
		}
	})

	t.Run("should have a google provider", func(t *testing.T) {
		if allProviders[0].Type() != "google" {
			t.Errorf("expected google provider, got %s", allProviders[0].Type())
		}
	})

	t.Run("should have a github provider", func(t *testing.T) {
		if allProviders[1].Type() != "github" {
			t.Errorf("expected github provider, got %s", allProviders[1].Type())
		}
	})
}
