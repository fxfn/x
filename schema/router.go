package schema

import (
	"github.com/gin-gonic/gin"
)

// RouterHelper provides methods to register routes with automatic type registration
type RouterHelper struct {
	*gin.Engine
}

// RouterGroup provides methods to register routes with automatic type registration within a group
type RouterGroup struct {
	*gin.RouterGroup
	groupSecuritySchemes []SecurityScheme
}

// NewRouter creates a new RouterHelper that wraps gin.Engine
func NewRouter() *RouterHelper {
	return &RouterHelper{gin.Default()}
}

// WrapRouter wraps an existing gin.Engine with RouterHelper functionality
func WrapRouter(engine *gin.Engine) *RouterHelper {
	return &RouterHelper{engine}
}

// Use adds middleware to the router
func (r *RouterHelper) Use(middleware ...gin.HandlerFunc) gin.IRoutes {
	return r.Engine.Use(middleware...)
}

// UseSecurity adds security middleware to the router (use with caution - applies globally)
func (r *RouterHelper) UseSecurity(schemes ...SecurityScheme) gin.IRoutes {
	var middlewares []gin.HandlerFunc
	for _, scheme := range schemes {
		middlewares = append(middlewares, scheme.Middleware())
	}
	return r.Engine.Use(middlewares...)
}

// Group creates a new route group with the given path prefix
func (r *RouterHelper) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{
		RouterGroup:          r.Engine.Group(relativePath, handlers...),
		groupSecuritySchemes: []SecurityScheme{},
	}
}

// Use adds middleware to the route group with automatic security detection
func (rg *RouterGroup) Use(middleware ...gin.HandlerFunc) gin.IRoutes {
	// Scan middleware for security schemes using reflection
	for _, handler := range middleware {
		if scheme, isSecurityMiddleware := IsSecurityMiddleware(handler); isSecurityMiddleware {
			rg.groupSecuritySchemes = append(rg.groupSecuritySchemes, scheme)
		}
	}

	return rg.RouterGroup.Use(middleware...)
}

// processHandlers processes a list of handlers and separates them by type
func processHandlers(method, path string, handlers []interface{}) ([]gin.HandlerFunc, TypedHandlerFunc, bool) {
	var middlewares []gin.HandlerFunc
	var securitySchemes []SecurityScheme
	var typedHandler TypedHandlerFunc
	var hasTypedHandler bool

	// Process all handlers to separate middleware and typed handlers
	for _, h := range handlers {
		switch v := h.(type) {
		case SecurityScheme:
			securitySchemes = append(securitySchemes, v)
			middlewares = append(middlewares, v.Middleware())
		case TypedHandlerFunc:
			typedHandler = v
			hasTypedHandler = true
			middlewares = append(middlewares, v.HandlerFunc())
		case gin.HandlerFunc:
			middlewares = append(middlewares, v)
		case func(*gin.Context):
			middlewares = append(middlewares, gin.HandlerFunc(v))
		}
	}

	// Register typed handler if present
	if hasTypedHandler {
		RegisterTypedHandler(method, path, typedHandler)
	}

	// Register security schemes
	if len(securitySchemes) > 0 {
		RegisterSecurityScheme(method, path, securitySchemes...)
	}

	return middlewares, typedHandler, hasTypedHandler
}

// GET registers a GET route with automatic type registration
func (r *RouterHelper) GET(path string, handlers ...interface{}) {
	middlewares, _, _ := processHandlers("GET", path, handlers)
	r.Engine.GET(path, middlewares...)
}

// POST registers a POST route with automatic type registration
func (r *RouterHelper) POST(path string, handlers ...interface{}) {
	middlewares, _, _ := processHandlers("POST", path, handlers)
	r.Engine.POST(path, middlewares...)
}

// PUT registers a PUT route with automatic type registration
func (r *RouterHelper) PUT(path string, handlers ...interface{}) {
	middlewares, _, _ := processHandlers("PUT", path, handlers)
	r.Engine.PUT(path, middlewares...)
}

// DELETE registers a DELETE route with automatic type registration
func (r *RouterHelper) DELETE(path string, handlers ...interface{}) {
	middlewares, _, _ := processHandlers("DELETE", path, handlers)
	r.Engine.DELETE(path, middlewares...)
}

// PATCH registers a PATCH route with automatic type registration
func (r *RouterHelper) PATCH(path string, handlers ...interface{}) {
	middlewares, _, _ := processHandlers("PATCH", path, handlers)
	r.Engine.PATCH(path, middlewares...)
}

// RouterGroup HTTP method handlers

// processGroupHandlers processes handlers for a route group
func (rg *RouterGroup) processGroupHandlers(method, path string, handlers []interface{}) []gin.HandlerFunc {
	fullPath := rg.RouterGroup.BasePath() + path
	middlewares, _, _ := processHandlers(method, fullPath, handlers)

	// Register group-level security schemes for this route
	if len(rg.groupSecuritySchemes) > 0 {
		RegisterSecurityScheme(method, fullPath, rg.groupSecuritySchemes...)
	}

	return middlewares
}

// GET registers a GET route with automatic type registration in a route group
func (rg *RouterGroup) GET(path string, handlers ...interface{}) {
	middlewares := rg.processGroupHandlers("GET", path, handlers)
	rg.RouterGroup.GET(path, middlewares...)
}

// POST registers a POST route with automatic type registration in a route group
func (rg *RouterGroup) POST(path string, handlers ...interface{}) {
	middlewares := rg.processGroupHandlers("POST", path, handlers)
	rg.RouterGroup.POST(path, middlewares...)
}

// PUT registers a PUT route with automatic type registration in a route group
func (rg *RouterGroup) PUT(path string, handlers ...interface{}) {
	middlewares := rg.processGroupHandlers("PUT", path, handlers)
	rg.RouterGroup.PUT(path, middlewares...)
}

// DELETE registers a DELETE route with automatic type registration in a route group
func (rg *RouterGroup) DELETE(path string, handlers ...interface{}) {
	middlewares := rg.processGroupHandlers("DELETE", path, handlers)
	rg.RouterGroup.DELETE(path, middlewares...)
}

// PATCH registers a PATCH route with automatic type registration in a route group
func (rg *RouterGroup) PATCH(path string, handlers ...interface{}) {
	middlewares := rg.processGroupHandlers("PATCH", path, handlers)
	rg.RouterGroup.PATCH(path, middlewares...)
}
