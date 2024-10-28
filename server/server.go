package server

import (
	"database/sql"

	"github.com/raian621/obsync-server/api"
	"github.com/raian621/obsync-server/filestore"
)

type ObsyncServer struct {
	db     *sql.DB
	fstore filestore.FileStore
}

// check that ObsyncServer implements ServerInterface:
var _ api.ServerInterface = (*ObsyncServer)(nil)

func NewServer(db *sql.DB, rootDir string) (*ObsyncServer, error) {
	fstore, err := filestore.NewFsFileStore(rootDir)
	if err != nil {
		return nil, err
	}
	return &ObsyncServer{
		db,
		fstore,
	}, nil
}
