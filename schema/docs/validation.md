# Schema Validation

Automatic request parsing and validation system that extracts and validates data from HTTP requests.

## Overview

The validation system:
- Parses query parameters, URL parameters, and request bodies
- Validates data according to struct tags
- Provides detailed error messages for validation failures
- Supports default values and type conversion
- Integrates with `github.com/go-playground/validator/v10`

## Benefits

- **Automatic Parsing**: No manual parameter extraction
- **Type Safety**: Compile-time type checking
- **Validation**: Comprehensive validation rules
- **Error Messages**: Clear, actionable error responses
- **Defaults**: Sensible default values

## Schema Structure

### Basic Schema
```go
type MySchema struct {
    Params struct {
        // URL parameters (/users/:id)
    }
    Query struct {
        // Query parameters (?limit=10&offset=0)
    }
    Body struct {
        // Request body (JSON)
    }
}
```

### Field Tags

#### Query Parameters
```go
type Schema struct {
    Query struct {
        Filter string `query:"filter"`                           // Maps to ?filter=value
        Limit  int    `query:"limit" default:"100"`             // Default value
        Active bool   `query:"active" default:"true"`           // Boolean conversion
        Sort   string `query:"sort" validate:"oneof=name date"` // Validation
    }
}
```

#### URL Parameters
```go
type Schema struct {
    Params struct {
        ID     string `param:"id" validate:"required"`      // Maps to /:id
        UserID string `param:"userId" validate:"required"`  // Maps to /:userId
        Slug   string `param:"slug" validate:"required"`    // Maps to /:slug
    }
}
```

#### Request Body
```go
type Schema struct {
    Body struct {
        Name  string `json:"name" validate:"required,min=2,max=50"`
        Email string `json:"email" validate:"required,email"`
        Age   int    `json:"age" validate:"min=1,max=150"`
    }
}
```

## API Reference

### Tag Types

#### `query:"name"`
Maps struct field to query parameter.

**Example:**
```go
Filter string `query:"filter"` // ?filter=value
```

#### `param:"name"`
Maps struct field to URL parameter.

**Example:**
```go
ID string `param:"id"` // /users/:id
```

#### `json:"name"`
Standard JSON tag for request body fields.

**Example:**
```go
Name string `json:"name"` // {"name": "value"}
```

#### `default:"value"`
Sets default value if parameter is not provided.

**Example:**
```go
Limit int `query:"limit" default:"100"`
```

#### `validate:"rules"`
Validation rules using go-playground/validator syntax.

**Example:**
```go
Email string `json:"email" validate:"required,email"`
```

## Validation Rules

### Common Rules

#### Required Fields
```go
Name string `validate:"required"`
```

#### String Validation
```go
Name     string `validate:"required,min=2,max=50"`
Username string `validate:"required,alphanum"`
Email    string `validate:"required,email"`
URL      string `validate:"required,url"`
UUID     string `validate:"required,uuid"`
```

#### Numeric Validation
```go
Age    int     `validate:"min=1,max=150"`
Price  float64 `validate:"min=0"`
Rating int     `validate:"min=1,max=5"`
```

#### Choice Validation
```go
Status   string `validate:"oneof=active inactive pending"`
Priority string `validate:"oneof=low medium high"`
```

#### Complex Rules
```go
Password string `validate:"required,min=8,containsany=!@#$%^&*"`
Phone    string `validate:"required,e164"`  // International phone format
```

## Examples

### Query Parameter Parsing
```go
type ListUsersSchema struct {
    Query struct {
        // Basic parameters
        Search string `query:"search"`
        Active bool   `query:"active" default:"true"`
        
        // Pagination
        Limit  int `query:"limit" default:"50" validate:"min=1,max=100"`
        Offset int `query:"offset" default:"0" validate:"min=0"`
        
        // Sorting
        SortBy    string `query:"sortBy" default:"created_at" validate:"oneof=name email created_at"`
        SortOrder string `query:"sortOrder" default:"desc" validate:"oneof=asc desc"`
        
        // Filtering
        Role      string `query:"role" validate:"omitempty,oneof=admin user guest"`
        CreatedAt string `query:"createdAt" validate:"omitempty,datetime=2006-01-02"`
    }
}

func ListUsers(c *gin.Context, req ListUsersSchema) ([]User, error) {
    // All parameters are parsed and validated
    users, err := userService.List(userService.ListOptions{
        Search:    req.Query.Search,
        Active:    req.Query.Active,
        Limit:     req.Query.Limit,
        Offset:    req.Query.Offset,
        SortBy:    req.Query.SortBy,
        SortOrder: req.Query.SortOrder,
        Role:      req.Query.Role,
        CreatedAt: req.Query.CreatedAt,
    })
    
    return users, err
}

// Usage: GET /users?search=john&limit=20&sortBy=name&role=admin
```

