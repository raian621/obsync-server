package server

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/raian621/obsync-server/api"
	"github.com/raian621/obsync-server/openapi"
)

// Create an API key
// (DELETE /eys)
func (o *ObsyncServer) DeleteApikeys(ctx echo.Context) error {
	return errors.New("not implemented yet")
}

// Get API key info
// (GET /eys)
func (o *ObsyncServer) GetApikeys(ctx echo.Context, params api.GetApikeysParams) error {
	return errors.New("not implemented yet")
}

// Create an API key
// (POST /eys)
func (o *ObsyncServer) PostApikeys(ctx echo.Context) error {
	return errors.New("not implemented yet")
}

// Get the Redoc OpenAPI documentation page
// (GET /docs)
func (o *ObsyncServer) GetDocs(ctx echo.Context) error {
	return ctx.Blob(200, "text/html", openapi.RedocPage)
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
	return ctx.Blob(200, "text/yaml", openapi.OpenApiSpec)
}

// Get the Redoc script that's stored locally on the server
// (GET /redoc.standalone.js)
func (o *ObsyncServer) GetRedocStandaloneJs(ctx echo.Context) error {
	return ctx.Blob(200, "application/javascript", openapi.RedocBundle)
}

// Delete a user
// (DELETE /user)
func (o *ObsyncServer) DeleteUser(ctx echo.Context) error {
	return errors.New("not implemented yet")
}

// Create a user
// (POST /user)
func (o *ObsyncServer) PostUser(ctx echo.Context) error {
	return errors.New("not implemented yet")
}

// Update a user's information
// (PUT /user)
func (o *ObsyncServer) PutUser(ctx echo.Context) error {
	return errors.New("not implemented yet")
}

// Log in a user
// (POST /user/login)
func (o *ObsyncServer) PostUserLogin(ctx echo.Context) error {
	return errors.New("not implemented yet")
}

// Log out a user
// (POST /user/logout)
func (o *ObsyncServer) PostUserLogout(ctx echo.Context) error {
	return errors.New("not implemented yet")
}

// Let users update their email
// (PUT /user/email)
func (o *ObsyncServer) PutUserEmail(ctx echo.Context) error {
	return errors.New("not implemented yet")
}

// Let users update their password
// (PUT /user/password)
func (o *ObsyncServer) PutUserPassword(ctx echo.Context) error {
	return errors.New("not implemented yet")
}

// Let users update their username
// (PUT /user/username)
func (o *ObsyncServer) PutUserUsername(ctx echo.Context) error {
	return errors.New("not implemented yet")
}
