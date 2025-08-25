# OpenAPI Documentation

Automatic OpenAPI 3.1 specification generation from your typed handlers and security schemes.

## Overview

The OpenAPI system:
- Generates complete OpenAPI 3.1 specifications automatically
- Extracts parameters from handler schemas
- Documents request/response bodies with JSON Schema
- Includes security schemes and requirements
- Outputs JSON and YAML formats
- Provides Swagger UI endpoints

## Benefits

- **Zero Maintenance**: Documentation stays in sync with code
- **Complete Coverage**: All endpoints, parameters, and schemas documented
- **Industry Standard**: OpenAPI 3.1 compatible
- **Interactive**: Built-in Swagger UI support
- **Type Safety**: Documentation reflects actual code structure

## API Reference

### `OpenAPI(router *gin.Engine, opts *OpenAPIOpts) *OpenAPISpec`

Generates OpenAPI specification from registered routes.

**Parameters:**
- `router`: Gin engine with registered routes
- `opts`: Configuration options

**Returns:**
- `*OpenAPISpec`: Generated specification with handler methods

### `OpenAPIOpts`
```go
type OpenAPIOpts struct {
    Title       string // API title
    Description string // API description  
    Version     string // API version
    Contact     string // Contact email
    License     string // License name
    OutputFile  string // Optional file output path
}
```

### `OpenAPISpec Methods`

#### `HandleGetSwagger(c *gin.Context)`
Gin handler that serves the OpenAPI spec as JSON or YAML.

#### `toJSON() string`
Returns the specification as JSON string.

#### `toYAML() string` 
Returns the specification as YAML string.

## Examples

### Basic Setup
```go
func main() {
    app := gin.Default()
    router := schema.WrapRouter(app)
    
    // Register your routes
    router.GET("/users/:id", schema.ValidateAndHandle(GetUser))
    router.POST("/users", schema.ValidateAndHandle(CreateUser))
    
    // Generate OpenAPI documentation
    openApi := schema.OpenAPI(router.Engine, &schema.OpenAPIOpts{
        Title:       "User Management API",
        Description: "API for managing users in the system",
        Version:     "1.0.0",
        Contact:     "api-team@company.com",
        License:     "MIT",
        OutputFile:  "swagger.json", // Optional: write to file
    })
    
    // Serve documentation endpoints
    app.GET("/swagger.json", openApi.HandleGetSwagger)
    app.GET("/swagger.yaml", openApi.HandleGetSwagger)
    
    router.Run(":8080")
}
```

### Complete Example with Security
```go
func main() {
    app := gin.Default()
    router := schema.WrapRouter(app)
    
    // Set up security
    apiKey := schema.NewAPIKeySecurity(schema.APIKeyConfig{
        Name:        "ApiKeyAuth",
        Description: "API key authentication via X-API-Key header",
        In:          schema.APIKeyLocationHeader,
        KeyName:     "X-API-Key",
        ValidateKey: func(key string) bool { return key == "secret" },
    })
    
    bearer := schema.NewBearerSecurity(schema.BearerConfig{
        Name:         "BearerAuth",
        Description:  "JWT Bearer token authentication",
        BearerFormat: "JWT",
        ValidateToken: func(token string) bool { return validateJWT(token) },
    })
    
    // Public routes
    public := router.Group("/api/v1")
    public.GET("/health", schema.ValidateAndHandle(GetHealth))
    public.POST("/auth/login", schema.ValidateAndHandle(Login))
    
    // API key protected routes
    apiRoutes := router.Group("/api/v1")
    apiRoutes.Use(apiKey.Middleware())
    {
        apiRoutes.GET("/users", schema.ValidateAndHandle(GetUsers))
        apiRoutes.GET("/users/:id", schema.ValidateAndHandle(GetUser))
    }
    
    // JWT protected routes
    userRoutes := router.Group("/api/v1")
    userRoutes.Use(bearer.Middleware())
    {
        userRoutes.POST("/users", schema.ValidateAndHandle(CreateUser))
        userRoutes.PUT("/users/:id", schema.ValidateAndHandle(UpdateUser))
        userRoutes.DELETE("/users/:id", schema.ValidateAndHandle(DeleteUser))
    }
    
    // Multi-auth routes (API key OR JWT)
    multiAuth := schema.NewMultiSecurity("FlexibleAuth", apiKey, bearer)
    flexible := router.Group("/api/v1")
    flexible.Use(multiAuth.Middleware())
    {
        flexible.GET("/profile", schema.ValidateAndHandle(GetProfile))
        flexible.PUT("/profile", schema.ValidateAndHandle(UpdateProfile))
    }
    
    // Generate comprehensive documentation
    openApi := schema.OpenAPI(router.Engine, &schema.OpenAPIOpts{
        Title:       "Complete API",
        Description: "Full-featured API with authentication and comprehensive endpoints",
        Version:     "2.0.0",
        Contact:     "api-support@company.com",
        License:     "Apache 2.0",
        OutputFile:  "api-spec.json",
    })
    
    // Documentation endpoints
    app.GET("/docs/swagger.json", openApi.HandleGetSwagger)
    app.GET("/docs/swagger.yaml", openApi.HandleGetSwagger)
    
    // Serve Swagger UI (if you have static files)
    app.Static("/docs", "./swagger-ui")
    
    router.Run(":8080")
}
```

