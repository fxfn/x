package main

import (
	"fmt"

	"github.com/fxfn/x/schema"
	"github.com/gin-gonic/gin"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GetUserSchema struct {
	Params struct {
		ID string `param:"id"`
	}
}

func main() {
	router := gin.New()

	// Example 1: Default wrapper (current behavior)
	fmt.Println("=== Example 1: Default Wrapper ===")

	handler1 := schema.ValidateAndHandle(func(c *gin.Context, req GetUserSchema) (*User, error) {
		return &User{ID: req.Params.ID, Name: "John Doe"}, nil
	})

	router.GET("/users/:id", handler1.HandlerFunc())

	// Example 2: Minimal wrapper (no wrapping, just return the data)
	fmt.Println("=== Example 2: Minimal Wrapper ===")

	schema.SetResponseWrapper(schema.MinimalWrapper{})

	handler2 := schema.ValidateAndHandle(func(c *gin.Context, req GetUserSchema) (*User, error) {
		return &User{ID: req.Params.ID, Name: "Jane Doe"}, nil
	})

	router.GET("/minimal/users/:id", handler2.HandlerFunc())

	// Example 3: Custom wrapper with different field names
	fmt.Println("=== Example 3: Custom Wrapper ===")

	customWrapper := schema.CustomWrapper{
		SuccessField: "ok",     // Instead of "success"
		DataField:    "result", // Instead of "data"
		ErrorField:   "error",  // Keep "error"
		AddTimestamp: true,     // Add timestamp field
	}

	schema.SetResponseWrapper(customWrapper)

	handler3 := schema.ValidateAndHandle(func(c *gin.Context, req GetUserSchema) (*User, error) {
		return &User{ID: req.Params.ID, Name: "Custom User"}, nil
	})

	router.GET("/custom/users/:id", handler3.HandlerFunc())

	// Example 4: Error handling with custom wrapper
	fmt.Println("=== Example 4: Error with Custom Wrapper ===")

	handler4 := schema.ValidateAndHandle(func(c *gin.Context, req GetUserSchema) (*User, error) {
		if req.Params.ID == "404" {
			return nil, schema.NewSchemaError("USER_NOT_FOUND", "User not found")
		}
		return &User{ID: req.Params.ID, Name: "Found User"}, nil
	})

	router.GET("/custom/users-with-error/:id", handler4.HandlerFunc())

	fmt.Println("Server configured with different wrapper examples")
	fmt.Println("Test endpoints:")
	fmt.Println("  GET /users/123 (default wrapper)")
	fmt.Println("  GET /minimal/users/123 (minimal wrapper)")
	fmt.Println("  GET /custom/users/123 (custom wrapper)")
	fmt.Println("  GET /custom/users-with-error/404 (custom error)")

	// Start server
	router.Run(":8080")
}
