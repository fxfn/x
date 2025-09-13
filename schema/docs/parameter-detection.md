# Parameter Detection in Schema Package

The schema package automatically detects and extracts different types of parameters from your schema structs to generate OpenAPI specifications.

## Parameter Types

### 1. Query Parameters

Query parameters can be defined in several ways:

#### Method 1: Explicit Query Struct (Recommended)
```go
type MySchema struct {
    Query struct {
        Search   string `json:"search" query:"q"`
        Page     int    `json:"page" query:"page" default:"1"`
        PageSize int    `json:"page_size" query:"page_size" default:"10"`
    }
}
```

#### Method 2: Auto-Detection (New Feature)
The package now automatically detects primitive fields as query parameters:

```go
type PostSearchSchema struct {
    Query string `json:"query" validate:"required"`  // Auto-detected as query parameter
    Page  int    `json:"page" default:"1"`           // Auto-detected as query parameter
}
```

**Auto-detection rules:**
- Field must be a primitive type (string, int, float, bool) or slice of primitives
- Struct and slice-of-struct fields are excluded (these should be in request body)
- Parameter name is taken from `json` tag, or lowercase field name if no tag

#### Method 3: Explicit Query Tags
```go
type MySchema struct {
    SearchTerm string `query:"q"`        // Explicit query parameter
    UserID     int    `query:"user_id"`  // Explicit query parameter
}
```

### 2. Path Parameters

Path parameters are extracted from fields in a `Params` struct:

```go
type MySchema struct {
    Params struct {
        ID     string `param:"id"`       // URL: /users/:id
        UserID string `param:"user_id"`  // URL: /users/:user_id/posts
    }
}
```

### 3. Request Body

Request body is extracted from fields in a `Body` struct:

```go
type MySchema struct {
    Body struct {
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
        Age   int    `json:"age" validate:"min=0,max=120"`
    }
}
```

## Complete Example

```go
type CompleteSchema struct {
    // Path parameters from URL
    Params struct {
        UserID string `param:"user_id"`
    }
    
    // Query parameters
    Query struct {
        Filter string `json:"filter" query:"filter"`
        Limit  int    `json:"limit" query:"limit" default:"10"`
    }
    
    // Request body (for POST/PUT/PATCH)
    Body struct {
        Name    string   `json:"name" validate:"required"`
        Email   string   `json:"email" validate:"required,email"`
        Tags    []string `json:"tags"`
        Profile struct {
            Bio string `json:"bio"`
            Age int    `json:"age" validate:"min=0"`
        } `json:"profile"`
    }
}
```

## Anonymous Struct Naming

Anonymous structs in response types are automatically named based on their context:

```go
type SearchResponse struct {
    Page struct {           // Becomes "SearchResponsePage" 
        Current int `json:"current"`
        Total   int `json:"total"`
    } `json:"page"`
    
    Results []struct {      // Items become "SearchResponseResultsItem"
        ID    string `json:"id"`
        Title string `json:"title"`
    } `json:"results"`
}
```

## Validation Tags

The package supports validation using the `validate` tag:

```go
type UserSchema struct {
    Body struct {
        Email    string `json:"email" validate:"required,email"`
        Age      int    `json:"age" validate:"min=0,max=120"`
        Password string `json:"password" validate:"required,min=8"`
    }
}
```

## Best Practices

1. **Use explicit Query/Params/Body structs** for complex schemas
2. **Use auto-detection** for simple schemas with only query parameters  
3. **Always add validation tags** for required fields and constraints
4. **Use meaningful field names** as they affect generated schema names
5. **Add JSON tags** to control serialization names
6. **Use default tags** for optional query parameters with default values

## Migration Guide

If you have existing schemas that aren't generating query parameters:

**Before:**
```go
type OldSchema struct {
    Query string `json:"query" validate:"required"`
}
```

**After (Option 1 - No changes needed):**
The new auto-detection will automatically treat this as a query parameter.

**After (Option 2 - Explicit structure):**
```go
type NewSchema struct {
    Query struct {
        SearchTerm string `json:"query" query:"q" validate:"required"`
    }
}
```
