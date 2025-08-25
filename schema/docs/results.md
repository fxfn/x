# Results

Standardized success and error response handling system.

## Overview

The Results system provides:
- Consistent response format across all endpoints
- Automatic success response wrapping
- Structured error handling with codes and messages
- Predefined error constructors for common cases
- Integration with validation errors

## Benefits

- **Consistency**: All API responses follow the same structure
- **Predictability**: Clients can rely on consistent error formats
- **Debugging**: Error codes and messages aid in troubleshooting
- **Type Safety**: Structured response types prevent formatting errors
- **Automation**: Success responses wrapped automatically

## Response Format

### Success Response
```json
{
  "success": true,
  "data": {
    // Your handler's response data
  },
  "error": null
}
```

### Error Response
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

## API Reference

### Types

#### `SuccessResult[T any]`
```go
type SuccessResult[T any] struct {
    Success   bool   `json:"success"`
    Data      T      `json:"data"`
    ErrorInfo *Error `json:"error"`
}
```

#### `ErrorResult`
```go
type ErrorResult struct {
    Success   bool   `json:"success"`
    Data      *any   `json:"data"`
    ErrorInfo Error  `json:"error"`
}
```

#### `Error`
```go
type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

### Error Constructors

#### `NewSchemaError(code, message string) error`
Creates a custom error with specific code and message.

#### `ValidationError(message string) error`
Creates a validation error.

#### `InvalidParamsError(message string) error`
Creates an invalid parameters error.

#### `InvalidQueryError(message string) error`
Creates an invalid query parameters error.

#### `InvalidBodyError(message string) error`
Creates an invalid request body error.

#### `MissingRequiredError(field string) error`
Creates a missing required field error.

#### `InvalidJSONError(message string) error`
Creates an invalid JSON format error.

## Examples

### Automatic Success Wrapping
```go
type UserResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func GetUser(c *gin.Context, req GetUserSchema) (UserResponse, error) {
    user := UserResponse{
        ID:   "123",
        Name: "John Doe",
        Age:  30,
    }
    
    // This response is automatically wrapped in SuccessResult
    return user, nil
}
```

**Generated Response:**
```json
{
  "success": true,
  "data": {
    "id": "123",
    "name": "John Doe",
    "age": 30
  },
  "error": null
}
```

### Custom Error Handling
```go
func GetUser(c *gin.Context, req GetUserSchema) (UserResponse, error) {
    userID := req.Params.ID
    
    // Business logic validation
    if userID == "banned" {
        return UserResponse{}, schema.NewSchemaError("USER_BANNED", "User account has been banned")
    }
    
    // Not found case
    user, err := userService.GetByID(userID)
    if err == userService.ErrNotFound {
        return UserResponse{}, schema.NewSchemaError("USER_NOT_FOUND", "User with specified ID does not exist")
    }
    
    // Database error
    if err != nil {
        return UserResponse{}, schema.NewSchemaError("DATABASE_ERROR", "Failed to retrieve user data")
    }
    
    return UserResponse{
        ID:   user.ID,
        Name: user.Name,
        Age:  user.Age,
    }, nil
}
```

**Error Response:**
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with specified ID does not exist"
  }
}
```

### Validation Error Handling
```go
func CreateUser(c *gin.Context, req CreateUserSchema) (UserResponse, error) {
    // Check for duplicate email
    if userService.EmailExists(req.Body.Email) {
        return UserResponse{}, schema.ValidationError("Email address is already registered")
    }
    
    // Validate age business rule
    if req.Body.Age < 13 {
        return UserResponse{}, schema.InvalidParamsError("Users must be at least 13 years old")
    }
    
    // Create user
    user, err := userService.Create(req.Body)
    if err != nil {
        return UserResponse{}, schema.NewSchemaError("CREATION_FAILED", "Failed to create user account")
    }
    
    return UserResponse{
        ID:   user.ID,
        Name: user.Name,
        Age:  user.Age,
    }, nil
}
```

### Predefined Error Constructors
```go
func ProcessRequest(c *gin.Context, req RequestSchema) (ResponseData, error) {
    // Use predefined error constructors for common cases
    
    // Validation errors
    if req.Body.Email == "" {
        return ResponseData{}, schema.MissingRequiredError("email")
    }
    
    if !isValidEmail(req.Body.Email) {
        return ResponseData{}, schema.InvalidBodyError("Email format is invalid")
    }
    
    // Query parameter errors
    if req.Query.Limit < 1 {
        return ResponseData{}, schema.InvalidQueryError("Limit must be greater than 0")
    }
    
    // URL parameter errors
    if !isValidUUID(req.Params.ID) {
        return ResponseData{}, schema.InvalidParamsError("ID must be a valid UUID")
    }
    
    // JSON parsing errors (usually handled automatically)
    if parseError != nil {
        return ResponseData{}, schema.InvalidJSONError("Request body contains malformed JSON")
    }
    
    // Generic validation error
    if businessValidationFails {
        return ResponseData{}, schema.ValidationError("Business rules validation failed")
    }
    
    // Success case
    return processSuccessfully(req)
}
```

