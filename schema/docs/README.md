# Schema Framework Documentation

A comprehensive, type-safe API framework for Go applications built on top of Gin. The Schema Framework provides automatic request validation, OpenAPI documentation generation, security middleware, and structured response handling.

## 📚 Documentation Structure

- **[Handlers](./handlers.md)** - Type-safe request handlers with automatic validation
- **[Schema Validation](./validation.md)** - Request parsing and validation system
- **[Security](./security.md)** - Authentication and authorization middleware
- **[OpenAPI](./openapi.md)** - Automatic API documentation generation
- **[Router](./router.md)** - Enhanced router with automatic type registration
- **[Results](./results.md)** - Standardized success and error response handling
- **[Middleware](./middleware.md)** - Framework and custom middleware integration

## 🚀 Quick Start

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/shipeedo/waypoints/pkg/schema"
)

// 1. Define your schema
type GetUserSchema struct {
    Params struct {
        ID string `param:"id" validate:"required"`
    }
}

type UserResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// 2. Create your handler
func GetUser(c *gin.Context, req GetUserSchema) (UserResponse, error) {
    return UserResponse{
        ID:   req.Params.ID,
        Name: "John Doe",
    }, nil
}

// 3. Set up your application
func main() {
    app := gin.Default()
    router := schema.WrapRouter(app)
    
    // Security
    apiKey := schema.NewAPIKeySecurity(schema.APIKeyConfig{
        Name:    "ApiKeyAuth",
        In:      schema.APIKeyLocationHeader,
        KeyName: "X-API-Key",
        ValidateKey: func(key string) bool { return key == "secret" },
    })
    
    // Routes
    protected := router.Group("/api/v1")
    protected.Use(apiKey.Middleware())
    protected.GET("/users/:id", schema.ValidateAndHandle(GetUser))
    
    // OpenAPI docs
    schema.OpenAPI(router.Engine, &schema.OpenAPIOpts{
        Title:   "My API",
        Version: "1.0.0",
        OutputFile: "swagger.json",
    })
    
    router.Run(":8080")
}
```

## 🏗️ Architecture

The Schema Framework follows a layered architecture:

```
┌─────────────────┐
│   Application   │  Your handlers and business logic
├─────────────────┤
│    Framework    │  Schema validation, OpenAPI, Security
├─────────────────┤
│      Gin        │  HTTP routing and middleware
├─────────────────┤
│   Go Standard   │  HTTP server and networking
└─────────────────┘
```

## ✨ Key Features

- **🔒 Type Safety**: Full compile-time type checking for requests and responses
- **📊 Auto Validation**: Automatic request parsing and validation with detailed error messages
- **📖 OpenAPI Generation**: Automatic Swagger/OpenAPI 3.x documentation
- **🛡️ Security Built-in**: API Key, Bearer Token, and multi-auth support
- **🎯 Minimal Boilerplate**: Clean, declarative API design
- **⚡ Performance**: Reflection-based registration with runtime efficiency
- **🔧 Extensible**: Plugin architecture for custom middleware and security schemes

## 🧭 Navigation

Choose a section from the list above to dive deeper into specific functionality.
