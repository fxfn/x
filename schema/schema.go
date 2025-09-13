package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Schema represents the interface that all schemas must implement
type Schema interface{}

// SchemaValidator is the global validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// HandlerFunc represents a schema-validated handler function that can return either:
// - (*result, nil) for success
// - (nil, SchemaError) for explicit errors
// - (nil, ErrorResult) for direct error result control
type HandlerFunc[T Schema, R any] func(c *gin.Context, schema T) (*R, error)

// TypedHandler interface for handlers that carry type information
type TypedHandler interface {
	GetSchemaType() reflect.Type
	GetResponseType() reflect.Type
	ServeHTTP(*gin.Context)
}

// TypedHandlerFunc represents a gin.HandlerFunc that carries type information
type TypedHandlerFunc struct {
	handler      gin.HandlerFunc
	schemaType   reflect.Type
	responseType reflect.Type
}

func (t TypedHandlerFunc) GetSchemaType() reflect.Type {
	return t.schemaType
}

func (t TypedHandlerFunc) GetResponseType() reflect.Type {
	return t.responseType
}

func (t TypedHandlerFunc) ServeHTTP(c *gin.Context) {
	t.handler(c)
}

// Convert TypedHandlerFunc to gin.HandlerFunc
func (t TypedHandlerFunc) HandlerFunc() gin.HandlerFunc {
	return t.handler
}

// Global registry to store typed handlers for OpenAPI generation
var typedHandlers = make(map[string]TypedHandlerFunc)

// RegisterTypedHandler stores a typed handler for OpenAPI generation
func RegisterTypedHandler(method, path string, handler TypedHandlerFunc) {
	key := method + " " + path
	typedHandlers[key] = handler
}

// GetTypedHandler retrieves a typed handler by method and path
func GetTypedHandler(method, path string) (TypedHandlerFunc, bool) {
	key := method + " " + path
	handler, exists := typedHandlers[key]
	return handler, exists
}

// ValidateAndHandle wraps a handler function with schema validation and type information
func ValidateAndHandle[T Schema, R any](handler HandlerFunc[T, R]) TypedHandlerFunc {
	var schema T
	var response R

	schemaType := reflect.TypeOf(schema)
	responseType := reflect.TypeOf(response)

	// Remove pointer if it's a pointer type
	if responseType.Kind() == reflect.Ptr {
		responseType = responseType.Elem()
	}

	ginHandler := func(c *gin.Context) {
		var schema T

		// Parse and validate the schema
		if err := parseSchema(c, &schema); err != nil {
			errorResult := convertToErrorResult(err)
			wrappedError := globalWrapper.WrapError(errorResult.ErrorInfo.Code, errorResult.ErrorInfo.Message)
			c.JSON(400, wrappedError)
			return
		}

		// Call the handler with validated schema
		result, err := handler(c, schema)
		if err != nil {
			// Check if the error is actually an ErrorResult (user wants direct control)
			if errorResult, ok := err.(ErrorResult); ok {
				wrappedError := globalWrapper.WrapError(errorResult.ErrorInfo.Code, errorResult.ErrorInfo.Message)
				c.JSON(400, wrappedError)
				return
			}

			// Otherwise convert the error to an ErrorResult
			errorResult := convertToErrorResult(err)
			wrappedError := globalWrapper.WrapError(errorResult.ErrorInfo.Code, errorResult.ErrorInfo.Message)
			c.JSON(400, wrappedError)
			return
		}

		// Check if result is nil (shouldn't happen with proper error handling)
		if result == nil {
			wrappedError := globalWrapper.WrapError("ERR_INTERNAL", "Handler returned nil result without error")
			c.JSON(500, wrappedError)
			return
		}

		// Wrap the result using the configured wrapper (dereference the pointer)
		wrappedResult := globalWrapper.WrapSuccess(*result)
		c.JSON(200, wrappedResult)
	}

	return TypedHandlerFunc{
		handler:      ginHandler,
		schemaType:   schemaType,
		responseType: responseType,
	}
}

// parseSchema extracts and validates data from the request into the schema
func parseSchema(c *gin.Context, schema any) error {
	schemaValue := reflect.ValueOf(schema).Elem()
	schemaType := schemaValue.Type()

	// First pass: parse and set values (including defaults)
	for i := 0; i < schemaValue.NumField(); i++ {
		field := schemaValue.Field(i)
		fieldType := schemaType.Field(i)
		fieldName := strings.ToLower(fieldType.Name)

		if !field.CanSet() {
			continue
		}

		switch fieldName {
		case "params":
			if err := parseParams(c, field); err != nil {
				return fmt.Errorf("params validation failed: %w", err)
			}
		case "query":
			if err := parseQuery(c, field); err != nil {
				return fmt.Errorf("query validation failed: %w", err)
			}
		case "body":
			if err := parseBody(c, field); err != nil {
				return fmt.Errorf("body validation failed: %w", err)
			}
		}
	}

	// Second pass: validate the entire schema after all values are set
	if err := validate.Struct(schema); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// convertToErrorResult converts any error to an ErrorResult
func convertToErrorResult(err error) ErrorResult {
	// Check if it's a SchemaError (explicit error from handler)
	if schemaErr, ok := err.(SchemaError); ok {
		return NotOk(schemaErr.Code, schemaErr.Message)
	}

	// Handle validation errors from parseSchema
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "params validation failed"):
		return NotOk("ERR_INVALID_PARAMS", extractValidationMessage(errMsg))
	case strings.Contains(errMsg, "query validation failed"):
		return NotOk("ERR_INVALID_QUERY", extractValidationMessage(errMsg))
	case strings.Contains(errMsg, "body validation failed"):
		return NotOk("ERR_INVALID_BODY", extractValidationMessage(errMsg))
	case strings.Contains(errMsg, "validation failed"):
		return NotOk("ERR_VALIDATION_FAILED", extractValidationMessage(errMsg))
	case strings.Contains(errMsg, "required") && strings.Contains(errMsg, "missing"):
		return NotOk("ERR_MISSING_REQUIRED", errMsg)
	case strings.Contains(errMsg, "invalid JSON"):
		return NotOk("ERR_INVALID_JSON", "Request body contains invalid JSON")
	default:
		// Generic error - use default code and message
		return NotOk("ERR_NOT_SPECIFIED", "An unknown exception occurred")
	}
}

