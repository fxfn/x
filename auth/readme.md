# Auth Package

A Go package for OAuth 2.0 and OpenID Connect authentication flows, providing a simple and flexible client for interacting with OAuth 2.0 authorization servers.

## Features

- **OpenID Connect Discovery**: Automatic server configuration discovery via `.well-known/openid-configuration`
- **OAuth 2.0 Grant Types**:
  - Client Credentials Grant
  - Resource Owner Password Credentials Grant
- **Token Introspection**: RFC 7662 compliant token introspection with generic response support
- **Custom Error Handling**: Structured error types for better error handling
- **Flexible Configuration**: Manual endpoint configuration or automatic discovery

## Installation

```bash
go get github.com/fxfn/x/auth
```

## Quick Start

### Using OpenID Connect Discovery

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/fxfn/x/auth"
)

func main() {
    // Automatic discovery from OpenID Connect provider
    client, err := auth.Discovery("https://your-auth-server.com")
    if err != nil {
        log.Fatal(err)
    }
    
    // Get access token using client credentials
    token, err := client.GrantClientCredentials(auth.GrantClientCredentialsOpts{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        Scope:        "read write",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Access Token: %s\n", token.AccessToken)
}
```

### Manual Configuration

```go
package main

import (
    "github.com/fxfn/x/auth"
)

func main() {
    client := auth.Default()
    
    // Set server configuration manually
    client.SetServer(&auth.Server{
        TokenEndpoint:         "https://auth.example.com/token",
        IntrospectionEndpoint: "https://auth.example.com/introspect",
        UserinfoEndpoint:      "https://auth.example.com/userinfo",
    })
    
    // Or use SetEndpoint for partial configuration
    client.SetEndpoint(&auth.SetEndpointOpts{
        TokenEndpoint:         "https://auth.example.com/token",
        IntrospectionEndpoint: "https://auth.example.com/introspect",
    })
}
```

## Grant Types

### Client Credentials Grant

The Client Credentials grant is used when applications request an access token to access their own resources, not on behalf of a user.

```go
token, err := client.GrantClientCredentials(auth.GrantClientCredentialsOpts{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    Scope:        "api:read api:write", // Optional
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Token: %s, Expires in: %d seconds\n", 
    token.AccessToken, token.ExpiresIn)
```

### Resource Owner Password Credentials Grant

**Note**: This grant type is generally discouraged for security reasons and should only be used when other flows are not viable.

```go
token, err := client.GrantPassword(auth.GrantPasswordOpts{
    Username:     "user@example.com",
    Password:     "user-password",
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    Scope:        "openid profile", // Optional
})
if err != nil {
    log.Fatal(err)
}
```

## Token Introspection

Validate and get information about access tokens using RFC 7662 token introspection.

### Basic Introspection

```go
response, err := client.Introspect(auth.IntrospectOpts{
    Token:        "access-token-to-validate",
    ClientId:     "your-client-id",
    ClientSecret: "your-client-secret",
})
if err != nil {
    log.Fatal(err)
}

if response.Active {
    fmt.Println("Token is active")
} else {
    fmt.Println("Token is inactive or invalid")
}
```

### Generic Introspection with Custom Response

For providers that return additional fields in introspection responses:

```go
type Customer struct {
  ID   int `json:"id"`
  Name string `json:"name"`
}

type CustomIntrospectResponse struct {
    Active   bool     `json:"active"`
    Customer *Customer `json:"customer"`
}

response, err := auth.IntrospectGeneric[CustomIntrospectResponse](client, auth.IntrospectOpts{
    Token:        "access-token-to-validate",
    ClientId:     "your-client-id",
    ClientSecret: "your-client-secret",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Token active: %t, User: %s, Scope: %s\n", 
    response.Active, response.Username, response.Scope)
```

## Error Handling

The package provides structured error types for better error handling:

```go
token, err := client.GrantClientCredentials(opts)
if err != nil {
    switch e := err.(type) {
    case *auth.InvalidClientError:
        fmt.Printf("Invalid client: %s\n", e.Error())
    case *auth.InvalidRequest:
        fmt.Printf("Invalid request: %s\n", e.Error())
    default:
        fmt.Printf("Other error: %s\n", err.Error())
    }
}
```

### Error Types

- `InvalidClientError`: Returned when client authentication fails
- `InvalidRequest`: Returned for malformed requests or missing required parameters

## API Reference

### Types

#### Auth
The main client struct for OAuth operations.

#### Token
Represents an OAuth 2.0 access token response:
```go
type Token struct {
    AccessToken  string `json:"access_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"`
    RefreshToken string `json:"refresh_token"`
    Scope        string `json:"scope"`
    IdToken      string `json:"id_token"`
}
```

#### Server
Contains OAuth 2.0 server metadata from OpenID Connect discovery:
```go
type Server struct {
    AuthorizationEndpoint    string   `json:"authorization_endpoint"`
    TokenEndpoint           string   `json:"token_endpoint"`
    UserinfoEndpoint        string   `json:"userinfo_endpoint"`
    IntrospectionEndpoint   string   `json:"introspection_endpoint"`
    JwksUri                 string   `json:"jwks_uri"`
    Issuer                  string   `json:"issuer"`
    GrantTypesSupported     []string `json:"grant_types_supported"`
    // ... additional fields
}
```

### Functions

#### `Default() *Auth`
Creates a new Auth client with default configuration.

#### `Discovery(endpoint string) (*Auth, error)`
Creates an Auth client using OpenID Connect discovery. Automatically appends `.well-known/openid-configuration` if not present.

#### `NewServer(metadata map[string]any) (*Server, error)`
Creates a Server instance from a metadata map.

### Methods

#### `SetServer(server *Server)`
Manually sets the OAuth server configuration.

#### `SetEndpoint(opts *SetEndpointOpts)`
Sets specific endpoints while preserving existing configuration.

#### `GrantClientCredentials(opts GrantClientCredentialsOpts) (*Token, error)`
Performs OAuth 2.0 Client Credentials grant.

#### `GrantPassword(opts GrantPasswordOpts) (*Token, error)`
Performs OAuth 2.0 Resource Owner Password Credentials grant.

#### `Introspect(opts IntrospectOpts) (*IntrospectResponse, error)`
Introspects a token using RFC 7662.

## Testing

The package includes comprehensive tests. To run tests with real OAuth servers, set the following environment variables:

```bash
export CLIENT_ID="your-test-client-id"
export CLIENT_SECRET="your-test-client-secret"
export AUTH_ENDPOINT="https://your-auth-server.com"
export USERNAME="test-username"      # For password grant tests
export PASSWORD="test-password"      # For password grant tests

go test ./...
```

Tests will be skipped if environment variables are not set.

## Security Considerations

1. **Client Credentials**: Store client secrets securely and never expose them in client-side code
2. **Password Grant**: Avoid using the password grant type when possible; prefer authorization code flow for user authentication
3. **Token Storage**: Store access tokens securely and implement proper token refresh logic
4. **HTTPS**: Always use HTTPS in production environments
5. **Token Validation**: Always validate tokens on the resource server side using introspection or JWT validation

## License

This package is part of the fxfn/x project. See the project's license for details.
