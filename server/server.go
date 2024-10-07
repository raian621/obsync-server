package server

import (
	"database/sql"

	"github.com/raian621/obsync-server/api"
)

type ObsyncServer struct {
	db *sql.DB
}

// check that ObsyncServer implements ServerInterface:
var _ api.ServerInterface = (*ObsyncServer)(nil)

func NewServer(db *sql.DB) *ObsyncServer {
	return &ObsyncServer{
		db,
	}
}