### Complex Error Scenarios
```go
func UpdateUser(c *gin.Context, req UpdateUserSchema) (UserResponse, error) {
    userID := req.Params.ID
    
    // Multi-step validation with specific error codes
    user, err := userService.GetByID(userID)
    if err == userService.ErrNotFound {
        return UserResponse{}, schema.NewSchemaError("USER_NOT_FOUND", "User does not exist")
    }
    if err != nil {
        return UserResponse{}, schema.NewSchemaError("DATABASE_ERROR", "Failed to retrieve user")
    }
    
    // Permission check
    if !user.CanBeModified() {
        return UserResponse{}, schema.NewSchemaError("USER_LOCKED", "User account is locked and cannot be modified")
    }
    
    // Email uniqueness check
    if req.Body.Email != user.Email && userService.EmailExists(req.Body.Email) {
        return UserResponse{}, schema.NewSchemaError("EMAIL_TAKEN", "Email address is already in use")
    }
    
    // Business rule validation
    if req.Body.Role == "admin" && !isAdminUser(c) {
        return UserResponse{}, schema.NewSchemaError("INSUFFICIENT_PERMISSIONS", "Only administrators can assign admin role")
    }
    
    // Update operation
    updatedUser, err := userService.Update(userID, req.Body)
    if err == userService.ErrConcurrentModification {
        return UserResponse{}, schema.NewSchemaError("CONCURRENT_MODIFICATION", "User was modified by another request")
    }
    if err != nil {
        return UserResponse{}, schema.NewSchemaError("UPDATE_FAILED", "Failed to update user")
    }
    
    return UserResponse{
        ID:   updatedUser.ID,
        Name: updatedUser.Name,
        Age:  updatedUser.Age,
    }, nil
}
```

### Different Response Types
```go
// Simple data response
func GetUser(c *gin.Context, req GetUserSchema) (UserResponse, error) {
    // Returns: {"success": true, "data": {...}, "error": null}
}

// List response
func GetUsers(c *gin.Context, req GetUsersSchema) ([]UserResponse, error) {
    // Returns: {"success": true, "data": [...], "error": null}
}

// Boolean response
func CheckUserExists(c *gin.Context, req CheckUserSchema) (bool, error) {
    // Returns: {"success": true, "data": true, "error": null}
}

// String response
func GetUserStatus(c *gin.Context, req GetUserSchema) (string, error) {
    // Returns: {"success": true, "data": "active", "error": null}
}

// Complex nested response
type GetUserProfileResponse struct {
    User UserData `json:"user"`
    Profile ProfileData `json:"profile"`
    Settings SettingsData `json:"settings"`
}

func GetUserProfile(c *gin.Context, req GetUserProfileSchema) (GetUserProfileResponse, error) {
    // Returns: {"success": true, "data": {"user": {...}, "profile": {...}, "settings": {...}}, "error": null}
}
```

### Error Code Conventions
```go
// Use consistent error code patterns
const (
    // Validation errors
    ErrValidationFailed = "VALIDATION_FAILED"
    ErrInvalidEmail     = "INVALID_EMAIL"
    ErrInvalidFormat    = "INVALID_FORMAT"
    
    // Authentication/Authorization
    ErrUnauthorized        = "UNAUTHORIZED"
    ErrForbidden          = "FORBIDDEN"
    ErrTokenExpired       = "TOKEN_EXPIRED"
    ErrInvalidCredentials = "INVALID_CREDENTIALS"
    
    // Resource errors
    ErrUserNotFound    = "USER_NOT_FOUND"
    ErrResourceExists  = "RESOURCE_EXISTS"
    ErrResourceLocked  = "RESOURCE_LOCKED"
    
    // Business logic errors
    ErrBusinessRule     = "BUSINESS_RULE_VIOLATION"
    ErrInsufficientFunds = "INSUFFICIENT_FUNDS"
    ErrQuotaExceeded    = "QUOTA_EXCEEDED"
    
    // System errors
    ErrDatabaseError   = "DATABASE_ERROR"
    ErrExternalService = "EXTERNAL_SERVICE_ERROR"
    ErrInternalError   = "INTERNAL_ERROR"
)

func CreatePayment(c *gin.Context, req CreatePaymentSchema) (PaymentResponse, error) {
    // Use consistent error codes
    account, err := accountService.GetByID(req.Body.AccountID)
    if err == accountService.ErrNotFound {
        return PaymentResponse{}, schema.NewSchemaError(ErrUserNotFound, "Account not found")
    }
    
    if account.Balance < req.Body.Amount {
        return PaymentResponse{}, schema.NewSchemaError(ErrInsufficientFunds, "Account balance is insufficient")
    }
    
    if account.DailySpent+req.Body.Amount > account.DailyLimit {
        return PaymentResponse{}, schema.NewSchemaError(ErrQuotaExceeded, "Daily spending limit exceeded")
    }
    
    // Process payment...
}
```

