# Middleware

Framework and custom middleware integration with automatic detection and OpenAPI support.

## Overview

The Middleware system provides:
- Seamless integration with Gin middleware
- Automatic security middleware detection using reflection
- OpenAPI documentation for security middleware
- Support for both framework and custom middleware
- Type-safe middleware composition

## Benefits

- **Gin Compatibility**: Works with all existing Gin middleware
- **Auto Detection**: Security middleware automatically detected and documented
- **Flexible Composition**: Mix security and regular middleware freely
- **Type Safety**: Compile-time guarantees for middleware setup
- **Zero Configuration**: Security documentation generated automatically

## Middleware Types

### Security Middleware
Middleware that implements the `SecurityScheme` interface and provides authentication/authorization.

### Regular Middleware
Standard Gin middleware for logging, CORS, rate limiting, etc.

### Custom Middleware
Application-specific middleware for business logic.

## API Reference

### Security Middleware Detection

#### `RegisterSecurityMiddleware(handler gin.HandlerFunc, scheme SecurityScheme)`
Registers a gin.HandlerFunc as originating from a SecurityScheme.

#### `IsSecurityMiddleware(handler gin.HandlerFunc) (SecurityScheme, bool)`
Checks if a gin.HandlerFunc was created by a SecurityScheme.

### Router Middleware Methods

#### `router.Use(middleware ...gin.HandlerFunc)`
Adds middleware with automatic security detection.

#### `group.Use(middleware ...gin.HandlerFunc)`
Adds middleware to route group with automatic security detection.

## Examples

### Security Middleware
```go
// Security middleware is automatically detected
apiKey := schema.NewAPIKeySecurity(schema.APIKeyConfig{
    Name:    "ApiKeyAuth",
    In:      schema.APIKeyLocationHeader,
    KeyName: "X-API-Key",
    ValidateKey: func(key string) bool { return key == "secret" },
})

// Apply to route group - automatically detected as security middleware
protected := router.Group("/api/v1")
protected.Use(apiKey.Middleware()) // Auto-detected and documented
{
    protected.GET("/users", schema.ValidateAndHandle(GetUsers))
}
```

### Custom Authentication Middleware
```go
// Create custom security middleware
type CustomAuth struct {
    Name string
}

func (c *CustomAuth) GetSecurityScheme() (string, map[string]interface{}) {
    return c.Name, map[string]interface{}{
        "type":        "http",
        "scheme":      "custom",
        "description": "Custom authentication scheme",
    }
}

func (c *CustomAuth) Middleware() gin.HandlerFunc {
    handler := func(ctx *gin.Context) {
        // Custom authentication logic
        token := ctx.GetHeader("X-Custom-Token")
        if !c.validateToken(token) {
            ctx.JSON(401, schema.ErrorResult{
                Success: false,
                ErrorInfo: schema.Error{
                    Code:    "UNAUTHORIZED",
                    Message: "Invalid custom token",
                },
                Data: nil,
            })
            ctx.Abort()
            return
        }
        
        ctx.Set("custom_token", token)
        ctx.Next()
    }
    
    // Register for automatic detection
    schema.RegisterSecurityMiddleware(handler, c)
    return handler
}

func (c *CustomAuth) validateToken(token string) bool {
    // Custom validation logic
    return token == "valid-custom-token"
}

// Usage
customAuth := &CustomAuth{Name: "CustomAuth"}
group.Use(customAuth.Middleware()) // Automatically detected
```

### Regular Middleware
```go
// Standard Gin middleware
router.Use(gin.Logger())
router.Use(gin.Recovery())

// CORS middleware
func corsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}

router.Use(corsMiddleware())

// Rate limiting middleware
func rateLimitMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Limit(100), 10) // 100 requests per second, burst of 10
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, schema.ErrorResult{
                Success: false,
                ErrorInfo: schema.Error{
                    Code:    "RATE_LIMIT_EXCEEDED",
                    Message: "Too many requests",
                },
                Data: nil,
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

router.Use(rateLimitMiddleware())
```

### Mixed Middleware Application
```go
// Combine security and regular middleware
group := router.Group("/api/v1")
group.Use(
    gin.Logger(),                    // Regular middleware
    corsMiddleware(),                // Custom middleware
    rateLimitMiddleware(),           // Custom middleware
    apiKeySecurity.Middleware(),     // Security middleware (auto-detected)
    requestIDMiddleware(),           // Custom middleware
)
{
    group.GET("/data", schema.ValidateAndHandle(GetData))
}
```

