package main

import (
	"fmt"

	"github.com/fxfn/x/inject"
)

type AuthProvider interface {
	Type() string
}

type authProvider struct{}

type GoogleAuthProvider struct {
	AuthProvider
}

func (g *GoogleAuthProvider) Type() string {
	return "google"
}

type GithubAuthProvider struct {
	AuthProvider
}

func (g *GithubAuthProvider) Type() string {
	return "github"
}

func NewGoogleAuthProvider(c *inject.Container) AuthProvider {
	return &GoogleAuthProvider{}
}

func NewGithubAuthProvider(c *inject.Container) AuthProvider {
	return &GithubAuthProvider{}
}

func main() {
	container := inject.NewContainer()

	inject.RegisterNamed[AuthProvider](container, authProvider{}, NewGoogleAuthProvider)
	inject.RegisterNamed[AuthProvider](container, authProvider{}, NewGithubAuthProvider)

	allProviders := inject.GetAllNamed[AuthProvider](container, authProvider{})
	fmt.Printf("Found %d providers\n", len(allProviders))
	for _, provider := range allProviders {
		fmt.Printf("Provider: %s\n", provider.Type())
	}
}
