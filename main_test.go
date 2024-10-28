package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/raian621/obsync-server/config"
	"github.com/raian621/obsync-server/database"
	"github.com/stretchr/testify/assert"
)

func TestStartServer(t *testing.T) {
	var (
		host = "0.0.0.0"
		port = 8001
		url  = fmt.Sprintf("http://%s:%d/api/v1/openapi.yaml", host, port)
	)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go startServer("test-start-server.db?mode=memory", &config.Config{
		Type: "FileSystem",
		Root: t.TempDir(),
		Host: "localhost",
		Port: 8001,
	}, ctx)

	// wait for the server to start up, then try to hit an endpoint
	time.Sleep(time.Second)

	res, err := http.Get(url)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	// wait for server to shutdown, then try to hit an endpoint (request should fail)
	<-ctx.Done()
	time.Sleep(time.Second)

	res, err = http.Get(url)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	database.SetDB(nil)
}
