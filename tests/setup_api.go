package tests

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gringolito/dnsmasq-manager/api"
	"github.com/gringolito/dnsmasq-manager/config"
	"github.com/stretchr/testify/require"
)

func SetupConfig(t *testing.T) *config.Config {
	configName := "unittest"
	cfg, err := config.Init(configName)
	require.NoError(t, err)

	return cfg
}

func SetupApp() *fiber.App {
	return fiber.New(fiber.Config{
		CaseSensitive:     true,
		EnablePrintRoutes: true,
	})
}

func SetupRouter(app *fiber.App, cfg *config.Config) api.Router {
	middleware, _ := api.NewMiddleware(nil, cfg)
	return api.NewRouter(app, middleware)
}
