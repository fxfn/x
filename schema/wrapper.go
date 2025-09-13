package schema

import (
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseWrapper interface allows customization of response wrapping
type ResponseWrapper interface {
	WrapSuccess(data interface{}) interface{}
	WrapError(code, message string) interface{}
}

// DefaultWrapper implements the current behavior
type DefaultWrapper struct{}

func (w DefaultWrapper) WrapSuccess(data interface{}) interface{} {
	return SuccessResult[interface{}]{
		Success: true,
		Data:    data,
		Error:   nil,
	}
}

func (w DefaultWrapper) WrapError(code, message string) interface{} {
	return ErrorResult{
		Success: false,
		ErrorInfo: Error{
			Code:    code,
			Message: message,
		},
		Data: nil,
	}
}

// MinimalWrapper returns just the data without wrapping
type MinimalWrapper struct{}

func (w MinimalWrapper) WrapSuccess(data interface{}) interface{} {
	return data
}

func (w MinimalWrapper) WrapError(code, message string) interface{} {
	return map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}
}

// CustomWrapper allows field name customization
type CustomWrapper struct {
	SuccessField string
	DataField    string
	ErrorField   string
	AddTimestamp bool
	AddRequestID bool
}

func (w CustomWrapper) WrapSuccess(data interface{}) interface{} {
	result := make(map[string]interface{})

	if w.SuccessField != "" {
		result[w.SuccessField] = true
	}

	dataField := w.DataField
	if dataField == "" {
		dataField = "data"
	}
	result[dataField] = data

	if w.ErrorField != "" {
		result[w.ErrorField] = nil
	}

	if w.AddTimestamp {
		result["timestamp"] = time.Now().Unix()
	}

	return result
}

func (w CustomWrapper) WrapError(code, message string) interface{} {
	result := make(map[string]interface{})

	if w.SuccessField != "" {
		result[w.SuccessField] = false
	}

	errorField := w.ErrorField
	if errorField == "" {
		errorField = "error"
	}
	result[errorField] = map[string]string{
		"code":    code,
		"message": message,
	}

	dataField := w.DataField
	if dataField == "" {
		dataField = "data"
	}
	result[dataField] = nil

	if w.AddTimestamp {
		result["timestamp"] = time.Now().Unix()
	}

	return result
}

// Global wrapper configuration
var globalWrapper ResponseWrapper = DefaultWrapper{}

// SetResponseWrapper sets the global response wrapper
func SetResponseWrapper(wrapper ResponseWrapper) {
	globalWrapper = wrapper
}

// GetResponseWrapper returns the current global wrapper
func GetResponseWrapper() ResponseWrapper {
	return globalWrapper
}

// Helper function to get request ID from context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// RequestIDWrapper adds request ID to responses
type RequestIDWrapper struct {
	BaseWrapper ResponseWrapper
}

func (w RequestIDWrapper) WrapSuccess(data interface{}) interface{} {
	result := w.BaseWrapper.WrapSuccess(data)

	// Add request ID if it's a map
	if resultMap, ok := result.(map[string]interface{}); ok {
		// Note: We can't access gin.Context here, so request ID would need to be passed differently
		// This is just an example of how you could extend wrappers
		return resultMap
	}

	return result
}

func (w RequestIDWrapper) WrapError(code, message string) interface{} {
	result := w.BaseWrapper.WrapError(code, message)

	// Add request ID if it's a map
	if resultMap, ok := result.(map[string]interface{}); ok {
		// Note: We can't access gin.Context here, so request ID would need to be passed differently
		return resultMap
	}

	return result
}
