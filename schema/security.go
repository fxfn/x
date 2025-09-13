package schema

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityScheme represents the interface that all security schemes must implement
// This defines the contract between the framework and security implementations
type SecurityScheme interface {
	// GetSecurityScheme returns the name and OpenAPI specification for this security scheme
	GetSecurityScheme() (name string, spec map[string]interface{})

	// Middleware returns the gin.HandlerFunc that implements the security logic
	Middleware() gin.HandlerFunc
}

// Global registry to track security middleware used in routes
var securitySchemeRegistry = make(map[string][]SecurityScheme)

// Registry to map gin.HandlerFunc to their SecurityScheme origin
var middlewareRegistry = make(map[uintptr]SecurityScheme)

// RegisterSecurityScheme registers security schemes for a route
func RegisterSecurityScheme(method, path string, schemes ...SecurityScheme) {
	key := method + " " + path
	securitySchemeRegistry[key] = append(securitySchemeRegistry[key], schemes...)
}

// GetSecuritySchemes retrieves security schemes for a route
func GetSecuritySchemes(method, path string) []SecurityScheme {
	key := method + " " + path
	return securitySchemeRegistry[key]
}

// ClearSecuritySchemes clears all registered security schemes (useful for testing)
func ClearSecuritySchemes() {
	securitySchemeRegistry = make(map[string][]SecurityScheme)
	middlewareRegistry = make(map[uintptr]SecurityScheme)
}

// RegisterSecurityMiddleware registers a gin.HandlerFunc as originating from a SecurityScheme
func RegisterSecurityMiddleware(handler gin.HandlerFunc, scheme SecurityScheme) {
	// Get the function pointer using reflection
	handlerValue := reflect.ValueOf(handler)
	if handlerValue.Kind() == reflect.Func {
		ptr := handlerValue.Pointer()
		middlewareRegistry[ptr] = scheme
	}
}

// IsSecurityMiddleware checks if a gin.HandlerFunc was created by a SecurityScheme
func IsSecurityMiddleware(handler gin.HandlerFunc) (SecurityScheme, bool) {
	handlerValue := reflect.ValueOf(handler)
	if handlerValue.Kind() == reflect.Func {
		ptr := handlerValue.Pointer()
		if scheme, exists := middlewareRegistry[ptr]; exists {
			return scheme, true
		}
	}
	return nil, false
}

type APIKeyLocation string

const (
	APIKeyLocationHeader APIKeyLocation = "header"
	APIKeyLocationQuery  APIKeyLocation = "query"
	APIKeyLocationCookie APIKeyLocation = "cookie"
)

// APIKeyConfig holds configuration for API key security schemes
type APIKeyConfig struct {
	Name        string                                   // Name for OpenAPI documentation (e.g., "ApiKeyAuth")
	Description string                                   // Description for OpenAPI documentation (optional)
	In          APIKeyLocation                           // Location: "header", "query", or "cookie"
	KeyName     string                                   // The name of the header, query parameter, or cookie
	ValidateKey func(c *gin.Context, apiKey string) bool // Function to validate the API key
}

// BearerConfig holds configuration for Bearer token security schemes
type BearerConfig struct {
	Name          string                                  // Name for OpenAPI documentation (e.g., "BearerAuth")
	Description   string                                  // Description for OpenAPI documentation (optional)
	BearerFormat  string                                  // Bearer format (e.g., "JWT") (optional)
	ValidateToken func(c *gin.Context, token string) bool // Function to validate the bearer token
}

// APIKeySecurity implements API key authentication
// This is the standard framework implementation
type APIKeySecurity struct {
	Name        string                                   // Name for OpenAPI documentation (e.g., "ApiKeyAuth")
	Description string                                   // Description for OpenAPI documentation
	In          APIKeyLocation                           // "header", "query", or "cookie"
	KeyName     string                                   // The name of the header, query parameter, or cookie
	ValidateKey func(c *gin.Context, apiKey string) bool // Function to validate the API key
}

// BearerSecurity implements Bearer token authentication
// This is the standard framework implementation
type BearerSecurity struct {
	Name          string                                  // Name for OpenAPI documentation (e.g., "BearerAuth")
	Description   string                                  // Description for OpenAPI documentation
	BearerFormat  string                                  // Bearer format (e.g., "JWT")
	ValidateToken func(c *gin.Context, token string) bool // Function to validate the bearer token
}

