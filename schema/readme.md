# Schema Framework

A comprehensive, type-safe API framework for Go applications built on Gin. Provides automatic request validation, OpenAPI documentation generation, security middleware, and structured response handling.

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
        Name:        "ApiKeyAuth",
        In:          schema.APIKeyLocationHeader,
        KeyName:     "X-API-Key",
        ValidateKey: func(key string) bool { return key == "secret" },
    })
    
    // Routes
    protected := router.Group("/api/v1")
    protected.Use(apiKey.Middleware())
    protected.GET("/users/:id", schema.ValidateAndHandle(GetUser))
    
    // OpenAPI docs
    schema.OpenAPI(router.Engine, &schema.OpenAPIOpts{
        Title:      "My API",
        Version:    "1.0.0",
        OutputFile: "swagger.json",
    })
    
    router.Run(":8080")
}
```

## ✨ Features

- **🔒 Type Safety**: Full compile-time type checking for requests and responses
- **📊 Auto Validation**: Automatic request parsing and validation with detailed error messages
- **📖 OpenAPI Generation**: Automatic Swagger/OpenAPI 3.x documentation
- **🛡️ Security Built-in**: API Key, Bearer Token, and multi-auth support
- **🎯 Minimal Boilerplate**: Clean, declarative API design
- **⚡ Performance**: Reflection-based registration with runtime efficiency
- **🔧 Extensible**: Plugin architecture for custom middleware and security schemes

## 📚 Complete Documentation

**[📖 Full Documentation →](./docs/README.md)**

### Core Components
- **[Handlers](./docs/handlers.md)** - Type-safe request handlers with automatic validation
- **[Schema Validation](./docs/validation.md)** - Request parsing and validation system
- **[Security](./docs/security.md)** - Authentication and authorization middleware
- **[OpenAPI](./docs/openapi.md)** - Automatic API documentation generation
- **[Router](./docs/router.md)** - Enhanced router with automatic type registration
- **[Results](./docs/results.md)** - Standardized success and error response handling
- **[Middleware](./docs/middleware.md)** - Framework and custom middleware integration

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

## 📖 Basic Usage

See **[Handlers Documentation](./docs/handlers.md)** for complete examples.

### Simple Handler
```go
func GetUser(c *gin.Context, req GetUserSchema) (UserResponse, error) {
    return UserResponse{
        ID:   req.Params.ID,
        Name: "John Doe",
    }, nil
}

router.GET("/users/:id", schema.ValidateAndHandle(GetUser))
```

### With Validation
```go
type CreateUserSchema struct {
    Body struct {
        Name  string `json:"name" validate:"required,min=2,max=50"`
        Email string `json:"email" validate:"required,email"`
    }
}

router.POST("/users", schema.ValidateAndHandle(CreateUser))
```

### With Security
```go
apiKey := schema.NewAPIKeySecurity(schema.APIKeyConfig{...})
protected := router.Group("/api/v1")
protected.Use(apiKey.Middleware())
protected.GET("/users", schema.ValidateAndHandle(GetUsers))
```

## 📋 Quick Reference

### Schema Structure
```go
type MySchema struct {
    Params struct { /* URL parameters */ }
    Query  struct { /* Query parameters */ }
    Body   struct { /* Request body */ }
}
```

### Common Tags
- `param:"id"` - URL parameter
- `query:"limit"` - Query parameter
- `json:"name"` - JSON field
- `default:"100"` - Default value
- `validate:"required,email"` - Validation rules

### Response Format
```json
{
  "success": true|false,
  "data": "...",
  "error": {"code": "...", "message": "..."}
}
```

### Error Handling
```go
return Response{}, schema.NewSchemaError("CODE", "message")
```

## 🔗 More Information

### Component Documentation
- **[📖 Complete Guide](./docs/README.md)** - Start here for comprehensive overview
- **[🎯 Handlers](./docs/handlers.md)** - Type-safe request handlers
- **[✅ Validation](./docs/validation.md)** - Request parsing and validation
- **[🔒 Security](./docs/security.md)** - Authentication and authorization
- **[📊 OpenAPI](./docs/openapi.md)** - Automatic documentation generation
- **[🚏 Router](./docs/router.md)** - Enhanced routing with auto-registration
- **[📤 Results](./docs/results.md)** - Standardized response handling
- **[🔧 Middleware](./docs/middleware.md)** - Framework and custom middleware

### External Resources
- [Gin Framework](https://gin-gonic.com/) - Underlying HTTP framework
- [Validator Package](https://pkg.go.dev/github.com/go-playground/validator/v10) - Validation rules reference
- [OpenAPI 3.1 Spec](https://spec.openapis.org/oas/v3.1.0) - OpenAPI specification