// extractValidationMessage extracts the meaningful part of validation error messages
func extractValidationMessage(errMsg string) string {
	// Extract the part after "validation failed: " if it exists
	if idx := strings.Index(errMsg, "validation failed: "); idx != -1 {
		return errMsg[idx+len("validation failed: "):]
	}

	// Extract the part after the first colon for other error types
	if idx := strings.Index(errMsg, ": "); idx != -1 {
		return errMsg[idx+2:]
	}

	return errMsg
}

// parseParams extracts URL parameters and maps them to the schema
func parseParams(c *gin.Context, field reflect.Value) error {
	fieldType := field.Type()

	for i := 0; i < field.NumField(); i++ {
		structField := field.Field(i)
		typeField := fieldType.Field(i)

		if !structField.CanSet() {
			continue
		}

		// Get param name from tag or use field name
		paramName := getTagValue(typeField, "param")
		if paramName == "" {
			paramName = strings.ToLower(typeField.Name)
		}

		paramValue := c.Param(paramName)
		if paramValue == "" {
			// Check if field is required
			if isRequired(typeField) {
				return fmt.Errorf("required param '%s' is missing", paramName)
			}
			continue
		}

		if err := setFieldValue(structField, paramValue); err != nil {
			return fmt.Errorf("invalid param '%s': %w", paramName, err)
		}
	}

	return nil
}

// parseQuery extracts query parameters and maps them to the schema
func parseQuery(c *gin.Context, field reflect.Value) error {
	fieldType := field.Type()

	for i := 0; i < field.NumField(); i++ {
		structField := field.Field(i)
		typeField := fieldType.Field(i)

		if !structField.CanSet() {
			continue
		}

		// Get query name from tag or use field name
		queryName := getTagValue(typeField, "query")
		if queryName == "" {
			// Try exact field name first, then lowercase
			queryName = typeField.Name
		}

		queryValue := c.Query(queryName)

		// If query tag exists but no value found, also try field name variants
		if queryValue == "" {
			// Try exact field name
			if fieldQueryValue := c.Query(typeField.Name); fieldQueryValue != "" {
				queryValue = fieldQueryValue
				queryName = typeField.Name
			} else if lowercaseQueryValue := c.Query(strings.ToLower(typeField.Name)); lowercaseQueryValue != "" {
				queryValue = lowercaseQueryValue
				queryName = strings.ToLower(typeField.Name)
			}
		}

		if queryValue == "" {
			// Check for default value
			if defaultVal := getTagValue(typeField, "default"); defaultVal != "" {
				queryValue = defaultVal
			} else if isRequired(typeField) {
				return fmt.Errorf("required query param '%s' is missing", queryName)
			} else {
				continue
			}
		}

		if err := setFieldValue(structField, queryValue); err != nil {
			return fmt.Errorf("invalid query param '%s': %w", queryName, err)
		}
	}

	return nil
}

// parseBody extracts the request body and maps it to the schema
func parseBody(c *gin.Context, field reflect.Value) error {
	if c.Request.ContentLength == 0 {
		// Check if body is required
		if hasRequiredFields(field.Type()) {
			return fmt.Errorf("request body is required")
		}
		return nil
	}

	// Create a pointer to the field for JSON unmarshaling
	bodyPtr := reflect.New(field.Type())
	bodyPtr.Elem().Set(field)

	if err := c.ShouldBindJSON(bodyPtr.Interface()); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}

	field.Set(bodyPtr.Elem())
	return nil
}

// Helper functions

func getTagValue(field reflect.StructField, tagName string) string {
	tag := field.Tag.Get(tagName)
	if tag == "" {
		return ""
	}

	// Handle tags like `query:"name,required"`
	parts := strings.Split(tag, ",")
	return parts[0]
}

func isRequired(field reflect.StructField) bool {
	// Check for required in various tags
	tags := []string{"validate", "binding", "query", "param", "json"}
	for _, tagName := range tags {
		tag := field.Tag.Get(tagName)
		if strings.Contains(tag, "required") {
			return true
		}
	}
	return false
}

func hasRequiredFields(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		if isRequired(t.Field(i)) {
			return true
		}
	}
	return false
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}
