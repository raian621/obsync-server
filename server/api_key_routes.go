package server

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/raian621/obsync-server/api"
)

// Delete an API key
// (DELETE /api-keys)
func (o *ObsyncServer) DeleteApikeys(ctx echo.Context) error {
	return errors.New("not implemented yet")
}

// Get API key info
// (GET /api-keys)
func (o *ObsyncServer) GetApikeys(ctx echo.Context, params api.GetApikeysParams) error {
	return errors.New("not implemented yet")
}

// Create an API key
// (POST /api-keys)
func (o *ObsyncServer) PostApikeys(ctx echo.Context) error {
	return errors.New("not implemented yet")
}
