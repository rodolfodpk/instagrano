// Package docs Simple Swagger documentation for Instagrano API
package docs

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = map[string]interface{}{
	"swagger": "2.0",
	"info": map[string]interface{}{
		"description": "A simple Instagram-like API built with Go and Fiber",
		"title":       "Instagrano API",
		"version":     "1.0",
	},
	"host":     "localhost:3000",
	"basePath": "/",
	"schemes":  []string{"http"},
}
