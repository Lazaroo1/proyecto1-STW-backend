package handlers

import (
	"net/http"
	"os"
)

type SwaggerHandler struct{}

func (s *SwaggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/swagger.yaml" {
		data, err := os.ReadFile("swagger.yaml")
		if err != nil {
			http.Error(w, "swagger.yaml not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(data)
		return
	}

	// GET /docs → Swagger UI
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <title>Series Tracker API Docs</title>
  <meta charset="utf-8"/>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
  SwaggerUIBundle({
    url: "/swagger.yaml",
    dom_id: "#swagger-ui",
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout"
  });
</script>
</body>
</html>`))
}
