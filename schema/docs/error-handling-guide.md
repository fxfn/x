# Error Handling Guide

The schema package provides a structured way to handle and return errors from your handlers. This guide explains the correct approaches to return meaningful error messages to your API clients.

## The Problem

When you return a generic Go error like this:

```go
documents, err := tenant.Collection.SearchDocuments(c.Request.Context(), customerId, &query)
if err != nil {
    return nil, fmt.Errorf("error searching documents: %w", err)
}
```

You get a generic error response:

```json
{
  "success": false,
  "error": {
    "code": "ERR_NOT_SPECIFIED", 
    "message": "An unknown exception occurred"
  },
  "data": null
}
```

This happens because the schema package converts unknown errors to generic messages for security reasons.

## The Solution

Use `SchemaError` to return structured, meaningful errors to your clients.

### Method 1: Using `NewSchemaError` (Recommended)

```go
import "github.com/fxfn/x/schema"

func SearchDocuments(c *gin.Context, req SearchSchema) (*SearchResponse, error) {
    documents, err := tenant.Collection.SearchDocuments(c.Request.Context(), customerId, &query)
    if err != nil {
        // Return a structured error with specific code and message
        return nil, schema.NewSchemaError("SEARCH_FAILED", "Failed to search documents: " + err.Error())
    }
    
    // ... rest of your logic
    return &response, nil
}
```

This will return:

```json
{
  "success": false,
  "error": {
    "code": "SEARCH_FAILED",
    "message": "Failed to search documents: connection timeout"
  },
  "data": null
}
```

### Method 2: Using Convenience Functions

```go
func SearchDocuments(c *gin.Context, req SearchSchema) (*SearchResponse, error) {
    documents, err := tenant.Collection.SearchDocuments(c.Request.Context(), customerId, &query)
    if err != nil {
        // Use convenience function for database errors
        return nil, schema.DatabaseError("document search")
    }
    
    return &response, nil
}
```

### Method 3: Conditional Error Handling

```go
func GetUser(c *gin.Context, req GetUserSchema) (*UserResponse, error) {
    user, err := userService.GetByID(req.Params.ID)
    
    switch {
    case err == userService.ErrNotFound:
        return nil, schema.NotFoundError("User")
        
    case err == userService.ErrPermissionDenied:
        return nil, schema.PermissionError("view user")
        
    case err != nil:
        // Log internal error details but return generic message
        log.Error("Database error", "error", err)
        return nil, schema.DatabaseError("user retrieval")
        
    default:
        return &UserResponse{...}, nil
    }
}
```

## Available Error Constructors

### Basic Constructor
```go
schema.NewSchemaError(code, message string) SchemaError
```

### Convenience Constructors
```go
schema.ValidationError(message string) SchemaError
schema.NotFoundError(resource string) SchemaError  
schema.DatabaseError(operation string) SchemaError
schema.PermissionError(action string) SchemaError
schema.ConflictError(message string) SchemaError
```

## Error Response Examples

### Validation Error
```go
if req.Query == "" {
    return nil, schema.ValidationError("Search query cannot be empty")
}
```

Response:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Search query cannot be empty"
  },
  "data": null
}
```

### Not Found Error
```go
if user == nil {
    return nil, schema.NotFoundError("User")
}
```

Response:
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found"
  },
  "data": null
}
```

### Custom Business Logic Error
```go
if account.Balance < amount {
    return nil, schema.NewSchemaError("INSUFFICIENT_FUNDS", 
        fmt.Sprintf("Account balance $%.2f is less than required $%.2f", 
            account.Balance, amount))
}
```

Response:
```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "Account balance $50.00 is less than required $100.00"
  },
  "data": null
}
```

## Best Practices

### 1. Use Descriptive Error Codes
```go
// Good
schema.NewSchemaError("USER_EMAIL_ALREADY_EXISTS", "Email address is already registered")
schema.NewSchemaError("PAYMENT_INSUFFICIENT_FUNDS", "Account balance is insufficient")

// Avoid
schema.NewSchemaError("ERROR", "Something went wrong")
schema.NewSchemaError("ERR_1", "Error occurred")
```

### 2. Provide Actionable Error Messages
```go
// Good - tells user what to do
schema.NewSchemaError("PASSWORD_TOO_WEAK", 
    "Password must be at least 8 characters with uppercase, lowercase, and numbers")

// Avoid - not actionable
schema.NewSchemaError("VALIDATION_ERROR", "Invalid input")
```

### 3. Don't Expose Internal Details
```go
// Good - log internal details, return generic message
log.Error("Database connection failed", "error", err)
return nil, schema.DatabaseError("user retrieval")

// Avoid - exposes internal details
return nil, schema.NewSchemaError("DB_ERROR", err.Error())
```

### 4. Use Consistent Error Code Patterns
```go
// Resource-based patterns
"USER_NOT_FOUND"
"USER_ALREADY_EXISTS" 
"USER_PERMISSION_DENIED"

// Operation-based patterns
"SEARCH_FAILED"
"CREATE_FAILED"
"UPDATE_FAILED"
"DELETE_FAILED"
```

### 5. Handle Different Error Types Appropriately
```go
func ProcessPayment(c *gin.Context, req PaymentSchema) (*PaymentResponse, error) {
    // Validation errors - client's fault
    if req.Body.Amount <= 0 {
        return nil, schema.ValidationError("Amount must be greater than zero")
    }
    
    // Business logic errors - expected scenarios
    if account.Balance < req.Body.Amount {
        return nil, schema.NewSchemaError("INSUFFICIENT_FUNDS", "Account balance is insufficient")
    }
    
    // External service errors - might be temporary
    payment, err := paymentService.Process(req.Body)
    if err != nil {
        log.Error("Payment service error", "error", err)
        return nil, schema.NewSchemaError("PAYMENT_FAILED", "Payment could not be processed")
    }
    
    return &PaymentResponse{...}, nil
}
```

## Your Specific Case

For your search documents scenario, here's the corrected approach:

```go
func SearchDocuments(c *gin.Context, req PostSearchSchema) (*SearchResponse, error) {
    documents, err := tenant.Collection.SearchDocuments(c.Request.Context(), customerId, &query)
    if err != nil {
        // Option 1: Include the actual error message (if safe to expose)
        return nil, schema.NewSchemaError("SEARCH_FAILED", 
            fmt.Sprintf("Failed to search documents: %v", err))
        
        // Option 2: Use generic message but log details
        log.Error("Document search failed", "error", err, "customer_id", customerId)
        return nil, schema.DatabaseError("document search")
        
        // Option 3: Handle specific error types
        if isTimeoutError(err) {
            return nil, schema.NewSchemaError("SEARCH_TIMEOUT", 
                "Search request timed out, please try again")
        }
        if isConnectionError(err) {
            return nil, schema.NewSchemaError("SEARCH_UNAVAILABLE", 
                "Search service is temporarily unavailable")
        }
        
        // Fallback for unknown errors
        return nil, schema.DatabaseError("document search")
    }
    
    // Build and return response
    response := &SearchResponse{
        // ... populate response
    }
    
    return response, nil
}
```

This will give you meaningful error responses that help both you (for debugging) and your API clients (for handling different error scenarios appropriately).
