package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-yaml/yaml"
)

type OpenAPIOpts struct {
	Title       string
	Description string
	Version     string
	Contact     string
	License     string
	OutputFile  string // Path to output swagger.json file
}

// OpenAPI 3.1 specification structures
type OpenAPISpec struct {
	OpenAPI    string              `json:"openapi" yaml:"openapi"`
	Info       Info                `json:"info" yaml:"info"`
	Paths      map[string]PathItem `json:"paths" yaml:"paths"`
	Components *Components         `json:"components,omitempty" yaml:"components,omitempty"`
}

type Info struct {
	Title       string   `json:"title" yaml:"title"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string   `json:"version" yaml:"version"`
	Contact     *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License     *License `json:"license,omitempty" yaml:"license,omitempty"`
}

type Contact struct {
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

type License struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Post   *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Put    *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
}

type Operation struct {
	Summary     string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses" yaml:"responses"`
	Tags        []string              `json:"tags,omitempty" yaml:"tags,omitempty"`
	Security    []map[string][]string `json:"security,omitempty" yaml:"security,omitempty"`
}

type Parameter struct {
	Name        string      `json:"name" yaml:"name"`
	In          string      `json:"in" yaml:"in"` // "query", "header", "path", "cookie"
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool        `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      *JSONSchema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type RequestBody struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType `json:"content" yaml:"content"`
	Required    bool                 `json:"required,omitempty" yaml:"required,omitempty"`
}

type Response struct {
	Description string               `json:"description" yaml:"description"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

type MediaType struct {
	Schema *JSONSchema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type Components struct {
	Schemas         map[string]*JSONSchema            `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	SecuritySchemes map[string]map[string]interface{} `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
}

type JSONSchema struct {
	Type                 string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Properties           map[string]*JSONSchema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items                *JSONSchema            `json:"items,omitempty" yaml:"items,omitempty"`
	Required             []string               `json:"required,omitempty" yaml:"required,omitempty"`
	Description          string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Example              interface{}            `json:"example,omitempty" yaml:"example,omitempty"`
	Default              interface{}            `json:"default,omitempty" yaml:"default,omitempty"`
	Minimum              *float64               `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	MinLength            *int                   `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength            *int                   `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Pattern              string                 `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Format               string                 `json:"format,omitempty" yaml:"format,omitempty"`
	Ref                  string                 `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	AdditionalProperties interface{}            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
}

// HandlerInfo stores information about a handler function
type HandlerInfo struct {
	SchemaType      reflect.Type
	ResponseType    reflect.Type
	Method          string
	Path            string
	SecuritySchemes []SecurityScheme
}

// Legacy HandlerTypeInfo for backward compatibility
type HandlerTypeInfo struct {
	SchemaType   reflect.Type
	ResponseType reflect.Type
}

type OutputFormat string

const (
	OutputFormatJSON OutputFormat = "json"
	OutputFormatYAML OutputFormat = "yaml"
)

func OpenAPI(router *gin.Engine, opts *OpenAPIOpts) *OpenAPISpec {
	spec := generateOpenAPISpec(router, opts)

	// Write to file if specified
	if opts.OutputFile != "" {
		var format OutputFormat
		if strings.Contains(opts.OutputFile, "json") {
			format = OutputFormatJSON
		} else {
			format = OutputFormatYAML
		}

		if err := writeSwaggerFile(spec, opts.OutputFile, format); err != nil {
			fmt.Printf("Error writing swagger file: %v\n", err)
		} else {
			fmt.Printf("Swagger specification written to %s\n", opts.OutputFile)
		}
	}

	return spec
}

func (o *OpenAPISpec) toJSON() string {
	json, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return ""
	}
	return string(json)
}

func (o *OpenAPISpec) toYAML() string {
	yaml, err := yaml.Marshal(o)
	if err != nil {
		return ""
	}
	return string(yaml)
}

func (o *OpenAPISpec) HandleGetSwagger(c *gin.Context) {
	if strings.Contains(c.Request.URL.Path, "json") {
		c.Data(200, "application/json", []byte(o.toJSON()))
	} else {
		c.Data(200, "text/vnd.yaml", []byte(o.toYAML()))
	}
}

func generateOpenAPISpec(router *gin.Engine, opts *OpenAPIOpts) *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI: "3.1.1",
		Info: Info{
			Title:       opts.Title,
			Description: opts.Description,
			Version:     opts.Version,
		},
		Paths: make(map[string]PathItem),
		Components: &Components{
			Schemas:         make(map[string]*JSONSchema),
			SecuritySchemes: make(map[string]map[string]interface{}),
		},
	}

	if opts.Contact != "" {
		spec.Info.Contact = &Contact{Email: opts.Contact}
	}
	if opts.License != "" {
		spec.Info.License = &License{Name: opts.License}
	}

	// Get all routes and analyze them
	routes := router.Routes()
	handlerInfos := extractHandlerInfos(routes)

	// Generate paths and schemas
	for _, info := range handlerInfos {
		// Convert Gin path format (:param) to OpenAPI format ({param})
		openAPIPath := convertGinPathToOpenAPI(info.Path)

		// Get existing path item or create new one
		pathItem, exists := spec.Paths[openAPIPath]
		if !exists {
			pathItem = PathItem{}
		}

		operation := generateOperation(info, spec.Components.Schemas, spec.Components.SecuritySchemes)

		switch strings.ToUpper(info.Method) {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "PATCH":
			pathItem.Patch = operation
		}

		spec.Paths[openAPIPath] = pathItem
	}

	return spec
}

func extractHandlerInfos(routes gin.RoutesInfo) []HandlerInfo {
	var handlerInfos []HandlerInfo

	for _, route := range routes {
		info := analyzeHandler(route)
		if info != nil {
			handlerInfos = append(handlerInfos, *info)
		}
	}

	return handlerInfos
}

func analyzeHandler(route gin.RouteInfo) *HandlerInfo {
	// Look up handler type information in the typed handlers registry
	typedHandler, exists := GetTypedHandler(route.Method, route.Path)

	if !exists {
		// If handler is not registered, skip this route
		return nil
	}

	// Get security schemes for this route
	securitySchemes := GetSecuritySchemes(route.Method, route.Path)

	return &HandlerInfo{
		SchemaType:      typedHandler.GetSchemaType(),
		ResponseType:    typedHandler.GetResponseType(),
		Method:          route.Method,
		Path:            route.Path,
		SecuritySchemes: securitySchemes,
	}
}

func generateOperation(info HandlerInfo, schemas map[string]*JSONSchema, securitySchemes map[string]map[string]interface{}) *Operation {
	operation := &Operation{
		Summary:   generateSummary(info.Method, info.Path),
		Responses: make(map[string]Response),
	}

	// Add security schemes to components and operation
	if len(info.SecuritySchemes) > 0 {
		var security []map[string][]string
		for _, scheme := range info.SecuritySchemes {
			// Check if this is a MultiSecurity scheme
			if multiSec, ok := scheme.(*MultiSecurity); ok {
				// For MultiSecurity, register each component scheme and create OR logic
				var multiSecurityReqs []string
				for _, componentScheme := range multiSec.GetComponentSchemes() {
					// Add component to securitySchemes if not already present
					schemeName, schemeSpec := componentScheme.GetSecurityScheme()
					if _, exists := securitySchemes[schemeName]; !exists {
						securitySchemes[schemeName] = schemeSpec
					}
					multiSecurityReqs = append(multiSecurityReqs, schemeName)
				}

				// In OpenAPI, multiple schemes in the same security requirement means AND logic
				// Multiple security requirements means OR logic
				// So we create separate requirements for each scheme (OR logic)
				for _, schemeName := range multiSecurityReqs {
					securityReq := map[string][]string{
						schemeName: {}, // Empty array means no specific scopes required
					}
					security = append(security, securityReq)
				}
			} else {
				// Regular security scheme
				schemeName, schemeSpec := scheme.GetSecurityScheme()
				if _, exists := securitySchemes[schemeName]; !exists {
					securitySchemes[schemeName] = schemeSpec
				}

				// Add to operation security requirements
				securityReq := map[string][]string{
					schemeName: {}, // Empty array means no specific scopes required
				}
				security = append(security, securityReq)
			}
		}
		operation.Security = security
	}

	// Generate parameters from schema
	if info.SchemaType != nil {
		parameters := extractParameters(info.SchemaType, schemas)
		operation.Parameters = parameters

		// Check for request body
		if requestBody := extractRequestBody(info.SchemaType, schemas); requestBody != nil {
			operation.RequestBody = requestBody
		}
	}

	// Generate responses
	operation.Responses["200"] = generateSuccessResponse(info.ResponseType, schemas)
	operation.Responses["400"] = generateErrorResponse(schemas)

	return operation
}

func extractParameters(schemaType reflect.Type, schemas map[string]*JSONSchema) []Parameter {
	var parameters []Parameter

	// Handle pointers
	if schemaType.Kind() == reflect.Ptr {
		schemaType = schemaType.Elem()
	}

	// Ensure we have a struct type before calling NumField
	if schemaType.Kind() != reflect.Struct {
		return parameters
	}

	// Walk through the schema struct fields
	for i := 0; i < schemaType.NumField(); i++ {
		field := schemaType.Field(i)
		fieldName := strings.ToLower(field.Name)

		switch fieldName {
		case "params":
			// Extract path parameters
			pathParams := extractPathParameters(field.Type, schemas)
			parameters = append(parameters, pathParams...)
		case "query":
			// Extract query parameters
			queryParams := extractQueryParameters(field.Type, schemas)
			parameters = append(parameters, queryParams...)
		default:
			// Check if this field has query tags - treat it as a query parameter
			if queryTag := getTagValue(field, "query"); queryTag != "" {
				paramName := queryTag

				jsonSchema := generateJSONSchemaFromType(field.Type, schemas)

				// Check if parameter has a default value
				if defaultVal := getTagValue(field, "default"); defaultVal != "" {
					jsonSchema.Default = parseDefaultValue(defaultVal, field.Type)
				}

				parameters = append(parameters, Parameter{
					Name:     paramName,
					In:       "query",
					Required: isRequired(field),
					Schema:   jsonSchema,
				})
			} else if isQueryParameter(field) {
				// Auto-detect query parameters based on field characteristics
				paramName := getQueryParameterName(field)

				jsonSchema := generateJSONSchemaFromType(field.Type, schemas)

				// Check if parameter has a default value
				if defaultVal := getTagValue(field, "default"); defaultVal != "" {
					jsonSchema.Default = parseDefaultValue(defaultVal, field.Type)
				}

				parameters = append(parameters, Parameter{
					Name:     paramName,
					In:       "query",
					Required: isRequired(field),
					Schema:   jsonSchema,
				})
			}
		}
	}

	return parameters
}

func extractPathParameters(paramType reflect.Type, schemas map[string]*JSONSchema) []Parameter {
	var parameters []Parameter

	// Handle pointers
	if paramType.Kind() == reflect.Ptr {
		paramType = paramType.Elem()
	}

	// Ensure we have a struct type before calling NumField
	if paramType.Kind() != reflect.Struct {
		return parameters
	}

	for i := 0; i < paramType.NumField(); i++ {
		field := paramType.Field(i)

		paramName := getTagValue(field, "param")
		if paramName == "" {
			paramName = strings.ToLower(field.Name)
		}

		jsonSchema := generateJSONSchemaFromType(field.Type, schemas)

		parameters = append(parameters, Parameter{
			Name:     paramName,
			In:       "path",
			Required: true, // Path parameters are always required
			Schema:   jsonSchema,
		})
	}

	return parameters
}

func extractQueryParameters(queryType reflect.Type, schemas map[string]*JSONSchema) []Parameter {
	var parameters []Parameter

	// Handle pointers
	if queryType.Kind() == reflect.Ptr {
		queryType = queryType.Elem()
	}

	// Ensure we have a struct type before calling NumField
	if queryType.Kind() != reflect.Struct {
		return parameters
	}

	for i := 0; i < queryType.NumField(); i++ {
		field := queryType.Field(i)

		paramName := getTagValue(field, "query")
		if paramName == "" {
			paramName = strings.ToLower(field.Name)
		}

		jsonSchema := generateJSONSchemaFromType(field.Type, schemas)

		// Check if parameter has a default value
		if defaultVal := getTagValue(field, "default"); defaultVal != "" {
			jsonSchema.Default = parseDefaultValue(defaultVal, field.Type)
		}

		parameters = append(parameters, Parameter{
			Name:     paramName,
			In:       "query",
			Required: isRequired(field),
			Schema:   jsonSchema,
		})
	}

	return parameters
}

func extractRequestBody(schemaType reflect.Type, schemas map[string]*JSONSchema) *RequestBody {
	// Handle pointers
	if schemaType.Kind() == reflect.Ptr {
		schemaType = schemaType.Elem()
	}

	// Ensure we have a struct type before calling NumField
	if schemaType.Kind() != reflect.Struct {
		return nil
	}

	// Look for a "Body" field in the schema
	for i := 0; i < schemaType.NumField(); i++ {
		field := schemaType.Field(i)
		if strings.ToLower(field.Name) == "body" {
			jsonSchema := generateJSONSchemaFromType(field.Type, schemas)

			return &RequestBody{
				Description: "Request body",
				Content: map[string]MediaType{
					"application/json": {
						Schema: jsonSchema,
					},
				},
				Required: hasRequiredFields(field.Type),
			}
		}
	}

	return nil
}

func generateSuccessResponse(responseType reflect.Type, schemas map[string]*JSONSchema) Response {
	if responseType == nil {
		return Response{
			Description: "Success",
		}
	}

	// Generate schema for the success result wrapper
	properties := map[string]*JSONSchema{
		"success": {
			Type:    "boolean",
			Default: true,
		},
		"data": generateJSONSchemaFromType(responseType, schemas),
		"error": {
			Type:    "null",
			Default: nil,
		},
	}

	successSchema := newJSONSchema("object", properties)
	successSchema.Required = []string{"success", "data", "error"}

	return Response{
		Description: "Success",
		Content: map[string]MediaType{
			"application/json": {
				Schema: successSchema,
			},
		},
	}
}

func generateErrorResponse(schemas map[string]*JSONSchema) Response {
	// Generate schema for error object
	errorObjProperties := map[string]*JSONSchema{
		"code": {
			Type: "string",
		},
		"message": {
			Type: "string",
		},
	}
	errorObj := newJSONSchema("object", errorObjProperties)
	errorObj.Required = []string{"code", "message"}

	// Generate schema for error result wrapper
	properties := map[string]*JSONSchema{
		"success": {
			Type:    "boolean",
			Default: false,
		},
		"error": errorObj,
		"data": {
			Type:    "null",
			Default: nil,
		},
	}

	errorSchema := newJSONSchema("object", properties)
	errorSchema.Required = []string{"success", "error", "data"}

	return Response{
		Description: "Error",
		Content: map[string]MediaType{
			"application/json": {
				Schema: errorSchema,
			},
		},
	}
}

func generateJSONSchemaFromType(t reflect.Type, schemas map[string]*JSONSchema) *JSONSchema {
	return generateJSONSchemaFromTypeWithContext(t, schemas, "")
}

func generateJSONSchemaFromTypeWithContext(t reflect.Type, schemas map[string]*JSONSchema, contextName string) *JSONSchema {
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return newJSONSchema("string", nil)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return newJSONSchema("integer", nil)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema := newJSONSchema("integer", nil)
		schema.Minimum = floatPtr(0)
		return schema
	case reflect.Float32, reflect.Float64:
		return newJSONSchema("number", nil)
	case reflect.Bool:
		return newJSONSchema("boolean", nil)
	case reflect.Slice, reflect.Array:
		schema := newJSONSchema("array", nil)
		schema.Items = generateJSONSchemaFromTypeWithContext(t.Elem(), schemas, contextName+"Item")
		return schema
	case reflect.Struct:
		return generateStructSchemaWithContext(t, schemas, contextName)
	default:
		return newJSONSchema("object", nil)
	}
}

// newJSONSchema creates a new JSONSchema with only the necessary fields
func newJSONSchema(schemaType string, properties map[string]*JSONSchema) *JSONSchema {
	schema := &JSONSchema{
		Type: schemaType,
	}
	if len(properties) > 0 {
		schema.Properties = properties
	}
	return schema
}

func generateStructSchema(t reflect.Type, schemas map[string]*JSONSchema) *JSONSchema {
	return generateStructSchemaWithContext(t, schemas, "")
}

func generateStructSchemaWithContext(t reflect.Type, schemas map[string]*JSONSchema, contextName string) *JSONSchema {
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Ensure we have a struct type before calling NumField
	if t.Kind() != reflect.Struct {
		return newJSONSchema("object", nil)
	}

	// Create a reference name for the schema
	schemaName := t.Name()
	if schemaName == "" {
		if contextName != "" {
			// Use context name for anonymous structs
			schemaName = contextName
		} else {
			schemaName = "AnonymousStruct"
		}
	}

	// Check if we already have this schema
	if _, exists := schemas[schemaName]; exists {
		return &JSONSchema{Ref: "#/components/schemas/" + schemaName}
	}

	// Create the schema
	properties := make(map[string]*JSONSchema)
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON field name
		jsonName := getJSONFieldName(field)
		if jsonName == "-" {
			continue
		}

		// Generate schema for field with context for anonymous structs
		fieldContextName := ""
		if field.Type.Kind() == reflect.Struct && field.Type.Name() == "" {
			// For anonymous structs, create a name based on the parent schema and field name
			parentName := schemaName
			if parentName == "AnonymousStruct" {
				parentName = contextName
			}
			// Capitalize the first letter of the field name for proper schema naming
			capitalizedJsonName := strings.ToUpper(jsonName[:1]) + jsonName[1:]
			fieldContextName = parentName + capitalizedJsonName
		}
		fieldSchema := generateJSONSchemaFromTypeWithContext(field.Type, schemas, fieldContextName)

		// Add validation constraints from tags
		addValidationConstraints(fieldSchema, field)

		properties[jsonName] = fieldSchema

		// Check if field is required
		if isRequired(field) {
			required = append(required, jsonName)
		}
	}

	// Create the schema with only necessary fields
	schema := newJSONSchema("object", properties)
	if len(required) > 0 {
		schema.Required = required
	}

	// Store the schema in components
	schemas[schemaName] = schema

	// Return a reference to the schema
	return &JSONSchema{Ref: "#/components/schemas/" + schemaName}
}

func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return strings.ToLower(field.Name)
	}

	parts := strings.Split(jsonTag, ",")
	if parts[0] == "" {
		return strings.ToLower(field.Name)
	}

	return parts[0]
}

func addValidationConstraints(schema *JSONSchema, field reflect.StructField) {
	validateTag := field.Tag.Get("validate")
	if validateTag == "" {
		return
	}

	// Parse validation constraints
	constraints := strings.Split(validateTag, ",")
	for _, constraint := range constraints {
		constraint = strings.TrimSpace(constraint)

		if strings.HasPrefix(constraint, "min=") {
			if val, err := strconv.ParseFloat(constraint[4:], 64); err == nil {
				if schema.Type == "string" {
					schema.MinLength = intPtr(int(val))
				} else {
					schema.Minimum = &val
				}
			}
		} else if strings.HasPrefix(constraint, "max=") {
			if val, err := strconv.ParseFloat(constraint[4:], 64); err == nil {
				if schema.Type == "string" {
					schema.MaxLength = intPtr(int(val))
				} else {
					schema.Maximum = &val
				}
			}
		} else if constraint == "email" {
			schema.Format = "email"
		}
	}
}

func generateSummary(method, path string) string {
	// Convert path parameters to readable format (handle both :param and {param} formats)
	readablePath := regexp.MustCompile(`[:{][^/}]+[}]?`).ReplaceAllString(path, "by ID")

	switch strings.ToUpper(method) {
	case "GET":
		return fmt.Sprintf("Get %s", readablePath)
	case "POST":
		return fmt.Sprintf("Create %s", readablePath)
	case "PUT":
		return fmt.Sprintf("Update %s", readablePath)
	case "DELETE":
		return fmt.Sprintf("Delete %s", readablePath)
	case "PATCH":
		return fmt.Sprintf("Patch %s", readablePath)
	default:
		return fmt.Sprintf("%s %s", method, readablePath)
	}
}

func parseDefaultValue(defaultVal string, fieldType reflect.Type) interface{} {
	switch fieldType.Kind() {
	case reflect.String:
		return defaultVal
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, err := strconv.ParseInt(defaultVal, 10, 64); err == nil {
			return val
		}
	case reflect.Float32, reflect.Float64:
		if val, err := strconv.ParseFloat(defaultVal, 64); err == nil {
			return val
		}
	case reflect.Bool:
		if val, err := strconv.ParseBool(defaultVal); err == nil {
			return val
		}
	}
	return defaultVal
}

func writeSwaggerFile(spec *OpenAPISpec, filename string, format OutputFormat) error {
	var data []byte
	var err error
	if format == OutputFormatJSON {
		data, err = json.MarshalIndent(spec, "", "  ")
	} else {
		data, err = yaml.Marshal(spec)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

// isQueryParameter determines if a field should be treated as a query parameter
func isQueryParameter(field reflect.StructField) bool {
	// Skip if it's a nested struct (these should be handled as body or explicit Query/Params fields)
	if field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
		return false
	}

	// Skip if it's a slice of structs (these should be in body)
	if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
		return false
	}

	// Check if field type is suitable for query parameters (primitives, strings, slices of primitives)
	switch field.Type.Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return true
	case reflect.Slice, reflect.Array:
		elemType := field.Type.Elem()
		switch elemType.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.Bool:
			return true
		}
	case reflect.Ptr:
		// Handle pointer to primitive types
		elemType := field.Type.Elem()
		switch elemType.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.Bool:
			return true
		}
	}

	return false
}

// getQueryParameterName extracts the query parameter name from field tags or field name
func getQueryParameterName(field reflect.StructField) string {
	// First check for explicit query tag
	if queryName := getTagValue(field, "query"); queryName != "" {
		return queryName
	}

	// Then check json tag
	if jsonName := getJSONFieldName(field); jsonName != "" && jsonName != "-" {
		return jsonName
	}

	// Fall back to lowercase field name
	return strings.ToLower(field.Name)
}

// convertGinPathToOpenAPI converts Gin path format (:param) to OpenAPI format ({param})
func convertGinPathToOpenAPI(ginPath string) string {
	// Use regex to replace :param with {param}
	re := regexp.MustCompile(`:([^/]+)`)
	return re.ReplaceAllString(ginPath, "{$1}")
}
