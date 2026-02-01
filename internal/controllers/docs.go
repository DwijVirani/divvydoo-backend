package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type DocsController struct {
	specPath string
}

func NewDocsController() *DocsController {
	return &DocsController{
		specPath: "openapi.yaml",
	}
}

func (dc *DocsController) GetOpenAPISpec(c *gin.Context) {
	// Serve Swagger UI HTML page
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DivvyDoo API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui.css">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/docs/openapi.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func (dc *DocsController) GetOpenAPIYAML(c *gin.Context) {
	// Read the OpenAPI spec file
	specData, err := os.ReadFile(dc.specPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read OpenAPI specification",
		})
		return
	}

	// Set proper headers to display inline
	c.Header("Content-Type", "application/x-yaml; charset=utf-8")
	c.Header("Content-Disposition", "inline")
	c.String(http.StatusOK, string(specData))
}
