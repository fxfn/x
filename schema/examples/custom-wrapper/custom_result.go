package main

import "fmt"

// Example: Custom wrapper with different field names and additional metadata
type CustomSuccessResult[T any] struct {
	Status    string      `json:"status" default:"success"`
	Result    T           `json:"result"`
	Error     interface{} `json:"error" default:"null"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

type CustomErrorResult struct {
	Status    string      `json:"status" default:"error"`
	Error     Error       `json:"error"`
	Result    interface{} `json:"result" default:"null"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Example: Minimal wrapper (just the data)
type MinimalSuccessResult[T any] struct {
	Data T `json:"data"`
}

// Example: API-standard wrapper
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Timestamp int64  `json:"timestamp"`
	Version   string `json:"version,omitempty"`
}

func main() {
	fmt.Println("Custom wrapper examples defined")
}
