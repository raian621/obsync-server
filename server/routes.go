package server

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/raian621/obsync-server/api"
)

// Get the Redoc OpenAPI documentation page
// (GET /docs)
func (o *ObsyncServer) GetDocs(ctx echo.Context) error {
	return ctx.Blob(200, "text/html", api.RedocPage)
}

// Delete a file on the sync server
// (DELETE /files/{filename})
func (o *ObsyncServer) DeleteFilesFilename(ctx echo.Context, filename string) error {
	return errors.New("not implemented yet")
}

// Download a file from the sync server
// (GET /files/{filename})
func (o *ObsyncServer) GetFilesFilename(ctx echo.Context, filename string, params api.GetFilesFilenameParams) error {
	return errors.New("not implemented yet")
}

// Upload a file to the sync server
// (POST /files/{filename})
func (o *ObsyncServer) PostFilesFilename(ctx echo.Context, filename string) error {
	return errors.New("not implemented yet")
}

// Update a file on the sync server
// (PUT /files/{filename})
func (o *ObsyncServer) PutFilesFilename(ctx echo.Context, filename string, params api.PutFilesFilenameParams) error {
	return errors.New("not implemented yet")
}

// Get a list of files that are synced to the server
// (GET /list-files)
func (o *ObsyncServer) GetListFiles(ctx echo.Context) error {
	return errors.New("not implemented yet")
}

// Get the OpenAPI spec in YAML format
// (GET /openapi.yaml)
func (o *ObsyncServer) GetOpenapiYaml(ctx echo.Context) error {
	return ctx.Blob(200, "text/yaml", api.OpenApiSpec)
}

// Get the Redoc script that's stored locally on the server
// (GET /redoc.standalone.js)
func (o *ObsyncServer) GetRedocStandaloneJs(ctx echo.Context) error {
	return ctx.Blob(200, "application/javascript", api.RedocBundle)
}
