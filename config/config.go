package config

import (
	"bytes"
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	ErrUnsupportedFileStoreType = errors.New("")
)

type Config struct {
	Type string `yaml:"type"`
	Root string `yaml:"root"`
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
}

func ReadConfig(source io.Reader) (*Config, error) {
	decoder := yaml.NewDecoder(source)
	var config Config
	if err := decoder.Decode(&config); err != nil {
		panic(err)
	}

	if config.Type != "FileSystem" {
		return nil, ErrUnsupportedFileStoreType
	}

	return &config, nil
}

func ReadConfigFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return ReadConfig(bytes.NewBuffer(data))
}
