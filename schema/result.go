package schema

import "fmt"

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SchemaError represents an error that can be returned from handlers
type SchemaError struct {
	Code    string
	Message string
}

func (e SchemaError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

type SuccessResult[T any] struct {
	Success bool        `json:"success" default:"true"`
	Data    T           `json:"data"`
	Error   interface{} `json:"error" default:"null"`
}

type ErrorResult struct {
	Success   bool        `json:"success" default:"false"`
	ErrorInfo Error       `json:"error"`
	Data      interface{} `json:"data" default:"null"`
}

// Implement error interface so ErrorResult can be returned as an error
func (er ErrorResult) Error() string {
	return fmt.Sprintf("[%s] %s", er.ErrorInfo.Code, er.ErrorInfo.Message)
}

// Result is a union type that can represent either success or error
type Result[T any] interface {
	isResult()
}

// Implement the interface for both types
func (s SuccessResult[T]) isResult() {}
func (e ErrorResult) isResult()      {}

// Helper functions for creating results
func Ok[T any](data T) SuccessResult[T] {
	return SuccessResult[T]{
		Success: true,
		Data:    data,
		Error:   nil,
	}
}

func NotOk(code, message string) ErrorResult {
	return ErrorResult{
		Success: false,
		ErrorInfo: Error{
			Code:    code,
			Message: message,
		},
		Data: nil,
	}
}

// Common error constructors that return ErrorResult directly
var (
	ErrUserNotFound   = ErrorResult{Success: false, ErrorInfo: Error{Code: "ERR_USER_NOT_FOUND", Message: "User not found"}, Data: nil}
	ErrInvalidRequest = ErrorResult{Success: false, ErrorInfo: Error{Code: "ERR_INVALID_REQUEST", Message: "Invalid request"}, Data: nil}
	ErrUnauthorized   = ErrorResult{Success: false, ErrorInfo: Error{Code: "ERR_UNAUTHORIZED", Message: "Unauthorized access"}, Data: nil}
	ErrForbidden      = ErrorResult{Success: false, ErrorInfo: Error{Code: "ERR_FORBIDDEN", Message: "Access forbidden"}, Data: nil}
	ErrInternalError  = ErrorResult{Success: false, ErrorInfo: Error{Code: "ERR_INTERNAL", Message: "Internal server error"}, Data: nil}
)