## Framework Integration

### Automatic Validation Errors
```go
// When validation fails, the framework automatically returns:
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERR_VALIDATION_FAILED",
    "message": "Field validation for 'Email' failed on the 'email' tag"
  }
}
```

### JSON Parsing Errors
```go
// When JSON parsing fails, the framework returns:
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERR_INVALID_JSON",
    "message": "Invalid JSON format in request body"
  }
}
```

### Default Error Handling
```go
// If an error is returned without a specific code:
return ResponseData{}, errors.New("something went wrong")

// The framework returns:
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERR_NOT_SPECIFIED",
    "message": "An unknown exception occurred"
  }
}
```

## Best Practices

### 1. Use Descriptive Error Codes
```go
// Good
schema.NewSchemaError("USER_EMAIL_ALREADY_EXISTS", "Email address is already registered")
schema.NewSchemaError("PAYMENT_INSUFFICIENT_FUNDS", "Account balance is insufficient for this transaction")

// Avoid
schema.NewSchemaError("ERROR", "Something went wrong")
schema.NewSchemaError("ERR_1", "Error occurred")
```

### 2. Provide Actionable Error Messages
```go
// Good
schema.NewSchemaError("PASSWORD_TOO_WEAK", "Password must be at least 8 characters with uppercase, lowercase, and numbers")
schema.NewSchemaError("RATE_LIMIT_EXCEEDED", "Too many requests. Please wait 60 seconds before trying again")

// Avoid
schema.NewSchemaError("VALIDATION_ERROR", "Invalid input")
schema.NewSchemaError("FORBIDDEN", "Access denied")
```

### 3. Use Consistent Error Code Patterns
```go
// Resource-based patterns
USER_NOT_FOUND
USER_ALREADY_EXISTS
USER_PERMISSION_DENIED

// Action-based patterns
CREATE_FAILED
UPDATE_FAILED
DELETE_FAILED

// Validation patterns
INVALID_EMAIL_FORMAT
INVALID_PASSWORD_LENGTH
MISSING_REQUIRED_FIELD
```

### 4. Log Internal Errors, Return Generic Messages
```go
func ProcessPayment(c *gin.Context, req PaymentSchema) (PaymentResponse, error) {
    payment, err := paymentService.Process(req.Body)
    if err != nil {
        // Log detailed error for debugging
        log.Error("Payment processing failed", 
            "user_id", req.Body.UserID,
            "amount", req.Body.Amount,
            "error", err)
        
        // Return generic error to client
        return PaymentResponse{}, schema.NewSchemaError("PAYMENT_FAILED", "Payment could not be processed")
    }
    
    return PaymentResponse{...}, nil
}
```

### 5. Handle Different Error Types Appropriately
```go
func GetUser(c *gin.Context, req GetUserSchema) (UserResponse, error) {
    user, err := userService.GetByID(req.Params.ID)
    
    switch {
    case err == userService.ErrNotFound:
        // Client error - user doesn't exist
        return UserResponse{}, schema.NewSchemaError("USER_NOT_FOUND", "User not found")
        
    case err == userService.ErrPermissionDenied:
        // Authorization error
        return UserResponse{}, schema.NewSchemaError("ACCESS_DENIED", "Insufficient permissions to view user")
        
    case err != nil:
        // Server error - don't expose internal details
        log.Error("Database error retrieving user", "error", err)
        return UserResponse{}, schema.NewSchemaError("INTERNAL_ERROR", "Unable to retrieve user data")
        
    default:
        // Success case
        return UserResponse{...}, nil
    }
}
```

## Integration

Results integrate with:
- **[Handlers](./handlers.md)**: Automatic response wrapping and error handling
- **[Validation](./validation.md)**: Automatic validation error formatting  
- **[OpenAPI](./openapi.md)**: Response schema documentation generation
- **[Security](./security.md)**: Authentication error standardization
