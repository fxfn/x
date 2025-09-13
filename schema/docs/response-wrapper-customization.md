# Response Wrapper Customization

The schema package automatically wraps all responses in a standard format. You can customize this wrapping behavior to match your API standards.

## Default Behavior

By default, all responses are wrapped as follows:

**Success Response:**
```json
{
  "success": true,
  "data": {
    "id": "123",
    "name": "John Doe"
  },
  "error": null
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User not found"
  },
  "data": null
}
```

## Customization Options

### Option 1: Minimal Wrapper (No Wrapping)

Return just the data without any wrapper:

```go
import "github.com/fxfn/x/schema"

func main() {
    // Set minimal wrapper globally
    schema.SetResponseWrapper(schema.MinimalWrapper{})
    
    // Your handlers will now return unwrapped data
}
```

**Success Response:**
```json
{
  "id": "123",
  "name": "John Doe"
}
```

**Error Response:**
```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User not found"
  }
}
```

### Option 2: Custom Field Names

Customize the field names in the wrapper:

```go
func main() {
    customWrapper := schema.CustomWrapper{
        SuccessField: "ok",           // Instead of "success"
        DataField:    "result",       // Instead of "data"
        ErrorField:   "error",        // Keep "error" 
        AddTimestamp: true,           // Add timestamp field
    }
    
    schema.SetResponseWrapper(customWrapper)
}
```

**Success Response:**
```json
{
  "ok": true,
  "result": {
    "id": "123", 
    "name": "John Doe"
  },
  "error": null,
  "timestamp": 1640995200
}
```

**Error Response:**
```json
{
  "ok": false,
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User not found"
  },
  "result": null,
  "timestamp": 1640995200
}
```

### Option 3: Custom Wrapper Implementation

Create your own wrapper by implementing the `ResponseWrapper` interface:

```go
type APIStandardWrapper struct {
    Version string
}

func (w APIStandardWrapper) WrapSuccess(data interface{}) interface{} {
    return map[string]interface{}{
        "status": "success",
        "data":   data,
        "meta": map[string]interface{}{
            "version":   w.Version,
            "timestamp": time.Now().Unix(),
        },
    }
}

func (w APIStandardWrapper) WrapError(code, message string) interface{} {
    return map[string]interface{}{
        "status": "error",
        "error": map[string]string{
            "code":    code,
            "message": message,
        },
        "meta": map[string]interface{}{
            "version":   w.Version,
            "timestamp": time.Now().Unix(),
        },
    }
}

func main() {
    wrapper := APIStandardWrapper{Version: "v1.0"}
    schema.SetResponseWrapper(wrapper)
}
```

**Success Response:**
```json
{
  "status": "success",
  "data": {
    "id": "123",
    "name": "John Doe"
  },
  "meta": {
    "version": "v1.0",
    "timestamp": 1640995200
  }
}
```

## Common Wrapper Patterns

### 1. JSend Standard
```go
type JSendWrapper struct{}

func (w JSendWrapper) WrapSuccess(data interface{}) interface{} {
    return map[string]interface{}{
        "status": "success",
        "data":   data,
    }
}

func (w JSendWrapper) WrapError(code, message string) interface{} {
    return map[string]interface{}{
        "status": "error",
        "message": message,
        "code":    code,
    }
}
```

### 2. JSON:API Standard
```go
type JSONAPIWrapper struct{}

func (w JSONAPIWrapper) WrapSuccess(data interface{}) interface{} {
    return map[string]interface{}{
        "data": data,
        "meta": map[string]interface{}{
            "timestamp": time.Now().Unix(),
        },
    }
}

func (w JSONAPIWrapper) WrapError(code, message string) interface{} {
    return map[string]interface{}{
        "errors": []map[string]interface{}{
            {
                "status": "400",
                "code":   code,
                "title":  message,
            },
        },
    }
}
```

### 3. Envelope Pattern
```go
type EnvelopeWrapper struct{}

func (w EnvelopeWrapper) WrapSuccess(data interface{}) interface{} {
    return map[string]interface{}{
        "envelope": map[string]interface{}{
            "success": true,
            "payload": data,
        },
    }
}

func (w EnvelopeWrapper) WrapError(code, message string) interface{} {
    return map[string]interface{}{
        "envelope": map[string]interface{}{
            "success": false,
            "error": map[string]string{
                "code":    code,
                "message": message,
            },
        },
    }
}
```

## Usage Examples

### Setting Wrapper Globally

```go
func main() {
    // Set once at application startup
    schema.SetResponseWrapper(schema.MinimalWrapper{})
    
    // All handlers will use this wrapper
    router := gin.New()
    
    handler := schema.ValidateAndHandle(func(c *gin.Context, req MySchema) (*MyResponse, error) {
        return &MyResponse{}, nil
    })
    
    router.GET("/api/endpoint", handler.HandlerFunc())
}
```

### Different Wrappers for Different Routes

```go
func main() {
    router := gin.New()
    
    // API v1 routes with default wrapper
    schema.SetResponseWrapper(schema.DefaultWrapper{})
    v1Handler := schema.ValidateAndHandle(v1Handler)
    router.GET("/api/v1/users", v1Handler.HandlerFunc())
    
    // API v2 routes with custom wrapper  
    schema.SetResponseWrapper(schema.MinimalWrapper{})
    v2Handler := schema.ValidateAndHandle(v2Handler)
    router.GET("/api/v2/users", v2Handler.HandlerFunc())
}
```

### Conditional Wrapping

```go
type ConditionalWrapper struct {
    BaseWrapper schema.ResponseWrapper
}

func (w ConditionalWrapper) WrapSuccess(data interface{}) interface{} {
    // Check if data should be wrapped based on type or other criteria
    if shouldWrap(data) {
        return w.BaseWrapper.WrapSuccess(data)
    }
    return data
}

func (w ConditionalWrapper) WrapError(code, message string) interface{} {
    return w.BaseWrapper.WrapError(code, message)
}
```

## Migration Guide

### From Default to Custom Wrapper

1. **Identify your desired response format**
2. **Choose or implement a wrapper**
3. **Set it globally at startup**
4. **Test all endpoints**

```go
// Before (default)
{
  "success": true,
  "data": {...},
  "error": null
}

// After (custom)
{
  "status": "ok", 
  "result": {...},
  "timestamp": 1640995200
}
```

### Backward Compatibility

If you need to maintain backward compatibility:

```go
type BackwardCompatibleWrapper struct {
    UseNewFormat bool
}

func (w BackwardCompatibleWrapper) WrapSuccess(data interface{}) interface{} {
    if w.UseNewFormat {
        return map[string]interface{}{
            "status": "success",
            "result": data,
        }
    }
    
    // Use default format
    return schema.DefaultWrapper{}.WrapSuccess(data)
}
```

## Best Practices

1. **Set wrapper once** at application startup
2. **Be consistent** across your API
3. **Document your format** for API consumers
4. **Consider versioning** if changing existing APIs
5. **Test thoroughly** when changing wrappers
6. **Keep it simple** - avoid overly complex wrapper logic

## Interface Definition

```go
type ResponseWrapper interface {
    WrapSuccess(data interface{}) interface{}
    WrapError(code, message string) interface{}
}
```

Implement this interface to create your own custom wrapper behavior.