## Generated Documentation Structure

### Basic Information
```yaml
openapi: 3.1.1
info:
  title: User Management API
  description: API for managing users in the system
  version: 1.0.0
  contact:
    email: api-team@company.com
  license:
    name: MIT
```

### Security Schemes
```yaml
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
      description: API key authentication via X-API-Key header
    
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: JWT Bearer token authentication
```

### Path Documentation
```yaml
paths:
  /api/v1/users/{id}:
    get:
      summary: Get /api/v1/users/by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    default: true
                  data:
                    $ref: '#/components/schemas/GetUserResponse'
                  error:
                    type: null
                required: [success, data, error]
        '400':
          description: Error
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    default: false
                  error:
                    type: object
                    properties:
                      code:
                        type: string
                      message:
                        type: string
                    required: [code, message]
                  data:
                    type: null
                required: [success, error, data]
      security:
        - ApiKeyAuth: []
```

### Schema Generation
```yaml
components:
  schemas:
    GetUserResponse:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        email:
          type: string
        age:
          type: integer
        profile:
          $ref: '#/components/schemas/UserProfile'
      required: [id, name]
    
    CreateUserRequest:
      type: object
      properties:
        name:
          type: string
          minLength: 2
          maxLength: 50
        email:
          type: string
          format: email
        age:
          type: integer
          minimum: 1
          maximum: 150
      required: [name, email]
```

## Handler Schema Mapping

### Query Parameters
```go
type ListUsersSchema struct {
    Query struct {
        Search string `query:"search"`
        Limit  int    `query:"limit" default:"50" validate:"min=1,max=100"`
        Offset int    `query:"offset" default:"0" validate:"min=0"`
        Active bool   `query:"active" default:"true"`
    }
}
```

**Generated OpenAPI:**
```yaml
parameters:
  - name: search
    in: query
    required: false
    schema:
      type: string
  - name: limit
    in: query
    required: false
    schema:
      type: integer
      default: 50
      minimum: 1
      maximum: 100
  - name: offset
    in: query
    required: false
    schema:
      type: integer
      default: 0
      minimum: 0
  - name: active
    in: query
    required: false
    schema:
      type: boolean
      default: true
```

### Path Parameters
```go
type GetUserSchema struct {
    Params struct {
        ID string `param:"id" validate:"required"`
    }
}
```

**Generated OpenAPI:**
```yaml
parameters:
  - name: id
    in: path
    required: true
    schema:
      type: string
```

### Request Body
```go
type CreateUserSchema struct {
    Body struct {
        Name  string `json:"name" validate:"required,min=2,max=50"`
        Email string `json:"email" validate:"required,email"`
        Age   int    `json:"age" validate:"min=1,max=150"`
    }
}
```

**Generated OpenAPI:**
```yaml
requestBody:
  description: Request body
  required: true
  content:
    application/json:
      schema:
        $ref: '#/components/schemas/CreateUserRequest'
```

## Validation Integration

Validation tags are automatically converted to OpenAPI constraints:

### String Validation
```go
Name string `validate:"required,min=2,max=50"`
```
```yaml
schema:
  type: string
  minLength: 2
  maxLength: 50
```

### Numeric Validation
```go
Age int `validate:"min=18,max=120"`
```
```yaml
schema:
  type: integer
  minimum: 18
  maximum: 120
```

### Format Validation
```go
Email string `validate:"required,email"`
```
```yaml
schema:
  type: string
  format: email
```

### Choice Validation
```go
Status string `validate:"oneof=active inactive pending"`
```
```yaml
schema:
  type: string
  enum: [active, inactive, pending]
```

