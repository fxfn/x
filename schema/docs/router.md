# Router

Enhanced router wrapper that provides automatic type registration and security middleware detection.

## Overview

The Router system:
- Wraps Gin's router with enhanced functionality
- Automatically registers typed handlers for OpenAPI generation
- Detects security middleware using reflection
- Provides type-safe route registration
- Supports both individual routes and route groups

## Benefits

- **Automatic Registration**: Handlers and security schemes registered automatically
- **Type Safety**: Compile-time guarantees for route setup
- **Security Detection**: Reflection-based middleware detection
- **Gin Compatibility**: Drop-in replacement for Gin router
- **Group Support**: Enhanced route groups with security inheritance

## API Reference

### `WrapRouter(engine *gin.Engine) *RouterHelper`

Wraps an existing Gin engine with enhanced functionality.

**Parameters:**
- `engine`: Gin engine instance

**Returns:**
- `*RouterHelper`: Enhanced router wrapper

### `NewRouter() *RouterHelper`

Creates a new router with default Gin engine.

**Returns:**
- `*RouterHelper`: New enhanced router

### RouterHelper Methods

#### HTTP Methods
```go
func (r *RouterHelper) GET(path string, handlers ...interface{})
func (r *RouterHelper) POST(path string, handlers ...interface{})
func (r *RouterHelper) PUT(path string, handlers ...interface{})
func (r *RouterHelper) DELETE(path string, handlers ...interface{})
func (r *RouterHelper) PATCH(path string, handlers ...interface{})
```

#### Middleware
```go
func (r *RouterHelper) Use(middleware ...gin.HandlerFunc) gin.IRoutes
func (r *RouterHelper) UseSecurity(schemes ...SecurityScheme) gin.IRoutes
```

#### Grouping
```go
func (r *RouterHelper) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup
```

### RouterGroup Methods

#### HTTP Methods
```go
func (rg *RouterGroup) GET(path string, handlers ...interface{})
func (rg *RouterGroup) POST(path string, handlers ...interface{})
func (rg *RouterGroup) PUT(path string, handlers ...interface{})
func (rg *RouterGroup) DELETE(path string, handlers ...interface{})
func (rg *RouterGroup) PATCH(path string, handlers ...interface{})
```

#### Middleware
```go
func (rg *RouterGroup) Use(middleware ...gin.HandlerFunc) gin.IRoutes
```

The router automatically detects `SecurityScheme` middleware using reflection.

## Examples

### Basic Router Setup
```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/shipeedo/waypoints/pkg/schema"
)

func main() {
    // Option 1: Wrap existing Gin engine
    app := gin.Default()
    router := schema.WrapRouter(app)
    
    // Option 2: Create new router
    // router := schema.NewRouter()
    
    // Register routes
    router.GET("/health", schema.ValidateAndHandle(GetHealth))
    router.GET("/users/:id", schema.ValidateAndHandle(GetUser))
    router.POST("/users", schema.ValidateAndHandle(CreateUser))
    
    router.Run(":8080")
}
```

### Route Groups
```go
func setupRoutes() *schema.RouterHelper {
    router := schema.NewRouter()
    
    // Public API group
    public := router.Group("/api/v1")
    {
        public.GET("/health", schema.ValidateAndHandle(GetHealth))
        public.POST("/auth/login", schema.ValidateAndHandle(Login))
        public.POST("/auth/register", schema.ValidateAndHandle(Register))
    }
    
    // Protected API group
    protected := router.Group("/api/v1")
    protected.Use(authMiddleware.Middleware())
    {
        protected.GET("/profile", schema.ValidateAndHandle(GetProfile))
        protected.PUT("/profile", schema.ValidateAndHandle(UpdateProfile))
        
        // User management
        users := protected.Group("/users")
        {
            users.GET("", schema.ValidateAndHandle(GetUsers))
            users.GET("/:id", schema.ValidateAndHandle(GetUser))
            users.POST("", schema.ValidateAndHandle(CreateUser))
            users.PUT("/:id", schema.ValidateAndHandle(UpdateUser))
            users.DELETE("/:id", schema.ValidateAndHandle(DeleteUser))
        }
        
        // Admin only
        admin := protected.Group("/admin")
        admin.Use(adminMiddleware.Middleware())
        {
            admin.GET("/stats", schema.ValidateAndHandle(GetStats))
            admin.GET("/logs", schema.ValidateAndHandle(GetLogs))
        }
    }
    
    return router
}
```

