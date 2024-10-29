//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=openapi/config.yaml openapi/openapi.yaml

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/raian621/obsync-server/api"
	"github.com/raian621/obsync-server/config"
	"github.com/raian621/obsync-server/database"
	"github.com/raian621/obsync-server/server"
)

func main() {
	config, err := config.ReadConfigFromFile("config.yaml")
	if err != nil {
		panic(err)
	}
	startServer("sqlite.db", config, context.Background())
}

func startServer(connStr string, cfg *config.Config, serverCtx context.Context) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Logger.SetLevel(log.INFO)
	db, err := database.NewDB(connStr)
	if err != nil {
		e.Logger.Fatal(err)
	}
	database.SetDB(db)
	if err := database.ApplyMigrations(db); err != nil {
		e.Logger.Fatal(err)
	}
	server, err := server.NewServer(db, cfg.Root)
	if err != nil {
		e.Logger.Fatal(err)
	}
	api.RegisterHandlersWithBaseURL(e, server, "/api/v1")

	ctx, stop := signal.NotifyContext(serverCtx, os.Interrupt)
	go func() {
		if err := e.Start(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// wait for the interrupt signal to gracefully shutdown the server after 5
	// seconds
	<-ctx.Done()
	stop()
	if err := db.Close(); err != nil {
		e.Logger.Fatal(err)
	}
	e.Logger.Info("Shutting server down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
