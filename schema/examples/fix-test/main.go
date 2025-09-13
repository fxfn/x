package main

import (
	"fmt"

	"github.com/fxfn/x/schema"
	"github.com/gin-gonic/gin"
)

// Reproduce the SearchResponse structure that was causing the issue
type SearchResponse struct {
	Page struct {
		Current      int `json:"current"`
		Size         int `json:"size"`
		TotalPages   int `json:"total_pages"`
		TotalResults int `json:"total_results"`
	}
	Query struct {
		CustomerId string `json:"customer_id"`
		Page       int    `json:"page"`
		PageSize   int    `json:"page_size"`
		Query      string `json:"query"`
		TenantId   string `json:"tenant_id"`
	}
	Results []struct {
		Type     string  `json:"type"`
		EntityId string  `json:"ref"`
		Score    float64 `json:"score"`
		Title    string  `json:"title"`
		Details  string  `json:"details"`
	}
}

// Example schema for testing - this would trigger the original bug
type SearchSchema struct {
	Query struct {
		CustomerId string `json:"customer_id" query:"customer_id"`
		TenantId   string `json:"tenant_id" query:"tenant_id"`
		Query      string `json:"query" query:"q"`
		Page       int    `json:"page" query:"page" default:"1"`
		PageSize   int    `json:"page_size" query:"page_size" default:"10"`
	}
}

func main() {
	fmt.Println("Testing the fix for NumField panic...")

	// Create a Gin router
	router := gin.New()

	// Register a handler that would trigger the original panic
	handler := schema.ValidateAndHandle(func(c *gin.Context, s SearchSchema) (*SearchResponse, error) {
		// Dummy implementation
		response := &SearchResponse{}
		response.Query.CustomerId = s.Query.CustomerId
		response.Query.TenantId = s.Query.TenantId
		response.Query.Query = s.Query.Query
		response.Query.Page = s.Query.Page
		response.Query.PageSize = s.Query.PageSize

		return response, nil
	})

	// Register the handler
	router.GET("/search", handler.HandlerFunc())
	schema.RegisterTypedHandler("GET", "/search", handler)

	// This call would previously panic due to the NumField bug
	fmt.Println("Generating OpenAPI spec...")
	spec := schema.OpenAPI(router, &schema.OpenAPIOpts{
		Title:       "Test API",
		Description: "Testing the NumField fix",
		Version:     "1.0.0",
	})

	fmt.Printf("Generated spec with %d paths\n", len(spec.Paths))
	fmt.Println("Test completed successfully - no panic!")

	// Print some details about the generated schema
	if pathItem, exists := spec.Paths["/search"]; exists {
		if pathItem.Get != nil {
			fmt.Printf("Found GET operation with %d parameters\n", len(pathItem.Get.Parameters))
		}
	}

	fmt.Printf("Generated %d component schemas\n", len(spec.Components.Schemas))
}
