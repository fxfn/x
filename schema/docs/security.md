# Security

Comprehensive authentication and authorization middleware system with automatic OpenAPI documentation.

## Overview

The security system provides:
- Built-in API Key and Bearer Token authentication
- Multi-authentication support (OR logic)
- Automatic OpenAPI security scheme generation
- Custom security scheme extensibility
- Reflection-based middleware detection

## Benefits

- **Multiple Auth Methods**: API Key, Bearer Token, and custom schemes
- **Flexible Application**: Per-route, per-group, or global security
- **OR Logic**: Accept any of multiple authentication methods
- **Auto Documentation**: OpenAPI security schemes generated automatically
- **Type Safety**: Compile-time guarantees for security configuration

## Built-in Security Schemes

### API Key Authentication

Supports API keys in headers, query parameters, or cookies.

```go
apiKeySecurity := schema.NewAPIKeySecurity(schema.APIKeyConfig{
    Name:        "ApiKeyAuth",
    Description: "API key authentication",
    In:          schema.APIKeyLocationHeader,
    KeyName:     "X-API-Key",
    ValidateKey: func(key string) bool {
        return key == "my-secret-key"
    },
})
```

### Bearer Token Authentication

Standard Bearer token authentication (typically JWT).

```go
bearerSecurity := schema.NewBearerSecurity(schema.BearerConfig{
    Name:         "BearerAuth",
    Description:  "JWT Bearer token authentication",
    BearerFormat: "JWT",
    ValidateToken: func(token string) bool {
        return validateJWTToken(token)
    },
})
```

### Multi-Authentication

Accept any of multiple authentication methods.

```go
multiAuth := schema.NewMultiSecurity("MultiAuth", apiKeySecurity, bearerSecurity)
```

## API Reference

### Types

#### `APIKeyLocation`
```go
type APIKeyLocation string

const (
    APIKeyLocationHeader APIKeyLocation = "header"
    APIKeyLocationQuery  APIKeyLocation = "query" 
    APIKeyLocationCookie APIKeyLocation = "cookie"
)
```

#### `APIKeyConfig`
```go
type APIKeyConfig struct {
    Name        string                   // OpenAPI scheme name
    Description string                   // Optional description
    In          APIKeyLocation           // Location of the key
    KeyName     string                   // Parameter/header name
    ValidateKey func(apiKey string) bool // Validation function
}
```

#### `BearerConfig`
```go
type BearerConfig struct {
    Name          string                  // OpenAPI scheme name
    Description   string                  // Optional description
    BearerFormat  string                  // Bearer format (e.g., "JWT")
    ValidateToken func(token string) bool // Validation function
}
```

### Functions

#### `NewAPIKeySecurity(config APIKeyConfig) *APIKeySecurity`
Creates a new API key security scheme.

#### `NewBearerSecurity(config BearerConfig) *BearerSecurity`
Creates a new Bearer token security scheme.

#### `NewMultiSecurity(name string, schemes ...SecurityScheme) *MultiSecurity`
Creates a multi-authentication scheme that accepts any of the provided schemes.

## Examples

### API Key in Header
```go
apiKey := schema.NewAPIKeySecurity(schema.APIKeyConfig{
    Name:        "ApiKeyAuth",
    Description: "API key in X-API-Key header",
    In:          schema.APIKeyLocationHeader,
    KeyName:     "X-API-Key",
    ValidateKey: func(key string) bool {
        return key == "secret-key"
    },
})

// Apply to specific routes
router.GET("/protected", apiKey, handlers.GetProtectedData())

// Apply to route group
protected := router.Group("/api/v1")
protected.Use(apiKey.Middleware())
{
    protected.GET("/users", handlers.GetUsers())
    protected.POST("/users", handlers.CreateUser())
}
```

**Usage:**
```bash
curl -H "X-API-Key: secret-key" http://localhost:8080/protected
```

### API Key in Query Parameter
```go
queryAuth := schema.NewAPIKeySecurity(schema.APIKeyConfig{
    Name:        "QueryAuth",
    Description: "API key in query parameter",
    In:          schema.APIKeyLocationQuery,
    KeyName:     "api_key",
    ValidateKey: func(key string) bool {
        return isValidAPIKey(key)
    },
})

router.GET("/data", queryAuth, handlers.GetData())
```

**Usage:**
```bash
curl "http://localhost:8080/data?api_key=secret-key"
```