### Request Context Middleware
```go
// Middleware that adds request context
func requestContextMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Generate request ID
        requestID := generateRequestID()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        
        // Add timestamp
        c.Set("request_start", time.Now())
        
        c.Next()
        
        // Log request completion
        duration := time.Since(c.GetTime("request_start"))
        log.Info("Request completed",
            "request_id", requestID,
            "method", c.Request.Method,
            "path", c.Request.URL.Path,
            "status", c.Writer.Status(),
            "duration", duration,
        )
    }
}

// User info middleware (depends on authentication)
func userContextMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get authenticated user info
        if userID, exists := c.Get("user_id"); exists {
            user, err := userService.GetByID(userID.(string))
            if err == nil {
                c.Set("user", user)
                c.Set("user_roles", user.Roles)
                c.Set("user_permissions", user.Permissions)
            }
        }
        
        c.Next()
    }
}

// Apply in order
protected := router.Group("/api/v1")
protected.Use(
    requestContextMiddleware(),  // First: Set up request context
    authMiddleware.Middleware(), // Second: Authenticate (sets user_id)
    userContextMiddleware(),     // Third: Load user data
)
```

### Multi-Authentication Middleware
```go
// Multi-auth automatically handles multiple schemes
apiKey := schema.NewAPIKeySecurity(schema.APIKeyConfig{...})
bearer := schema.NewBearerSecurity(schema.BearerConfig{...})
session := schema.NewSessionSecurity("SessionAuth", "session_id")

// Accept any of the three authentication methods
multiAuth := schema.NewMultiSecurity("FlexibleAuth", apiKey, bearer, session)

flexible := router.Group("/api/v1")
flexible.Use(multiAuth.Middleware()) // All three schemes documented in OpenAPI
{
    flexible.GET("/profile", schema.ValidateAndHandle(GetProfile))
}
```

### Custom Business Logic Middleware
```go
// Tenant isolation middleware
func tenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.GetHeader("X-Tenant-ID")
        if tenantID == "" {
            c.JSON(400, schema.ErrorResult{
                Success: false,
                ErrorInfo: schema.Error{
                    Code:    "MISSING_TENANT",
                    Message: "Tenant ID is required",
                },
                Data: nil,
            })
            c.Abort()
            return
        }
        
        // Validate tenant exists and is active
        tenant, err := tenantService.GetByID(tenantID)
        if err != nil || !tenant.Active {
            c.JSON(403, schema.ErrorResult{
                Success: false,
                ErrorInfo: schema.Error{
                    Code:    "INVALID_TENANT",
                    Message: "Invalid or inactive tenant",
                },
                Data: nil,
            })
            c.Abort()
            return
        }
        
        c.Set("tenant", tenant)
        c.Set("tenant_id", tenantID)
        c.Next()
    }
}

// Feature flag middleware
func featureFlagMiddleware(feature string) gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID, _ := c.Get("tenant_id")
        userID, _ := c.Get("user_id")
        
        if !featureFlagService.IsEnabled(feature, tenantID, userID) {
            c.JSON(404, schema.ErrorResult{
                Success: false,
                ErrorInfo: schema.Error{
                    Code:    "FEATURE_NOT_AVAILABLE",
                    Message: "This feature is not available",
                },
                Data: nil,
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Usage
api := router.Group("/api/v1")
api.Use(
    authMiddleware.Middleware(),              // Authentication
    tenantMiddleware(),                       // Tenant isolation
)
{
    // Feature-gated endpoint
    beta := api.Group("/beta")
    beta.Use(featureFlagMiddleware("beta_features"))
    {
        beta.GET("/new-feature", schema.ValidateAndHandle(GetNewFeature))
    }
}
```

### Error Handling Middleware
```go
// Custom error recovery middleware
func errorRecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // Log the panic
                log.Error("Panic recovered",
                    "error", err,
                    "path", c.Request.URL.Path,
                    "method", c.Request.Method,
                )
                
                // Return standardized error response
                c.JSON(500, schema.ErrorResult{
                    Success: false,
                    ErrorInfo: schema.Error{
                        Code:    "INTERNAL_SERVER_ERROR",
                        Message: "An unexpected error occurred",
                    },
                    Data: nil,
                })
                c.Abort()
            }
        }()
        
        c.Next()
    }
}

// Timeout middleware
func timeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
        defer cancel()
        
        c.Request = c.Request.WithContext(ctx)
        
        finished := make(chan struct{})
        go func() {
            c.Next()
            finished <- struct{}{}
        }()
        
        select {
        case <-finished:
            // Request completed normally
        case <-ctx.Done():
            // Request timed out
            c.JSON(408, schema.ErrorResult{
                Success: false,
                ErrorInfo: schema.Error{
                    Code:    "REQUEST_TIMEOUT",
                    Message: "Request timed out",
                },
                Data: nil,
            })
            c.Abort()
        }
    }
}

// Apply error handling
router.Use(errorRecoveryMiddleware())
router.Use(timeoutMiddleware(30 * time.Second))
```

## Middleware Order and Dependencies

