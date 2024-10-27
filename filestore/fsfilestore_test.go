package filestore

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFsFileStore(t *testing.T) {
	// rootDir exists
	rootDir := t.TempDir()
	fstore, err := NewFsFileStore(rootDir)
	assert.NoError(t, err)
	assert.Equal(t, fstore.rootDir, rootDir)

	// rootDir does not exist
	fstore, err = NewFsFileStore(rootDir + "astringthatshouldntbehere")
	assert.ErrorIs(t, err, ErrDirNotFound)
	assert.Nil(t, fstore)
}

func TestFsFileStoreSaveFile(t *testing.T) {
	rootDir := t.TempDir()
	fstore, err := NewFsFileStore(rootDir)
	assert.NoError(t, err)

	testCases := []struct {
		name    string
		path    string
		data    []byte
		wantErr error
	}{
		{
			name: "file in root directory",
			data: []byte("abcdefghijklmnopqrstuvwxyz"),
			path: "alphabet.txt",
		},
		{
			name: "file in nested directory",
			data: []byte("abcdefghijklmnopqrstuvwxyz"),
			path: "directory/alphabet.txt",
		},
		{
			name:    "handle sneaky relative paths",
			data:    []byte("abcdefghijklmnopqrstuvwxyz"),
			path:    "../directory/alphabet.txt",
			wantErr: ErrFileNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := fstore.SaveFile(tc.path, tc.data)
			assert.ErrorIs(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.NoError(t, fstore.SaveFile(tc.path, tc.data))
			path := filepath.Join(fstore.rootDir, tc.path)
			assert.True(t, pathExists(path))
			data, err := os.ReadFile(path)
			assert.NoError(t, err)
			assert.Equal(t, tc.data, data)
			fileinfo, err := os.Stat(path)
			assert.NoError(t, err)
			assert.Equal(t, fs.FileMode(0640), fileinfo.Mode())
		})
	}
}

func TestFsFileStoreCheckFilepath(t *testing.T) {
	rootDir := filepath.Join("tmp", "directory")
	filePath := filepath.Join("tmp", "directory", "file")
	assert.True(t, pathInRootDir(rootDir, filePath))
	filePath = filepath.Join("tmp", "important", "secret")
	assert.False(t, pathInRootDir(rootDir, filePath))
	filePath = filepath.Join("home", "ryan", "Homework")
	assert.False(t, pathInRootDir(rootDir, filePath))
	filePath = filepath.Join("bin", "bash")
	assert.False(t, pathInRootDir(rootDir, filePath))
}

func TestFsFileGetFileEtag(t *testing.T) {
	rootDir := t.TempDir()
	fstore, err := NewFsFileStore(rootDir)
	assert.NoError(t, err)

	testCases := []struct {
		name      string
		path      string
		data      []byte
		queryPath string
		wantErr   error
	}{
		{
			name:      "etag matches",
			path:      "alphabet.txt",
			data:      []byte("abcdefghijklmnopqrstuvwxyz"),
			queryPath: "alphabet.txt",
		},
		{
			name:      "file not found",
			path:      "alphabet.txt",
			data:      []byte("abcdefghijklmnopqrstuvwxyz"),
			queryPath: "not-alphabet.txt",
			wantErr:   ErrFileNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, fstore.SaveFile(tc.path, tc.data))
			etag, err := fstore.GetFileEtag(tc.queryPath)
			if !assert.ErrorIs(t, tc.wantErr, err) {
				t.FailNow()
			} else if err != nil {
				return
			}
			wantEtag := getEtag(tc.data)
			assert.Equal(t, wantEtag, etag)
			os.Remove(filepath.Join(fstore.rootDir, tc.path))
		})
	}
}

func TestFsFileDeleteFile(t *testing.T) {
	rootDir := t.TempDir()
	fstore, err := NewFsFileStore(rootDir)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// discard sneaky paths
	if !assert.ErrorIs(t, ErrFileNotFound, fstore.DeleteFile("../alphabet.txt")) {
		t.FailNow()
	}

	// file exists
	if !assert.NoError(t, fstore.SaveFile("alphabet.txt", []byte("abcdef"))) {
		t.FailNow()
	}

	// file exists
	if !assert.NoError(t, fstore.DeleteFile("alphabet.txt")) {
		t.FailNow()
	}

	// check that file was actually deleted
	path := filepath.Join(fstore.rootDir, "alphabet.txt")
	if !assert.False(t, pathExists(path)) {
		t.FailNow()
	}

	// file doesn't exist
	assert.ErrorIs(t, ErrFileNotFound, fstore.DeleteFile("alphabet.txt"))
}

func TestFsFileStoreLoadFile(t *testing.T) {
	rootDir := t.TempDir()
	fstore, err := NewFsFileStore(rootDir)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NoError(t, fstore.SaveFile("alphabet.txt", []byte("abcdefg"))) {
		t.FailNow()
	}

	// load data
	data, err := fstore.LoadFile("alphabet.txt")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, []byte("abcdefg"), data) {
		t.FailNow()
	}

	// file doesn't exist
	_, err = fstore.LoadFile("not-alphabet.txt")
	if !assert.ErrorIs(t, ErrFileNotFound, err) {
		t.FailNow()
	}
}