// GetSecurityScheme returns the OpenAPI security scheme definition
func (a *APIKeySecurity) GetSecurityScheme() (string, map[string]interface{}) {
	spec := map[string]interface{}{
		"type": "apiKey",
		"in":   a.In,
		"name": a.KeyName,
	}

	if a.Description != "" {
		spec["description"] = a.Description
	}

	return a.Name, spec
}

// Middleware returns the gin.HandlerFunc for API key authentication
func (a *APIKeySecurity) Middleware() gin.HandlerFunc {
	handler := func(c *gin.Context) {
		var apiKey string

		switch a.In {
		case "header":
			apiKey = c.GetHeader(a.KeyName)
		case "query":
			apiKey = c.Query(a.KeyName)
		case "cookie":
			apiKey, _ = c.Cookie(a.KeyName)
		default:
			c.JSON(500, ErrorResult{
				Success: false,
				ErrorInfo: Error{
					Code:    "INTERNAL_ERROR",
					Message: "Invalid API key location configuration",
				},
				Data: nil,
			})
			c.Abort()
			return
		}

		if apiKey == "" {
			c.JSON(401, ErrorResult{
				Success: false,
				ErrorInfo: Error{
					Code:    "UNAUTHORIZED",
					Message: "API key required",
				},
				Data: nil,
			})
			c.Abort()
			return
		}

		// Validate the API key
		if a.ValidateKey != nil && !a.ValidateKey(c, apiKey) {
			c.JSON(401, ErrorResult{
				Success: false,
				ErrorInfo: Error{
					Code:    "UNAUTHORIZED",
					Message: "Invalid API key",
				},
				Data: nil,
			})
			c.Abort()
			return
		}

		// Store API key for handler use
		c.Set("api_key", apiKey)
		c.Next()
	}

	// Register this handler with the security scheme
	RegisterSecurityMiddleware(handler, a)
	return handler
}

// GetSecurityScheme returns the OpenAPI security scheme definition
func (b *BearerSecurity) GetSecurityScheme() (string, map[string]interface{}) {
	spec := map[string]interface{}{
		"type":   "http",
		"scheme": "bearer",
	}

	if b.Description != "" {
		spec["description"] = b.Description
	}

	if b.BearerFormat != "" {
		spec["bearerFormat"] = b.BearerFormat
	}

	return b.Name, spec
}

// Middleware returns the gin.HandlerFunc for Bearer token authentication
func (b *BearerSecurity) Middleware() gin.HandlerFunc {
	handler := func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, ErrorResult{
				Success: false,
				ErrorInfo: Error{
					Code:    "UNAUTHORIZED",
					Message: "Authorization header required",
				},
				Data: nil,
			})
			c.Abort()
			return
		}

		// Check for Bearer prefix
		if len(authHeader) < 7 || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.JSON(401, ErrorResult{
				Success: false,
				ErrorInfo: Error{
					Code:    "UNAUTHORIZED",
					Message: "Invalid authorization header format",
				},
				Data: nil,
			})
			c.Abort()
			return
		}

		token := authHeader[7:]
		if token == "" {
			c.JSON(401, ErrorResult{
				Success: false,
				ErrorInfo: Error{
					Code:    "UNAUTHORIZED",
					Message: "Bearer token required",
				},
				Data: nil,
			})
			c.Abort()
			return
		}

		// Validate the token
		if b.ValidateToken != nil && !b.ValidateToken(c, token) {
			c.JSON(401, ErrorResult{
				Success: false,
				ErrorInfo: Error{
					Code:    "UNAUTHORIZED",
					Message: "Invalid bearer token",
				},
				Data: nil,
			})
			c.Abort()
			return
		}

		// Store token for handler use
		c.Set("bearer_token", token)
		c.Next()
	}

	// Register this handler with the security scheme
	RegisterSecurityMiddleware(handler, b)
	return handler
}

// NewAPIKeySecurity creates a new API key security scheme
func NewAPIKeySecurity(config APIKeyConfig) *APIKeySecurity {
	return &APIKeySecurity{
		Name:        config.Name,
		Description: config.Description,
		In:          config.In,
		KeyName:     config.KeyName,
		ValidateKey: config.ValidateKey,
	}
}