### Security Middleware Detection
```go
func setupSecurity() *schema.RouterHelper {
    router := schema.NewRouter()
    
    // Create security schemes
    apiKey := schema.NewAPIKeySecurity(schema.APIKeyConfig{
        Name:    "ApiKeyAuth",
        In:      schema.APIKeyLocationHeader,
        KeyName: "X-API-Key",
        ValidateKey: func(key string) bool { return key == "secret" },
    })
    
    bearer := schema.NewBearerSecurity(schema.BearerConfig{
        Name:         "BearerAuth",
        BearerFormat: "JWT",
        ValidateToken: validateJWT,
    })
    
    // Method 1: Direct middleware application (automatically detected)
    protected := router.Group("/api/v1")
    protected.Use(apiKey.Middleware()) // Automatically detected as SecurityScheme
    {
        protected.GET("/users", schema.ValidateAndHandle(GetUsers))
    }
    
    // Method 2: Mixed middleware
    mixed := router.Group("/api/v2")
    mixed.Use(
        gin.Logger(),              // Regular middleware
        apiKey.Middleware(),       // Security middleware (auto-detected)
        gin.Recovery(),            // Regular middleware
    )
    {
        mixed.GET("/data", schema.ValidateAndHandle(GetData))
    }
    
    // Method 3: Multi-auth
    multiAuth := schema.NewMultiSecurity("FlexibleAuth", apiKey, bearer)
    flexible := router.Group("/api/v3")
    flexible.Use(multiAuth.Middleware()) // Multi-auth detection
    {
        flexible.GET("/profile", schema.ValidateAndHandle(GetProfile))
    }
    
    return router
}
```

### Handler Type Registration
```go
// The router automatically registers handler types for OpenAPI generation
func setupTypedRoutes() *schema.RouterHelper {
    router := schema.NewRouter()
    
    // These are automatically registered with their types
    router.GET("/users/:id", schema.ValidateAndHandle(GetUser))
    router.POST("/users", schema.ValidateAndHandle(CreateUser))
    router.PUT("/users/:id", schema.ValidateAndHandle(UpdateUser))
    
    // Security schemes are also automatically detected and registered
    apiKey := schema.NewAPIKeySecurity(schema.APIKeyConfig{...})
    
    protected := router.Group("/protected")
    protected.Use(apiKey.Middleware()) // Auto-detected and registered
    {
        protected.GET("/data", schema.ValidateAndHandle(GetProtectedData))
    }
    
    return router
}
```

### Complex Routing Example
```go
func setupComplexAPI() *schema.RouterHelper {
    router := schema.NewRouter()
    
    // Global middleware
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    router.Use(corsMiddleware())
    
    // Security schemes
    apiKey := schema.NewAPIKeySecurity(schema.APIKeyConfig{
        Name:    "ServiceAuth",
        In:      schema.APIKeyLocationHeader,
        KeyName: "X-Service-Key",
        ValidateKey: validateServiceKey,
    })
    
    userAuth := schema.NewBearerSecurity(schema.BearerConfig{
        Name:         "UserAuth",
        BearerFormat: "JWT",
        ValidateToken: validateUserToken,
    })
    
    adminAuth := schema.NewBearerSecurity(schema.BearerConfig{
        Name:         "AdminAuth",
        BearerFormat: "JWT", 
        ValidateToken: validateAdminToken,
    })
    
    // Public endpoints
    public := router.Group("/api/v1")
    {
        public.GET("/health", schema.ValidateAndHandle(GetHealth))
        public.GET("/version", schema.ValidateAndHandle(GetVersion))
        public.POST("/auth/login", schema.ValidateAndHandle(Login))
        public.POST("/auth/refresh", schema.ValidateAndHandle(RefreshToken))
    }
    
    // Service-to-service endpoints
    services := router.Group("/api/v1/internal")
    services.Use(apiKey.Middleware())
    {
        services.POST("/webhooks/user-created", schema.ValidateAndHandle(HandleUserCreated))
        services.GET("/metrics", schema.ValidateAndHandle(GetMetrics))
    }
    
    // User endpoints
    users := router.Group("/api/v1/users")
    users.Use(userAuth.Middleware())
    {
        users.GET("/profile", schema.ValidateAndHandle(GetProfile))
        users.PUT("/profile", schema.ValidateAndHandle(UpdateProfile))
        users.GET("/preferences", schema.ValidateAndHandle(GetPreferences))
        users.PUT("/preferences", schema.ValidateAndHandle(UpdatePreferences))
        
        // User's own data
        userOwned := users.Group("/:userId")
        userOwned.Use(ownershipMiddleware()) // Custom middleware
        {
            userOwned.GET("/posts", schema.ValidateAndHandle(GetUserPosts))
            userOwned.POST("/posts", schema.ValidateAndHandle(CreatePost))
            userOwned.PUT("/posts/:postId", schema.ValidateAndHandle(UpdatePost))
            userOwned.DELETE("/posts/:postId", schema.ValidateAndHandle(DeletePost))
        }
    }
    
    // Admin endpoints
    admin := router.Group("/api/v1/admin")
    admin.Use(adminAuth.Middleware())
    {
        admin.GET("/users", schema.ValidateAndHandle(GetAllUsers))
        admin.GET("/users/:id", schema.ValidateAndHandle(GetUserByID))
        admin.PUT("/users/:id/status", schema.ValidateAndHandle(UpdateUserStatus))
        admin.DELETE("/users/:id", schema.ValidateAndHandle(DeleteUser))
        
        admin.GET("/analytics", schema.ValidateAndHandle(GetAnalytics))
        admin.GET("/logs", schema.ValidateAndHandle(GetLogs))
        admin.POST("/maintenance", schema.ValidateAndHandle(TriggerMaintenance))
    }
    
    // Mixed auth endpoints (API key OR user token)
    multiAuth := schema.NewMultiSecurity("FlexibleAuth", apiKey, userAuth)
    flexible := router.Group("/api/v1/data")
    flexible.Use(multiAuth.Middleware())
    {
        flexible.GET("/public-stats", schema.ValidateAndHandle(GetPublicStats))
        flexible.GET("/reports/:id", schema.ValidateAndHandle(GetReport))
    }
    
    return router
}
```

