package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"trox.dev/file-converter/internal/convert"
)

type HTTPConfig struct {
	Port    int
	Address string
}

type Config struct {
	HTTP HTTPConfig
}

type Application struct {
	Config   *Config
	Registry *convert.Registry
	Server   *http.Server
}

func main() {
	slog.Info("Setting up file-converter...")

	app := &Application{
		Config: readConfig(),
	}
	defer app.stop()

	app.setupRegistry()
	app.setupHTTP()

	// graceful shutdown handling
	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("HTTP Server running", "address", app.Server.Addr)
		if err := app.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			slog.Error("http server failed to start", "err", err)
		}
	case <-ctx.Done():
		slog.Info("Received shutdown signal")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Error during HTTP server shutdown", "err", err)
	}

	slog.Info("Application stopped successfully")
}

func readConfig() *Config {
	slog.Info("Reading config...")

	cfg := &Config{
		HTTP: HTTPConfig{
			Port:    8080,
			Address: "0.0.0.0",
		},
	}

	slog.Info("Config initialized")

	return cfg
}

func (app *Application) setupRegistry() {
	slog.Info("Setting up converter registry...")
	app.Registry = convert.NewRegistry()
	app.Registry.Register(convert.NewVipsConverter())

	if err := app.Registry.StartAll(); err != nil {
		slog.Error("Failed to start converter registry.", "err", err)
		os.Exit(1)
	}

	var convNames []string
	for _, conv := range app.Registry.GetAll() {
		convNames = append(convNames, conv.Name())
	}

	slog.Info("Converter registry setup successful.", "converters", convNames)
}

func (app *Application) setupHTTP() {
	app.Server = &http.Server{
		Addr:              app.Config.HTTP.Address + ":" + strconv.Itoa(app.Config.HTTP.Port),
		Handler:           http.NewServeMux(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func (app *Application) stop() {
	if err := app.Registry.StopAll(); err != nil {
		slog.Error("Failed to stop registry", "err", err)
	}
}
