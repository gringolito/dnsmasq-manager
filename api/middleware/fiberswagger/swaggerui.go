package fiberswagger

import (
	"github.com/gofiber/fiber/v2"
	upstream "github.com/gringolito/fiberswagger"
)

// Middleware returns a fiber.Handler (middleware) that renders OpenAPI specification using SwaggerUI.
// Deprecated: use github.com/gringolito/fiberswagger.Middleware directly.
func Middleware(config ...Config) fiber.Handler {
	return upstream.Middleware(config...)
}

// Router creates routes with handlers to renders OpenAPI specification using SwaggerUI.
// Deprecated: use github.com/gringolito/fiberswagger.Router directly.
func Router(router fiber.Router, config ...Config) {
	upstream.Router(router, config...)
}