### URL Parameter Extraction
```go
type GetUserSchema struct {
    Params struct {
        ID string `param:"id" validate:"required"`
    }
}

type GetPostSchema struct {
    Params struct {
        UserID string `param:"userId" validate:"required,uuid"`
        PostID string `param:"postId" validate:"required"`
    }
}

func GetUserPost(c *gin.Context, req GetPostSchema) (Post, error) {
    // Parameters are automatically extracted and validated
    post, err := postService.GetByUserAndID(req.Params.UserID, req.Params.PostID)
    return post, err
}

// Usage: GET /users/123e4567-e89b-12d3-a456-426614174000/posts/my-post-slug
```

### Request Body Validation
```go
type CreateUserSchema struct {
    Body struct {
        // Required fields
        Name  string `json:"name" validate:"required,min=2,max=50"`
        Email string `json:"email" validate:"required,email"`
        
        // Optional fields with defaults
        Role   string `json:"role" default:"user" validate:"oneof=admin user guest"`
        Active bool   `json:"active" default:"true"`
        
        // Nested validation
        Profile struct {
            Bio       string `json:"bio" validate:"max=500"`
            Website   string `json:"website" validate:"omitempty,url"`
            Location  string `json:"location" validate:"max=100"`
        } `json:"profile"`
        
        // Array validation
        Tags []string `json:"tags" validate:"max=10,dive,min=1,max=20"`
    }
}

func CreateUser(c *gin.Context, req CreateUserSchema) (User, error) {
    // Request body is parsed and validated
    user := &User{
        Name:    req.Body.Name,
        Email:   req.Body.Email,
        Role:    req.Body.Role,
        Active:  req.Body.Active,
        Profile: UserProfile{
            Bio:      req.Body.Profile.Bio,
            Website:  req.Body.Profile.Website,
            Location: req.Body.Profile.Location,
        },
        Tags: req.Body.Tags,
    }
    
    return userService.Create(user)
}
```

### Complex Combined Schema
```go
type UpdateUserSchema struct {
    Params struct {
        ID string `param:"id" validate:"required,uuid"`
    }
    Query struct {
        Notify    bool `query:"notify" default:"false"`
        DryRun    bool `query:"dryRun" default:"false"`
        UpdatedBy string `query:"updatedBy" validate:"required"`
    }
    Body struct {
        Name  string `json:"name" validate:"omitempty,min=2,max=50"`
        Email string `json:"email" validate:"omitempty,email"`
        
        // Partial update - all fields optional but validated when present
        Profile struct {
            Bio      string `json:"bio" validate:"omitempty,max=500"`
            Website  string `json:"website" validate:"omitempty,url"`
            Location string `json:"location" validate:"omitempty,max=100"`
        } `json:"profile"`
    }
}

func UpdateUser(c *gin.Context, req UpdateUserSchema) (User, error) {
    userID := req.Params.ID
    
    // Validation ensures all data is clean
    updateData := userService.UpdateData{
        Name:    req.Body.Name,
        Email:   req.Body.Email,
        Profile: req.Body.Profile,
    }
    
    if req.Query.DryRun {
        // Return what would be updated without making changes
        return userService.PreviewUpdate(userID, updateData)
    }
    
    user, err := userService.Update(userID, updateData)
    if err != nil {
        return User{}, err
    }
    
    if req.Query.Notify {
        notificationService.UserUpdated(user, req.Query.UpdatedBy)
    }
    
    return user, nil
}
```

