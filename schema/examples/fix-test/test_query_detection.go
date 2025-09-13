package main

import (
	"fmt"

	"github.com/fxfn/x/schema"
	"github.com/gin-gonic/gin"
)

// This represents your PostSearchSchema
type PostSearchSchema struct {
	Query string `json:"query" validate:"required"`
}

// Example response
type SearchResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func main() {
	fmt.Println("Testing query parameter auto-detection...")

	// Create a Gin router
	router := gin.New()

	// Register a POST handler with query parameters
	handler := schema.ValidateAndHandle(func(c *gin.Context, s PostSearchSchema) (*[]SearchResult, error) {
		// Dummy implementation
		results := []SearchResult{
			{ID: "1", Title: "Example result for: " + s.Query},
		}
		return &results, nil
	})

	// Register the handler
	router.POST("/api/search", handler.HandlerFunc())
	schema.RegisterTypedHandler("POST", "/api/search", handler)

	// Generate OpenAPI spec
	fmt.Println("Generating OpenAPI spec...")
	spec := schema.OpenAPI(router, &schema.OpenAPIOpts{
		Title:       "Test Search API",
		Description: "Testing query parameter detection",
		Version:     "1.0.0",
	})

	fmt.Printf("Generated spec with %d paths\n", len(spec.Paths))

	// Check if the POST operation has query parameters
	if pathItem, exists := spec.Paths["/api/search"]; exists {
		if pathItem.Post != nil {
			fmt.Printf("POST operation found with %d parameters\n", len(pathItem.Post.Parameters))

			// Look for the query parameter
			found := false
			for _, param := range pathItem.Post.Parameters {
				if param.Name == "query" && param.In == "query" {
					found = true
					fmt.Printf("✓ Found query parameter: %s (required: %v)\n", param.Name, param.Required)
					break
				}
			}

			if !found {
				fmt.Println("✗ Query parameter not found!")
			} else {
				fmt.Println("✓ Test passed! Query parameter was auto-detected.")
			}
		} else {
			fmt.Println("✗ POST operation not found!")
		}
	} else {
		fmt.Println("✗ Path /api/search not found!")
	}

	fmt.Printf("Generated %d component schemas\n", len(spec.Components.Schemas))
}
