package filestore

import (
	"log"
	"os"
	"path/filepath"
)

var _ FileStore = &FsFileStore{}

type FsFileStore struct {
	rootDir string
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func pathInRootDir(rootDir, filePath string) bool {
	if len(filePath) < len(rootDir) {
		return false
	}

	for i := range rootDir {
		if filePath[i] != rootDir[i] {
			return false
		}
	}

	return true
}

func NewFsFileStore(rootDir string) (*FsFileStore, error) {
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}
	if !pathExists(rootDir) {
		return nil, ErrDirNotFound
	}

	return &FsFileStore{
		rootDir: rootDir,
	}, nil
}

func (f *FsFileStore) DeleteFile(filePath string) error {
	path, err := f.GetFilePath(filePath)
	if err != nil {
		return err
	}

	return os.Remove(path)
}

func (f *FsFileStore) GetFileEtag(filePath string) (string, error) {
	path, err := f.GetFilePath(filePath)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return getEtag(data), nil
}

func (f *FsFileStore) LoadFile(filePath string) ([]byte, error) {
	path, err := f.GetFilePath(filePath)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(path)
}

func (f *FsFileStore) RenameFile(filePath string) error {
	panic("unimplemented")
}

func (f *FsFileStore) SaveFile(filePath string, data []byte) error {
	path := filepath.Join(f.rootDir, filePath)
	if !pathInRootDir(f.rootDir, path) {
		return ErrFileNotFound
	}

	baseDir := filepath.Dir(path)
	if !pathExists(baseDir) {
		if err := os.MkdirAll(baseDir, 0777); err != nil {
			return err
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := file.Chmod(0640); err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Println("Unexpected error:", err)
		}
	}()
	_, err = file.Write(data)

	return err
}

func (f *FsFileStore) GetFilePath(filePath string) (string, error) {
	path := filepath.Join(f.rootDir, filePath)
	if !pathInRootDir(f.rootDir, path) || !pathExists(path) {
		return "", ErrFileNotFound
	}
	return path, nil
}
