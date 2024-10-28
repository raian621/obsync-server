package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		configText string
		wantConfig Config
		wantErr    error
	}{
		{
			name: "load config successfully",
			configText: `type: FileSystem
root: /tmp/obsync-dev
host: localhost
port: 8000`,
			wantConfig: Config{
				Type: "FileSystem",
				Root: "/tmp/obsync-dev",
				Host: "localhost",
				Port: 8000,
			},
		},
		{
			name: "incorrect filestore type",
			configText: `type: CarrierPigeon
root: /tmp/obsync-dev
host: localhost
port: 8000`,
			wantErr: ErrUnsupportedFileStoreType,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.configText)
			config, err := ReadConfig(buf)
			assert.ErrorIs(t, tc.wantErr, err)
			if err == nil {
				assert.Equal(t, tc.wantConfig, *config)
			}
		})
	}
}

func TestReadConfigFromFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		configText string
		wantConfig Config
		wantErr    error
	}{
		{
			name: "load config successfully",
			configText: `type: FileSystem
root: /tmp/obsync-dev
host: localhost
port: 8000`,
			wantConfig: Config{
				Type: "FileSystem",
				Root: "/tmp/obsync-dev",
				Host: "localhost",
				Port: 8000,
			},
		},
		{
			name: "incorrect filestore type",
			configText: `type: CarrierPigeon
root: /tmp/obsync-dev
host: localhost
port: 8000`,
			wantErr: ErrUnsupportedFileStoreType,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			file, err := os.CreateTemp(t.TempDir(), "test.yaml")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer func() {
				if err := file.Close(); err != nil {
					t.Log(err)
				}
				if err := os.Remove(file.Name()); err != nil {
					t.Log(err)
				}
			}()
			_, err = file.WriteString(tc.configText)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			config, err := ReadConfigFromFile(file.Name())
			assert.ErrorIs(t, tc.wantErr, err)
			if err == nil {
				assert.Equal(t, tc.wantConfig, *config)
			}
		})
	}
}