### Custom Types and Validation
```go
type Status string

const (
    StatusActive   Status = "active"
    StatusInactive Status = "inactive"
    StatusPending  Status = "pending"
)

type Priority int

const (
    PriorityLow Priority = iota + 1
    PriorityMedium
    PriorityHigh
)

type CreateTaskSchema struct {
    Body struct {
        Title       string   `json:"title" validate:"required,min=3,max=100"`
        Description string   `json:"description" validate:"max=1000"`
        Status      Status   `json:"status" default:"pending" validate:"oneof=active inactive pending"`
        Priority    Priority `json:"priority" default:"2" validate:"min=1,max=3"`
        Tags        []string `json:"tags" validate:"max=5,dive,min=1,max=20"`
        DueDate     string   `json:"dueDate" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
    }
}
```

## Error Responses

### Validation Error Format
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERR_VALIDATION_FAILED",
    "message": "Field validation for 'Name' failed on the 'required' tag"
  }
}
```

### Common Error Types

#### Missing Required Field
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERR_VALIDATION_FAILED", 
    "message": "Field validation for 'Email' failed on the 'required' tag"
  }
}
```

#### Invalid Format
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERR_VALIDATION_FAILED",
    "message": "Field validation for 'Email' failed on the 'email' tag"
  }
}
```

#### Out of Range
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERR_VALIDATION_FAILED",
    "message": "Field validation for 'Age' failed on the 'min' tag"
  }
}
```

## Type Conversions

### Automatic Conversions

The framework automatically converts query parameters and URL parameters to the appropriate Go types:

```go
type Schema struct {
    Query struct {
        // String to int
        Limit int `query:"limit" default:"50"`
        
        // String to bool (accepts: true, false, 1, 0, yes, no)
        Active bool `query:"active" default:"true"`
        
        // String to float
        Rating float64 `query:"rating"`
        
        // Arrays (comma-separated: ?tags=go,api,web)
        Tags []string `query:"tags"`
        
        // Custom types (implements encoding.TextUnmarshaler)
        Status CustomStatus `query:"status"`
    }
}
```

### Handling Conversion Errors

Invalid conversions are automatically caught and return validation errors:

```go
// Request: ?limit=abc
// Response: Field validation for 'Limit' failed on type conversion
```

## Best Practices

### 1. Use Descriptive Field Names
```go
// Good
type Schema struct {
    Query struct {
        SearchQuery string `query:"q"`
        PageSize    int    `query:"limit" default:"50"`
        PageOffset  int    `query:"offset" default:"0"`
    }
}

// Avoid generic names
type Schema struct {
    Query struct {
        Data   string `query:"data"`
        Number int    `query:"num"`
    }
}
```

### 2. Provide Sensible Defaults
```go
type ListSchema struct {
    Query struct {
        Limit     int    `query:"limit" default:"50" validate:"min=1,max=100"`
        Offset    int    `query:"offset" default:"0" validate:"min=0"`
        SortBy    string `query:"sortBy" default:"created_at"`
        SortOrder string `query:"sortOrder" default:"desc" validate:"oneof=asc desc"`
    }
}
```

### 3. Use Appropriate Validation Rules
```go
type CreateUserSchema struct {
    Body struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required,min=8"`
        Age      int    `json:"age" validate:"min=13,max=120"`
        Website  string `json:"website" validate:"omitempty,url"`
    }
}
```

### 4. Document Complex Validation
```go
type CreateProductSchema struct {
    Body struct {
        // SKU must be alphanumeric, 6-12 characters
        SKU string `json:"sku" validate:"required,alphanum,min=6,max=12"`
        
        // Price in cents (e.g., $19.99 = 1999)
        PriceCents int `json:"priceCents" validate:"required,min=1"`
        
        // Category must be one of predefined values
        Category string `json:"category" validate:"required,oneof=electronics clothing books"`
    }
}
```

## Integration

The validation system integrates seamlessly with:
- **[Handlers](./handlers.md)**: Automatic validation before handler execution
- **[OpenAPI](./openapi.md)**: Parameter and schema documentation generation
- **[Results](./results.md)**: Standardized error response format
