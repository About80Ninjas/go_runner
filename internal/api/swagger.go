// internal/api/swagger.go
package api

import (
	"encoding/json"
	"net/http"
)

// swaggerUIHandler serves the Swagger UI
func (s *Server) swaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	swaggerHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Binary Executor API Documentation</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui.min.css">
    <style>
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui-bundle.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui-standalone-preset.min.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/api/v1/openapi.json",
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout",
                deepLinking: true
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(swaggerHTML))
}

// openAPIHandler returns the OpenAPI specification
func (s *Server) openAPIHandler(w http.ResponseWriter, r *http.Request) {
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Binary Executor API",
			"description": "API for managing and executing Go binaries from Git repositories",
			"version":     "1.0.0",
		},
		"servers": []map[string]interface{}{
			{
				"url":         "/api/v1",
				"description": "API Server",
			},
		},
		"paths": map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Health Check",
					"description": "Returns the health status of the service",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service is healthy",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"status":    map[string]string{"type": "string"},
											"timestamp": map[string]string{"type": "string", "format": "date-time"},
										},
									},
								},
							},
						},
					},
				},
			},
			"/binaries": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List Binaries",
					"description": "Returns all managed binaries",
					"security":    []map[string][]string{{"bearerAuth": {}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of binaries",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"$ref": "#/components/schemas/Binary",
										},
									},
								},
							},
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Create Binary",
					"description": "Creates a new managed binary",
					"security":    []map[string][]string{{"bearerAuth": {}}},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/BinaryInput",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Binary created",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/Binary",
									},
								},
							},
						},
					},
				},
			},
			"/binaries/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get Binary",
					"description": "Returns a specific binary",
					"security":    []map[string][]string{{"bearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"schema":      map[string]string{"type": "string"},
							"description": "Binary ID",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Binary details",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/Binary",
									},
								},
							},
						},
					},
				},
				"delete": map[string]interface{}{
					"summary":     "Delete Binary",
					"description": "Deletes a binary",
					"security":    []map[string][]string{{"bearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"schema":      map[string]string{"type": "string"},
							"description": "Binary ID",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Binary deleted",
						},
					},
				},
			},
			"/binaries/{id}/build": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Build Binary",
					"description": "Triggers a build for the binary",
					"security":    []map[string][]string{{"bearerAuth": {}}},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"schema":      map[string]string{"type": "string"},
							"description": "Binary ID",
						},
					},
					"responses": map[string]interface{}{
						"202": map[string]interface{}{
							"description": "Build started",
						},
					},
				},
			},
			"/execute": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Execute Binary",
					"description": "Executes a binary with given parameters",
					"security":    []map[string][]string{{"apiKey": {}}},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ExecutionRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Execution result",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ExecutionResult",
									},
								},
							},
						},
					},
				},
			},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"bearerAuth": map[string]interface{}{
					"type":   "http",
					"scheme": "bearer",
				},
				"apiKey": map[string]interface{}{
					"type": "apiKey",
					"in":   "header",
					"name": "X-API-Key",
				},
			},
			"schemas": map[string]interface{}{
				"Binary": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":          map[string]string{"type": "string"},
						"name":        map[string]string{"type": "string"},
						"description": map[string]string{"type": "string"},
						"repo_url":    map[string]string{"type": "string"},
						"branch":      map[string]string{"type": "string"},
						"build_path":  map[string]string{"type": "string"},
						"status":      map[string]string{"type": "string", "enum": "pending,building,ready,failed"},
						"version":     map[string]string{"type": "string"},
						"last_built":  map[string]string{"type": "string", "format": "date-time"},
						"created_at":  map[string]string{"type": "string", "format": "date-time"},
						"updated_at":  map[string]string{"type": "string", "format": "date-time"},
					},
				},
				"BinaryInput": map[string]interface{}{
					"type":     "object",
					"required": []string{"name", "repo_url", "branch"},
					"properties": map[string]interface{}{
						"name":        map[string]string{"type": "string"},
						"description": map[string]string{"type": "string"},
						"repo_url":    map[string]string{"type": "string"},
						"branch":      map[string]string{"type": "string"},
						"build_path":  map[string]string{"type": "string"},
					},
				},
				"ExecutionRequest": map[string]interface{}{
					"type":     "object",
					"required": []string{"binary_id"},
					"properties": map[string]interface{}{
						"binary_id": map[string]string{"type": "string"},
						"args": map[string]interface{}{
							"type":  "array",
							"items": map[string]string{"type": "string"},
						},
						"env": map[string]interface{}{
							"type":  "array",
							"items": map[string]string{"type": "string"},
						},
						"stdin":   map[string]string{"type": "string"},
						"timeout": map[string]string{"type": "integer", "description": "Timeout in seconds"},
					},
				},
				"ExecutionResult": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":          map[string]string{"type": "string"},
						"binary_id":   map[string]string{"type": "string"},
						"status":      map[string]string{"type": "string", "enum": "running,completed,failed,timeout"},
						"exit_code":   map[string]string{"type": "integer"},
						"stdout":      map[string]string{"type": "string"},
						"stderr":      map[string]string{"type": "string"},
						"started_at":  map[string]string{"type": "string", "format": "date-time"},
						"finished_at": map[string]string{"type": "string", "format": "date-time"},
						"duration_ms": map[string]string{"type": "integer"},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}