## Security Documentation

### Single Authentication
```go
// Route with single auth requirement
router.GET("/protected", apiKey, handler)
```

**Generated:**
```yaml
security:
  - ApiKeyAuth: []
```

### Multi-Authentication (OR Logic)
```go
// Route with multiple auth options
multiAuth := schema.NewMultiSecurity("FlexibleAuth", apiKey, bearer)
router.GET("/flexible", multiAuth, handler)
```

**Generated:**
```yaml
security:
  - ApiKeyAuth: []
  - BearerAuth: []
```

## Serving Documentation

### JSON Endpoint
```go
app.GET("/swagger.json", openApi.HandleGetSwagger)
```

**Usage:**
```bash
curl http://localhost:8080/swagger.json
```

### YAML Endpoint
```go
app.GET("/swagger.yaml", openApi.HandleGetSwagger)
```

**Usage:**
```bash
curl http://localhost:8080/swagger.yaml
```

### Static File Generation
```go
schema.OpenAPI(router.Engine, &schema.OpenAPIOpts{
    Title:      "My API",
    Version:    "1.0.0",
    OutputFile: "docs/swagger.json", // Writes to file
})
```

### Swagger UI Integration
```go
// Serve Swagger UI static files
app.Static("/docs", "./swagger-ui-dist")

// API specification endpoint
app.GET("/docs/swagger.json", openApi.HandleGetSwagger)
```

**Swagger UI HTML:**
```html
<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
    <link rel="stylesheet" type="text/css" href="swagger-ui-bundle.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/docs/swagger.json',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ]
        });
    </script>
</body>
</html>
```

## Advanced Features

### Custom Response Descriptions
```go
// Add custom descriptions via comments or tags
type UserResponse struct {
    ID   string `json:"id" description:"Unique user identifier"`
    Name string `json:"name" description:"User's full name"`
    Age  int    `json:"age" description:"User's age in years"`
}
```

### Complex Nested Schemas
```go
type CreateUserSchema struct {
    Body struct {
        PersonalInfo struct {
            FirstName string `json:"firstName" validate:"required"`
            LastName  string `json:"lastName" validate:"required"`
            BirthDate string `json:"birthDate" validate:"required,datetime=2006-01-02"`
        } `json:"personalInfo"`
        
        ContactInfo struct {
            Email   string `json:"email" validate:"required,email"`
            Phone   string `json:"phone" validate:"omitempty,e164"`
            Address struct {
                Street  string `json:"street" validate:"required"`
                City    string `json:"city" validate:"required"`
                Country string `json:"country" validate:"required,iso3166_1_alpha2"`
            } `json:"address"`
        } `json:"contactInfo"`
        
        Preferences []string `json:"preferences" validate:"max=10,dive,min=1"`
    }
}
```

### Multiple Response Types
```go
// The framework automatically generates 200 and 400 responses
// For custom response codes, use manual OpenAPI customization
```

## Best Practices

### 1. Use Descriptive Titles and Descriptions
```go
schema.OpenAPI(router.Engine, &schema.OpenAPIOpts{
    Title:       "User Management Service API",
    Description: "Comprehensive API for user account management, authentication, and profile operations",
    Version:     "2.1.0",
    Contact:     "api-support@company.com",
    License:     "MIT",
})
```

### 2. Version Your APIs
```go
// Use semantic versioning
Version: "1.2.3"

// Include version in paths
router.Group("/api/v1")
router.Group("/api/v2")
```

### 3. Provide Meaningful Examples
```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required" example:"John Doe"`
    Email string `json:"email" validate:"required,email" example:"john@example.com"`
    Age   int    `json:"age" validate:"min=18" example:"25"`
}
```

### 4. Document Error Cases
```go
// Use specific error codes in handlers
return UserResponse{}, schema.NewSchemaError("USER_NOT_FOUND", "User with specified ID does not exist")
```

### 5. Keep Schemas Focused
```go
// Good: Specific schemas for each operation
type GetUserSchema struct { ... }
type CreateUserSchema struct { ... }
type UpdateUserSchema struct { ... }

// Avoid: Generic schemas used everywhere
type UserSchema struct { ... }
```

## Integration

OpenAPI generation integrates with:
- **[Handlers](./handlers.md)**: Automatic parameter and response documentation
- **[Validation](./validation.md)**: Constraint documentation from validation tags
- **[Security](./security.md)**: Security scheme and requirement documentation
- **[Router](./router.md)**: Route discovery and registration
