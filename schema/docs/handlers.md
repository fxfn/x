# Handlers

Type-safe request handlers that automatically validate incoming requests and provide structured responses.

## Overview

Handlers in the Schema Framework are functions that:
- Accept a Gin context and a typed schema
- Return a typed response and optional error
- Are automatically wrapped with validation and error handling
- Generate OpenAPI documentation from their types

## Benefits

- **Type Safety**: Compile-time guarantees for request/response structure
- **Auto Validation**: Request parsing and validation happens automatically
- **Clean Code**: Focus on business logic, not boilerplate
- **Documentation**: OpenAPI specs generated from handler signatures

## Handler Function Signatures

### Basic Handler
```go
func HandlerName(c *gin.Context, schema SchemaType) (ResponseType, error)
```

### Handler Without Schema
```go
func HandlerName(c *gin.Context) (ResponseType, error)
```

### Handler Without Response
```go
func HandlerName(c *gin.Context, schema SchemaType) error
```

## API Reference

### `ValidateAndHandle[S, R any](handler func(*gin.Context, S) (R, error)) gin.HandlerFunc`

Wraps a typed handler function with automatic validation and response handling.

**Type Parameters:**
- `S`: Schema type for request validation
- `R`: Response type for the handler return

**Parameters:**
- `handler`: Your business logic function

**Returns:**
- `gin.HandlerFunc`: Middleware-compatible handler

### `RegisterTypedHandler(method, path string, handler TypedHandlerFunc)`

Registers handler type information for OpenAPI generation.

**Parameters:**
- `method`: HTTP method (GET, POST, etc.)
- `path`: Route path
- `handler`: Typed handler implementation

## Examples

### Simple GET Handler
```go
type GetUserSchema struct {
    Params struct {
        ID string `param:"id" validate:"required"`
    }
}

type UserResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func GetUser(c *gin.Context, req GetUserSchema) (UserResponse, error) {
    // Access validated parameters
    userID := req.Params.ID
    
    // Business logic
    user, err := userService.GetByID(userID)
    if err != nil {
        return UserResponse{}, schema.NewSchemaError("USER_NOT_FOUND", "User not found")
    }
    
    return UserResponse{
        ID:   user.ID,
        Name: user.Name,
        Age:  user.Age,
    }, nil
}

// Register with router
router.GET("/users/:id", schema.ValidateAndHandle(GetUser))
```

### POST Handler with Body
```go
type CreateUserSchema struct {
    Body struct {
        Name  string `json:"name" validate:"required,min=2,max=50"`
        Email string `json:"email" validate:"required,email"`
        Age   int    `json:"age" validate:"min=1,max=150"`
    }
}

type CreateUserResponse struct {
    ID      string `json:"id"`
    Message string `json:"message"`
}

func CreateUser(c *gin.Context, req CreateUserSchema) (CreateUserResponse, error) {
    // Create user with validated data
    user := &User{
        Name:  req.Body.Name,
        Email: req.Body.Email,
        Age:   req.Body.Age,
    }
    
    if err := userService.Create(user); err != nil {
        return CreateUserResponse{}, schema.NewSchemaError("CREATE_FAILED", err.Error())
    }
    
    return CreateUserResponse{
        ID:      user.ID,
        Message: "User created successfully",
    }, nil
}

router.POST("/users", schema.ValidateAndHandle(CreateUser))
```

### Complex Handler with Query, Params, and Body
```go
type UpdateUserSchema struct {
    Params struct {
        ID string `param:"id" validate:"required"`
    }
    Query struct {
        Notify bool `query:"notify" default:"false"`
    }
    Body struct {
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
    }
}

func UpdateUser(c *gin.Context, req UpdateUserSchema) (UserResponse, error) {
    userID := req.Params.ID
    shouldNotify := req.Query.Notify
    
    // Update user
    user, err := userService.Update(userID, userService.UpdateData{
        Name:  req.Body.Name,
        Email: req.Body.Email,
    })
    if err != nil {
        return UserResponse{}, schema.NewSchemaError("UPDATE_FAILED", err.Error())
    }
    
    // Send notification if requested
    if shouldNotify {
        notificationService.SendUserUpdated(user)
    }
    
    return UserResponse{
        ID:   user.ID,
        Name: user.Name,
        Age:  user.Age,
    }, nil
}

router.PUT("/users/:id", schema.ValidateAndHandle(UpdateUser))
```