// NewBearerSecurity creates a new Bearer token security scheme
func NewBearerSecurity(config BearerConfig) *BearerSecurity {
	return &BearerSecurity{
		Name:          config.Name,
		Description:   config.Description,
		BearerFormat:  config.BearerFormat,
		ValidateToken: config.ValidateToken,
	}
}

// MultiSecurity implements OR logic for multiple authentication schemes
// A request is valid if ANY of the provided schemes validate successfully
type MultiSecurity struct {
	Name    string           // Name for OpenAPI documentation
	Schemes []SecurityScheme // List of security schemes to try
}

// GetSecurityScheme returns the OpenAPI security scheme definition for multi-auth
// Note: MultiSecurity doesn't register itself, it registers its component schemes
func (m *MultiSecurity) GetSecurityScheme() (string, map[string]interface{}) {
	// MultiSecurity is special - it doesn't register itself as a scheme
	// Instead, it should register each of its component schemes
	// This method should not be called for OpenAPI generation
	return m.Name, map[string]interface{}{
		"type":        "http", // This is a placeholder - shouldn't be used
		"description": "Multiple authentication methods accepted (any one will work)",
	}
}

// GetComponentSchemes returns the individual security schemes for OpenAPI registration
func (m *MultiSecurity) GetComponentSchemes() []SecurityScheme {
	return m.Schemes
}

// Middleware returns a gin.HandlerFunc that tries each security scheme in order
func (m *MultiSecurity) Middleware() gin.HandlerFunc {
	handler := func(c *gin.Context) {
		// Try each security scheme in order
		for _, scheme := range m.Schemes {
			// Try this scheme's middleware directly on the context
			success := m.tryScheme(scheme, c)
			if success {
				c.Next()
				return
			}
		}

		// None of the schemes worked
		c.JSON(401, ErrorResult{
			Success: false,
			ErrorInfo: Error{
				Code:    "UNAUTHORIZED",
				Message: "Valid authentication required (API key, bearer token, etc.)",
			},
			Data: nil,
		})
		c.Abort()
	}

	// Register this handler with the multi-security scheme
	RegisterSecurityMiddleware(handler, m)
	return handler
}

// tryScheme attempts to validate a request using a specific security scheme
func (m *MultiSecurity) tryScheme(scheme SecurityScheme, c *gin.Context) bool {
	switch s := scheme.(type) {
	case *APIKeySecurity:
		return m.tryAPIKey(s, c)
	case *BearerSecurity:
		return m.tryBearer(s, c)
	default:
		// For custom security schemes, we'd need a different approach
		// For now, return false for unknown types
		return false
	}
}

// tryAPIKey attempts API key authentication
func (m *MultiSecurity) tryAPIKey(apiKey *APIKeySecurity, c *gin.Context) bool {
	var key string

	switch apiKey.In {
	case APIKeyLocationHeader:
		key = c.GetHeader(apiKey.KeyName)
	case APIKeyLocationQuery:
		key = c.Query(apiKey.KeyName)
	case APIKeyLocationCookie:
		key, _ = c.Cookie(apiKey.KeyName)
	default:
		return false
	}

	if key == "" {
		return false
	}

	if apiKey.ValidateKey != nil && !apiKey.ValidateKey(c, key) {
		return false
	}

	// Store the API key for handler use
	c.Set("api_key", key)
	c.Set("auth_method", "api_key")
	return true
}

// tryBearer attempts Bearer token authentication
func (m *MultiSecurity) tryBearer(bearer *BearerSecurity, c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	if len(authHeader) < 7 || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return false
	}

	token := authHeader[7:]
	if token == "" {
		return false
	}

	if bearer.ValidateToken != nil && !bearer.ValidateToken(c, token) {
		return false
	}

	// Store the token for handler use
	c.Set("bearer_token", token)
	c.Set("auth_method", "bearer")
	return true
}

// NewMultiSecurity creates a new multi-authentication security scheme
func NewMultiSecurity(name string, schemes ...SecurityScheme) *MultiSecurity {
	return &MultiSecurity{
		Name:    name,
		Schemes: schemes,
	}
}
