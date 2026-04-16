package fiberslog

import (
	"github.com/gofiber/fiber/v2"
	upstream "github.com/gringolito/fiberslog"
)

// New returns a fiber.Handler (middleware) that logs requests using slog.
// Deprecated: use github.com/gringolito/fiberslog.New directly.
func New(config ...Config) fiber.Handler {
	return upstream.New(config...)
}
