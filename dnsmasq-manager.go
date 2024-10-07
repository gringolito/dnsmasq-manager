package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gringolito/dnsmasq-manager/api"
	"github.com/gringolito/dnsmasq-manager/api/handler"
	"github.com/gringolito/dnsmasq-manager/config"
	"github.com/gringolito/dnsmasq-manager/pkg/host"
	"golang.org/x/exp/slog"
)

const (
	AppName    = "dnsmasq Manager"
	AppVersion = "v0.1.0"
)

// These properties are variables because release builds will change their values
var (
	BuildMode = "Development"
	// This file path will be changed to a proper absolute path on the OS directory tree
	OpenApiSpecFile = "api/spec/openapi.yaml"
)

func setupLogger(cfg *config.Config) *slog.Logger {
	// Defaults to slog.LevelInfo
	logLevel := map[string]slog.Level{
		config.LogLevelError:   slog.LevelError,
		config.LogLevelWarning: slog.LevelWarn,
		config.LogLevelInfo:    slog.LevelInfo,
		config.LogLevelDebug:   slog.LevelDebug,
	}

	options := &slog.HandlerOptions{
		AddSource: cfg.Log.Source,
		Level:     logLevel[cfg.Log.Level],
	}

	var output io.Writer = os.Stdout
	if cfg.Log.File != "" {
		logFile, err := os.OpenFile(cfg.Log.File, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
		if err != nil {
			log.Fatal("Failed to create log output file", err)
		}

		output = logFile
	}

	var handler slog.Handler
	if cfg.Log.Format == config.LogFormatPlainText {
		handler = slog.NewTextHandler(output, options)
	} else {
		handler = slog.NewJSONHandler(output, options)
	}

	handler = handler.WithAttrs([]slog.Attr{
		slog.Group("app",
			slog.String("name", AppName),
			slog.String("version", AppVersion),
		),
	})

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}

func addStaticHostApi(router api.Router, cfg *config.Config) {
	hostRepository := host.NewRepository(cfg.Host.Static.File)
	hostService := host.NewService(hostRepository)
	handler.RouteStaticHosts(router, hostService)
}

func main() {
	configName := "test"
	cfg, err := config.Init(configName)
	if err != nil {
		log.Fatal(err)
	}

	logger := setupLogger(cfg)
	logger.Info("Starting app", slog.String("config", configName))

	app := fiber.New(fiber.Config{
		CaseSensitive:     true,
		EnablePrintRoutes: true,
		AppName:           fmt.Sprintf("%s %s (%s build)", AppName, AppVersion, BuildMode),
	})

	middleware, err := api.NewMiddleware(logger, cfg)
	if err != nil {
		logger.Error(err.Error(), slog.String("config", configName))
		os.Exit(1)
	}

	router := api.NewRouter(app, middleware)
	router.AddSwaggerUIRoute(OpenApiSpecFile)
	router.AddMetricsRoute(monitor.Config{
		Title: fmt.Sprintf("%s Monitor", AppName),
	})
	addStaticHostApi(router, cfg)

	if err := app.Listen(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		logger.Error(err.Error(), slog.Int("listeningPort", cfg.Server.Port))
		os.Exit(1)
	}
}
