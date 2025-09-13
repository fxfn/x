package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fxfn/x/schema"
)

// Test struct with anonymous structs like SearchResponse
type TestResponse struct {
	Page struct {
		Current      int `json:"current"`
		Size         int `json:"size"`
		TotalPages   int `json:"total_pages"`
		TotalResults int `json:"total_results"`
	} `json:"page"`
	Query struct {
		CustomerId string `json:"customer_id"`
		Page       int    `json:"page"`
		PageSize   int    `json:"page_size"`
		Query      string `json:"query"`
		TenantId   string `json:"tenant_id"`
	} `json:"query"`
	Results []struct {
		Type     string  `json:"type"`
		EntityId string  `json:"ref"`
		Score    float64 `json:"score"`
		Title    string  `json:"title"`
		Details  string  `json:"details"`
	} `json:"results"`
}

func main() {
	fmt.Println("Testing anonymous struct naming fix...")

	// Create schemas map to collect generated schemas
	schemas := make(map[string]*schema.JSONSchema)

	// Test the schema generation
	testType := reflect.TypeOf(TestResponse{})

	// Generate the schema - this should now create properly named anonymous struct schemas
	generatedSchema := generateTestSchema(testType, schemas)

	fmt.Printf("Generated schema ref: %s\n", generatedSchema.Ref)
	fmt.Printf("Total schemas created: %d\n", len(schemas))

	fmt.Println("\nGenerated schema names:")
	for name, _ := range schemas {
		fmt.Printf("  - %s\n", name)
	}

	// Check if we have the expected schema names
	expectedNames := []string{"TestResponse", "TestResponsePage", "TestResponseQuery", "TestResponseResultsItem"}
	fmt.Println("\nExpected vs Generated:")
	for _, expected := range expectedNames {
		if _, exists := schemas[expected]; exists {
			fmt.Printf("  ✓ %s - Found\n", expected)
		} else {
			fmt.Printf("  ✗ %s - Missing\n", expected)
		}
	}
}

// Simple test function to generate schema
func generateTestSchema(t reflect.Type, schemas map[string]*schema.JSONSchema) *schema.JSONSchema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return &schema.JSONSchema{Type: "object"}
	}

	schemaName := t.Name()
	if schemaName == "" {
		schemaName = "AnonymousStruct"
	}

	if _, exists := schemas[schemaName]; exists {
		return &schema.JSONSchema{Ref: "#/components/schemas/" + schemaName}
	}

	properties := make(map[string]*schema.JSONSchema)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		jsonName := getJSONName(field)
		if jsonName == "-" {
			continue
		}

		var fieldSchema *schema.JSONSchema
		fieldType := field.Type

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		switch fieldType.Kind() {
		case reflect.String:
			fieldSchema = &schema.JSONSchema{Type: "string"}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldSchema = &schema.JSONSchema{Type: "integer"}
		case reflect.Float32, reflect.Float64:
			fieldSchema = &schema.JSONSchema{Type: "number"}
		case reflect.Bool:
			fieldSchema = &schema.JSONSchema{Type: "boolean"}
		case reflect.Slice, reflect.Array:
			fieldSchema = &schema.JSONSchema{Type: "array"}
			if fieldType.Elem().Kind() == reflect.Struct && fieldType.Elem().Name() == "" {
				// Anonymous struct in slice
				capitalizedJsonName := strings.ToUpper(jsonName[:1]) + jsonName[1:]
				itemSchemaName := schemaName + capitalizedJsonName + "Item"
				itemSchema := generateTestSchemaWithName(fieldType.Elem(), schemas, itemSchemaName)
				fieldSchema.Items = itemSchema
			} else {
				fieldSchema.Items = generateTestSchema(fieldType.Elem(), schemas)
			}
		case reflect.Struct:
			if fieldType.Name() == "" {
				// Anonymous struct
				capitalizedJsonName := strings.ToUpper(jsonName[:1]) + jsonName[1:]
				contextName := schemaName + capitalizedJsonName
				fieldSchema = generateTestSchemaWithName(fieldType, schemas, contextName)
			} else {
				fieldSchema = generateTestSchema(fieldType, schemas)
			}
		default:
			fieldSchema = &schema.JSONSchema{Type: "object"}
		}

		properties[jsonName] = fieldSchema
	}

	schema := &schema.JSONSchema{
		Type:       "object",
		Properties: properties,
	}

	schemas[schemaName] = schema
	return &schema.JSONSchema{Ref: "#/components/schemas/" + schemaName}
}

func generateTestSchemaWithName(t reflect.Type, schemas map[string]*schema.JSONSchema, schemaName string) *schema.JSONSchema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return &schema.JSONSchema{Type: "object"}
	}

	if _, exists := schemas[schemaName]; exists {
		return &schema.JSONSchema{Ref: "#/components/schemas/" + schemaName}
	}

	properties := make(map[string]*schema.JSONSchema)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		jsonName := getJSONName(field)
		if jsonName == "-" {
			continue
		}

		var fieldSchema *schema.JSONSchema
		fieldType := field.Type

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		switch fieldType.Kind() {
		case reflect.String:
			fieldSchema = &schema.JSONSchema{Type: "string"}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldSchema = &schema.JSONSchema{Type: "integer"}
		case reflect.Float32, reflect.Float64:
			fieldSchema = &schema.JSONSchema{Type: "number"}
		case reflect.Bool:
			fieldSchema = &schema.JSONSchema{Type: "boolean"}
		case reflect.Slice, reflect.Array:
			fieldSchema = &schema.JSONSchema{Type: "array"}
			fieldSchema.Items = generateTestSchema(fieldType.Elem(), schemas)
		case reflect.Struct:
			if fieldType.Name() == "" {
				// Anonymous struct
				capitalizedJsonName := strings.ToUpper(jsonName[:1]) + jsonName[1:]
				contextName := schemaName + capitalizedJsonName
				fieldSchema = generateTestSchemaWithName(fieldType, schemas, contextName)
			} else {
				fieldSchema = generateTestSchema(fieldType, schemas)
			}
		default:
			fieldSchema = &schema.JSONSchema{Type: "object"}
		}

		properties[jsonName] = fieldSchema
	}

	schema := &schema.JSONSchema{
		Type:       "object",
		Properties: properties,
	}

	schemas[schemaName] = schema
	return &schema.JSONSchema{Ref: "#/components/schemas/" + schemaName}
}

func getJSONName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return field.Name
	}
	parts := strings.Split(jsonTag, ",")
	if parts[0] == "" {
		return field.Name
	}
	return parts[0]
}
