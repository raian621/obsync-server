package filestore

import "errors"

var (
	ErrFileNotFound = errors.New("file not found at specified filePath")
	ErrDirNotFound  = errors.New("directory not found at specified path")
)

type FileStore interface {
	SaveFile(filePath string, data []byte) error
	LoadFile(filePath string) ([]byte, error)
	RenameFile(filePath string) error
	DeleteFile(filePath string) error
	GetFileEtag(filePath string) (string, error)
	GetFilePath(filePath string) (string, error)
}