### Typical Middleware Order
```go
router.Use(
    // 1. Infrastructure middleware
    gin.Logger(),
    gin.Recovery(),
    corsMiddleware(),
    
    // 2. Request setup
    requestIDMiddleware(),
    timeoutMiddleware(30 * time.Second),
    
    // 3. Rate limiting
    rateLimitMiddleware(),
    
    // 4. Authentication (if global)
    // authMiddleware.Middleware(),
)

// Group-specific middleware
protected := router.Group("/api/v1")
protected.Use(
    // 5. Authentication
    authMiddleware.Middleware(),
    
    // 6. Authorization/business logic
    tenantMiddleware(),
    permissionMiddleware(),
    
    // 7. Feature flags
    featureFlagMiddleware("api_v1"),
)
```

### Dependency Chain Example
```go
// Each middleware depends on the previous ones
api := router.Group("/api/v1")
api.Use(
    requestIDMiddleware(),      // Sets request_id
    authMiddleware.Middleware(), // Sets user_id (needs request_id for logging)
    userContextMiddleware(),     // Sets user object (needs user_id)
    tenantMiddleware(),          // Sets tenant (needs user for validation)
    permissionMiddleware(),      // Checks permissions (needs user and tenant)
)
{
    api.GET("/data", schema.ValidateAndHandle(GetData))
}
```

## Reflection-Based Security Detection

### How It Works
```go
// 1. Security middleware registers itself during creation
func (a *APIKeySecurity) Middleware() gin.HandlerFunc {
    handler := func(c *gin.Context) { /* ... */ }
    
    // This creates the association
    schema.RegisterSecurityMiddleware(handler, a)
    return handler
}

// 2. Router detects security middleware during application
func (rg *RouterGroup) Use(middleware ...gin.HandlerFunc) gin.IRoutes {
    for _, handler := range middleware {
        // This checks the registry
        if scheme, isSecurityMiddleware := schema.IsSecurityMiddleware(handler); isSecurityMiddleware {
            rg.groupSecuritySchemes = append(rg.groupSecuritySchemes, scheme)
        }
    }
    
    return rg.RouterGroup.Use(middleware...)
}

// 3. OpenAPI generation uses detected schemes
func generateOperation(info HandlerInfo, ...) *Operation {
    // Security schemes are automatically included in documentation
    for _, scheme := range info.SecuritySchemes {
        // Add to OpenAPI spec
    }
}
```

### Custom Security Scheme Registration
```go
type DatabaseAuth struct {
    Name string
}

func (d *DatabaseAuth) Middleware() gin.HandlerFunc {
    handler := func(c *gin.Context) {
        // Authentication logic
    }
    
    // IMPORTANT: Register for automatic detection
    schema.RegisterSecurityMiddleware(handler, d)
    return handler
}

// This will be automatically detected and documented
group.Use(dbAuth.Middleware())
```

## Best Practices

### 1. Apply Middleware in Logical Order
```go
// Good: Logical dependency order
router.Use(
    gin.Logger(),               // Infrastructure
    corsMiddleware(),           // Infrastructure
    rateLimitMiddleware(),      // Rate limiting
    authMiddleware.Middleware(), // Authentication
    permissionMiddleware(),     // Authorization
)

// Avoid: Random order that may break dependencies
router.Use(
    permissionMiddleware(),     // Needs auth context
    rateLimitMiddleware(),      
    authMiddleware.Middleware(), // Should come before permissions
)
```

### 2. Use Appropriate Scope
```go
// Global infrastructure middleware
router.Use(gin.Logger())
router.Use(gin.Recovery())
router.Use(corsMiddleware())

// Group-specific business logic
protected := router.Group("/api/v1")
protected.Use(authMiddleware.Middleware())

// Route-specific features
beta := protected.Group("/beta")
beta.Use(featureFlagMiddleware("beta"))
```

### 3. Handle Errors Consistently
```go
func customMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := validateSomething(); err != nil {
            // Use consistent error format
            c.JSON(400, schema.ErrorResult{
                Success: false,
                ErrorInfo: schema.Error{
                    Code:    "VALIDATION_FAILED",
                    Message: err.Error(),
                },
                Data: nil,
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 4. Store Context Information Properly
```go
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Store useful information for downstream middleware and handlers
        c.Set("user_id", userID)
        c.Set("user_roles", roles)
        c.Set("auth_method", "bearer")
        c.Set("token_expires", expiresAt)
        
        c.Next()
    }
}
```

### 5. Make Middleware Configurable
```go
// Good: Configurable middleware
func rateLimitMiddleware(requestsPerSecond int, burst int) gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
    return func(c *gin.Context) { /* ... */ }
}

func timeoutMiddleware(duration time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) { /* ... */ }
}

// Usage
router.Use(rateLimitMiddleware(100, 10))
router.Use(timeoutMiddleware(30 * time.Second))
```

## Integration

Middleware integrates with:
- **[Security](./security.md)**: Automatic detection and OpenAPI documentation
- **[Router](./router.md)**: Enhanced middleware application and detection
- **[OpenAPI](./openapi.md)**: Security middleware documentation generation
- **[Results](./results.md)**: Standardized error response format