### Custom Middleware Integration
```go
func customMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Custom logic
        c.Set("custom_data", "value")
        c.Next()
    }
}

func setupWithCustomMiddleware() *schema.RouterHelper {
    router := schema.NewRouter()
    
    // Global custom middleware
    router.Use(customMiddleware())
    
    // Group with mixed middleware
    api := router.Group("/api/v1")
    api.Use(
        requestIDMiddleware(),     // Custom middleware
        rateLimitMiddleware(),     // Custom middleware
        authScheme.Middleware(),   // Security middleware (auto-detected)
        loggingMiddleware(),       // Custom middleware
    )
    {
        api.GET("/data", schema.ValidateAndHandle(GetData))
    }
    
    return router
}
```

## Reflection-Based Detection

The router uses reflection to automatically detect security middleware:

### How It Works
```go
// When you register security middleware
apiKey := schema.NewAPIKeySecurity(config)
group.Use(apiKey.Middleware())

// The router:
// 1. Gets the function pointer using reflection
// 2. Looks it up in the middleware registry
// 3. Finds the associated SecurityScheme
// 4. Registers it for OpenAPI generation
```

### Registration Process
```go
// During SecurityScheme.Middleware() creation
func (a *APIKeySecurity) Middleware() gin.HandlerFunc {
    handler := func(c *gin.Context) {
        // ... authentication logic ...
    }
    
    // This registers the handler with its scheme
    schema.RegisterSecurityMiddleware(handler, a)
    return handler
}

// During route group middleware application
func (rg *RouterGroup) Use(middleware ...gin.HandlerFunc) gin.IRoutes {
    for _, handler := range middleware {
        // This checks if the handler is registered as security middleware
        if scheme, isSecurityMiddleware := schema.IsSecurityMiddleware(handler); isSecurityMiddleware {
            rg.groupSecuritySchemes = append(rg.groupSecuritySchemes, scheme)
        }
    }
    
    return rg.RouterGroup.Use(middleware...)
}
```

## Best Practices

### 1. Use Route Groups for Organization
```go
// Good: Organized by functionality
api := router.Group("/api/v1")
users := api.Group("/users")
posts := api.Group("/posts")
admin := api.Group("/admin")

// Avoid: Flat structure
router.GET("/api/v1/users/list", ...)
router.GET("/api/v1/users/create", ...)
router.GET("/api/v1/posts/list", ...)
```

### 2. Apply Security at Appropriate Levels
```go
// Good: Security at group level
protected := router.Group("/api/v1")
protected.Use(authMiddleware.Middleware())

// Avoid: Repeating security on every route
router.GET("/users", auth, handler1)
router.GET("/posts", auth, handler2)
router.GET("/comments", auth, handler3)
```

### 3. Use Descriptive Group Names
```go
// Good
publicAPI := router.Group("/api/v1/public")
userAPI := router.Group("/api/v1/users")
adminAPI := router.Group("/api/v1/admin")

// Avoid
group1 := router.Group("/api/v1/g1")
group2 := router.Group("/api/v1/g2")
```

### 4. Leverage Nested Groups
```go
api := router.Group("/api/v1")
{
    // Public endpoints
    public := api.Group("/public")
    {
        public.GET("/health", handlers.Health)
    }
    
    // Protected endpoints
    protected := api.Group("/protected")
    protected.Use(authMiddleware.Middleware())
    {
        // User endpoints
        users := protected.Group("/users")
        {
            users.GET("", handlers.GetUsers)
            users.POST("", handlers.CreateUser)
        }
        
        // Admin endpoints
        admin := protected.Group("/admin")
        admin.Use(adminMiddleware.Middleware())
        {
            admin.GET("/stats", handlers.GetStats)
        }
    }
}
```

### 5. Mix Security and Regular Middleware Correctly
```go
// Good: Security middleware mixed with regular middleware
group.Use(
    gin.Logger(),                // Regular
    rateLimitMiddleware(),       // Regular
    authScheme.Middleware(),     // Security (auto-detected)
    requestIDMiddleware(),       // Regular
)

// Works: Security middleware is automatically detected
```

## Integration

The Router integrates with:
- **[Handlers](./handlers.md)**: Automatic type registration for OpenAPI generation
- **[Security](./security.md)**: Reflection-based security middleware detection
- **[OpenAPI](./openapi.md)**: Route and security scheme documentation
- **[Middleware](./middleware.md)**: Enhanced middleware support with auto-detection
