package openapi

import _ "embed"

var (
	//go:embed openapi.yaml
	OpenApiSpec []byte
	//go:embed redoc.html
	RedocPage []byte
	//go:embed redoc.standalone.js
	RedocBundle []byte
)