### Handler Without Schema (No Request Data)
```go
type HealthResponse struct {
    Status    string `json:"status"`
    Timestamp int64  `json:"timestamp"`
}

func GetHealth(c *gin.Context) (HealthResponse, error) {
    return HealthResponse{
        Status:    "healthy",
        Timestamp: time.Now().Unix(),
    }, nil
}

// Register without schema validation
router.GET("/health", schema.ValidateAndHandle(GetHealth))
```

### Error Handling
```go
func GetUser(c *gin.Context, req GetUserSchema) (UserResponse, error) {
    userID := req.Params.ID
    
    // Validation error
    if userID == "banned" {
        return UserResponse{}, schema.NewSchemaError("USER_BANNED", "User account is banned")
    }
    
    // Not found error
    user, err := userService.GetByID(userID)
    if err == userService.ErrNotFound {
        return UserResponse{}, schema.NewSchemaError("USER_NOT_FOUND", "User not found")
    }
    
    // Internal error
    if err != nil {
        return UserResponse{}, schema.NewSchemaError("INTERNAL_ERROR", "Failed to fetch user")
    }
    
    return UserResponse{
        ID:   user.ID,
        Name: user.Name,
        Age:  user.Age,
    }, nil
}
```

## Best Practices

### 1. Use Descriptive Schema Names
```go
// Good
type GetUserByIDSchema struct { ... }
type CreateUserSchema struct { ... }

// Avoid
type UserSchema struct { ... }
type Schema struct { ... }
```

### 2. Group Related Fields
```go
type UpdateUserSchema struct {
    Params struct {
        ID string `param:"id" validate:"required"`
    }
    Query struct {
        Notify bool   `query:"notify" default:"false"`
        Force  bool   `query:"force" default:"false"`
    }
    Body struct {
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
    }
}
```

### 3. Use Validation Tags
```go
type CreateUserSchema struct {
    Body struct {
        Name     string `json:"name" validate:"required,min=2,max=50"`
        Email    string `json:"email" validate:"required,email"`
        Age      int    `json:"age" validate:"min=1,max=150"`
        Country  string `json:"country" validate:"required,oneof=US CA UK"`
    }
}
```

### 4. Handle Errors Appropriately
```go
func CreateUser(c *gin.Context, req CreateUserSchema) (CreateUserResponse, error) {
    // Use specific error codes for different scenarios
    if userService.EmailExists(req.Body.Email) {
        return CreateUserResponse{}, schema.NewSchemaError("EMAIL_EXISTS", "Email already registered")
    }
    
    if err := userService.Create(req.Body); err != nil {
        // Log internal errors but return generic message
        log.Error("Failed to create user", "error", err)
        return CreateUserResponse{}, schema.NewSchemaError("CREATE_FAILED", "Failed to create user")
    }
    
    return CreateUserResponse{...}, nil
}
```

### 5. Leverage Context for Request-Scoped Data
```go
func GetUser(c *gin.Context, req GetUserSchema) (UserResponse, error) {
    // Access authentication info set by middleware
    if apiKey, exists := c.Get("api_key"); exists {
        log.Info("Request authenticated", "api_key", apiKey)
    }
    
    // Access request ID for tracing
    if requestID, exists := c.Get("request_id"); exists {
        log.Info("Processing request", "request_id", requestID)
    }
    
    // Your handler logic...
}
```

## Integration with OpenAPI

Handlers automatically contribute to OpenAPI documentation:

- **Parameters**: Extracted from `Params` and `Query` fields
- **Request Body**: Generated from `Body` field
- **Responses**: Generated from return type
- **Security**: Applied from middleware detection

See [OpenAPI Documentation](./openapi.md) for more details.