### Bearer Token (JWT)
```go
jwtAuth := schema.NewBearerSecurity(schema.BearerConfig{
    Name:         "JWTAuth",
    Description:  "JWT Bearer token",
    BearerFormat: "JWT",
    ValidateToken: func(token string) bool {
        // Parse and validate JWT
        claims, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })
        
        if err != nil {
            return false
        }
        
        return claims.Valid() == nil
    },
})

router.GET("/profile", jwtAuth, handlers.GetProfile())
```

**Usage:**
```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." http://localhost:8080/profile
```

### Multi-Authentication (OR Logic)
```go
// Accept either API key OR Bearer token
multiAuth := schema.NewMultiSecurity("FlexibleAuth", apiKeySecurity, bearerSecurity)

// Apply to route group
api := router.Group("/api/v1")
api.Use(multiAuth.Middleware())
{
    api.GET("/users", handlers.GetUsers())
    api.POST("/users", handlers.CreateUser())
}
```

**Usage (either works):**
```bash
# Using API key
curl -H "X-API-Key: secret-key" http://localhost:8080/api/v1/users

# Using Bearer token
curl -H "Authorization: Bearer eyJhbGci..." http://localhost:8080/api/v1/users
```

### Database-Backed API Key Validation
```go
dbAuth := schema.NewAPIKeySecurity(schema.APIKeyConfig{
    Name:        "DatabaseAuth",
    Description: "Database-validated API key",
    In:          schema.APIKeyLocationHeader,
    KeyName:     "X-API-Key",
    ValidateKey: func(key string) bool {
        // Look up API key in database
        apiKey, err := db.GetAPIKey(key)
        if err != nil {
            return false
        }
        
        // Check if key is active and not expired
        return apiKey.Active && apiKey.ExpiresAt.After(time.Now())
    },
})
```

### Role-Based Bearer Token
```go
adminAuth := schema.NewBearerSecurity(schema.BearerConfig{
    Name:         "AdminAuth",
    Description:  "Admin-only JWT authentication",
    BearerFormat: "JWT",
    ValidateToken: func(token string) bool {
        claims, err := parseJWTToken(token)
        if err != nil {
            return false
        }
        
        // Check if user has admin role
        return claims.Role == "admin" && claims.Valid() == nil
    },
})

// Admin-only routes
admin := router.Group("/admin")
admin.Use(adminAuth.Middleware())
{
    admin.GET("/users", handlers.AdminGetUsers())
    admin.DELETE("/users/:id", handlers.AdminDeleteUser())
}
```

### Context Access in Handlers
```go
func GetProfile(c *gin.Context, req ProfileSchema) (ProfileResponse, error) {
    // Check which authentication method was used
    if authMethod, exists := c.Get("auth_method"); exists {
        log.Info("Authentication method", "method", authMethod)
    }
    
    // Access API key if used
    if apiKey, exists := c.Get("api_key"); exists {
        log.Info("API key", "key", apiKey)
    }
    
    // Access bearer token if used
    if token, exists := c.Get("bearer_token"); exists {
        log.Info("Bearer token", "token", token)
    }
    
    // Your handler logic...
}
```

## Custom Security Schemes

### Implementing SecurityScheme Interface
```go
type CustomOAuthSecurity struct {
    Name         string
    ClientID     string
    ClientSecret string
}

func (c *CustomOAuthSecurity) GetSecurityScheme() (string, map[string]interface{}) {
    return c.Name, map[string]interface{}{
        "type": "oauth2",
        "flows": map[string]interface{}{
            "authorizationCode": map[string]interface{}{
                "authorizationUrl": "https://auth.example.com/oauth/authorize",
                "tokenUrl":         "https://auth.example.com/oauth/token",
                "scopes": map[string]string{
                    "read":  "Read access",
                    "write": "Write access",
                },
            },
        },
    }
}

func (c *CustomOAuthSecurity) Middleware() gin.HandlerFunc {
    handler := func(ctx *gin.Context) {
        authHeader := ctx.GetHeader("Authorization")
        if authHeader == "" {
            ctx.JSON(401, schema.ErrorResult{
                Success: false,
                ErrorInfo: schema.Error{
                    Code:    "UNAUTHORIZED",
                    Message: "OAuth token required",
                },
                Data: nil,
            })
            ctx.Abort()
            return
        }
        
        // Validate OAuth token
        if !c.validateOAuthToken(authHeader) {
            ctx.JSON(401, schema.ErrorResult{
                Success: false,
                ErrorInfo: schema.Error{
                    Code:    "INVALID_TOKEN",
                    Message: "Invalid OAuth token",
                },
                Data: nil,
            })
            ctx.Abort()
            return
        }
        
        ctx.Set("oauth_token", authHeader)
        ctx.Next()
    }
    
    // Register for reflection-based detection
    schema.RegisterSecurityMiddleware(handler, c)
    return handler
}

func (c *CustomOAuthSecurity) validateOAuthToken(token string) bool {
    // Custom OAuth validation logic
    return true
}
```

### Session-Based Authentication
```go
type SessionSecurity struct {
    Name       string
    CookieName string
}

func NewSessionSecurity(name, cookieName string) *SessionSecurity {
    return &SessionSecurity{
        Name:       name,
        CookieName: cookieName,
    }
}

func (s *SessionSecurity) GetSecurityScheme() (string, map[string]interface{}) {
    return s.Name, map[string]interface{}{
        "type":        "apiKey",
        "in":          "cookie",
        "name":        s.CookieName,
        "description": "Session-based authentication",
    }
}

func (s *SessionSecurity) Middleware() gin.HandlerFunc {
    handler := func(c *gin.Context) {
        sessionID, err := c.Cookie(s.CookieName)
        if err != nil {
            c.JSON(401, schema.ErrorResult{
                Success: false,
                ErrorInfo: schema.Error{
                    Code:    "NO_SESSION",
                    Message: "Valid session required",
                },
                Data: nil,
            })
            c.Abort()
            return
        }
        
        // Validate session
        session, err := sessionStore.Get(sessionID)
        if err != nil || session.Expired() {
            c.JSON(401, schema.ErrorResult{
                Success: false,
                ErrorInfo: schema.Error{
                    Code:    "INVALID_SESSION",
                    Message: "Session expired or invalid",
                },
                Data: nil,
            })
            c.Abort()
            return
        }
        
        c.Set("session", session)
        c.Set("user_id", session.UserID)
        c.Next()
    }
    
    schema.RegisterSecurityMiddleware(handler, s)
    return handler
}
```

## OpenAPI Integration

Security schemes are automatically documented in OpenAPI specifications:

### Generated securitySchemes
```yaml
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
      description: API key authentication
    
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: JWT Bearer token authentication
```

### Operation Security (Single Auth)
```yaml
paths:
  /protected:
    get:
      security:
        - ApiKeyAuth: []
```

### Operation Security (Multi-Auth OR Logic)
```yaml
paths:
  /flexible:
    get:
      security:
        - ApiKeyAuth: []
        - BearerAuth: []
```

## Error Responses

### Missing Authentication
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "API key required"
  }
}
```

### Invalid Credentials
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid API key"
  }
}
```

### Multi-Auth Failure
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Valid authentication required (API key, bearer token, etc.)"
  }
}
```

## Best Practices

### 1. Use Descriptive Names
```go
// Good
apiKeySecurity := schema.NewAPIKeySecurity(schema.APIKeyConfig{
    Name: "ApiKeyAuth",  // Clear, descriptive name
    ...
})

// Avoid
apiKeySecurity := schema.NewAPIKeySecurity(schema.APIKeyConfig{
    Name: "Auth",  // Too generic
    ...
})
```

### 2. Implement Proper Validation
```go
ValidateKey: func(key string) bool {
    // Check against database
    apiKey, err := db.GetAPIKey(key)
    if err != nil {
        return false
    }
    
    // Verify key is active and not expired
    return apiKey.Active && 
           apiKey.ExpiresAt.After(time.Now()) &&
           !apiKey.Revoked
}
```

### 3. Use Multi-Auth for Flexibility
```go
// Allow multiple authentication methods for better UX
multiAuth := schema.NewMultiSecurity("FlexibleAuth", 
    apiKeySecurity,    // For service-to-service
    bearerSecurity,    // For user sessions
    sessionSecurity,   // For web app
)
```

### 4. Apply Security at Appropriate Level
```go
// Global security (not recommended)
router.Use(authMiddleware.Middleware())

// Group-level security (recommended)
api := router.Group("/api/v1")
api.Use(authMiddleware.Middleware())

// Route-level security (for mixed access)
router.GET("/public", handlers.Public())        // No auth
router.GET("/protected", auth, handlers.Protected()) // With auth
```

### 5. Store Context Information
```go
func (a *APIKeySecurity) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... validation logic ...
        
        // Store useful information for handlers
        c.Set("api_key", apiKey)
        c.Set("auth_method", "api_key")
        c.Set("user_id", getUserIDFromAPIKey(apiKey))
        c.Set("permissions", getPermissions(apiKey))
        
        c.Next()
    }
}
```

## Integration

Security integrates with:
- **[Router](./router.md)**: Automatic middleware detection and registration
- **[OpenAPI](./openapi.md)**: Security scheme documentation generation
- **[Handlers](./handlers.md)**: Context access for authentication info
- **[Results](./results.md)**: Standardized error responses
