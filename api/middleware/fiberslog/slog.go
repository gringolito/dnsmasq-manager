package fiberslog

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slog"
)

// New returns a fiber.Handler (middleware) that logs requests using slog.
// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Set PID once
	pid := os.Getpid()

	// put ignore uri into a map for faster match
	skipURIs := make(map[string]struct{})
	for _, uri := range cfg.SkipURIs {
		skipURIs[uri] = struct{}{}
	}

	// Return new handler
	return func(c *fiber.Ctx) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// skip uri
		if _, ok := skipURIs[c.Path()]; ok {
			return c.Next()
		}

		start := time.Now()

		err := c.Next()

		end := time.Now()

		var logMessage string
		var logLevel slog.Level
		statusCode := c.Response().StatusCode()
		switch {
		case statusCode >= 500:
			logLevel = slog.LevelError
			logMessage = "Server error"
		case statusCode >= 400:
			logLevel = slog.LevelWarn
			logMessage = "Client error"
		default:
			logLevel = slog.LevelInfo
			logMessage = "Success"
		}

		// Log disabled for the current level, short-circuit return
		if !cfg.Logger.Enabled(c.UserContext(), logLevel) {
			return err
		}

		// Add fields
		fields := make([]slog.Attr, 0, len(cfg.Fields)+1)
		if err != nil {
			fields = append(fields, slog.String("error", err.Error()))
		}

		for _, field := range cfg.Fields {
			switch field {
			case "referer":
				fields = append(fields, slog.String("referer", c.Get(fiber.HeaderReferer)))
			case "protocol":
				fields = append(fields, slog.String("protocol", c.Protocol()))
			case "pid":
				fields = append(fields, slog.Int("pid", pid))
			case "port":
				fields = append(fields, slog.String("port", c.Port()))
			case "ip":
				fields = append(fields, slog.String("ip", c.IP()))
			case "ips":
				fields = append(fields, slog.String("ips", c.Get(fiber.HeaderXForwardedFor)))
			case "host":
				fields = append(fields, slog.String("host", c.Hostname()))
			case "path":
				fields = append(fields, slog.String("path", c.Path()))
			case "url":
				fields = append(fields, slog.String("url", c.OriginalURL()))
			case "user-agent":
				fields = append(fields, slog.String("user-agent", string(c.Context().UserAgent())))
			case "latency":
				fields = append(fields, slog.Duration("latency", end.Sub(start)))
			case "status":
				fields = append(fields, slog.Int("status", c.Response().StatusCode()))
			case "responseBody":
				if cfg.SkipResBody == nil || !cfg.SkipResBody(c) {
					if cfg.GetResBody == nil {
						fields = append(fields, slog.Any("responseBody", c.Response().Body()))
					} else {
						fields = append(fields, slog.Any("responseBody", cfg.GetResBody(c)))
					}
				}
			case "queryParams":
				fields = append(fields, slog.String("queryParams", c.Request().URI().QueryArgs().String()))
			case "body":
				if cfg.SkipBody == nil || !cfg.SkipBody(c) {
					fields = append(fields, slog.Any("body", c.Body()))
				}
			case "bytesReceived":
				fields = append(fields, slog.Int("bytesReceived", len(c.Request().Body())))
			case "bytesSent":
				fields = append(fields, slog.Int("bytesSent", len(c.Response().Body())))
			case "route":
				fields = append(fields, slog.String("route", c.Route().Path))
			case "method":
				fields = append(fields, slog.String("method", c.Method()))
			case "requestId":
				fields = append(fields, slog.String("requestId", c.GetRespHeader(fiber.HeaderXRequestID)))
			case "requestHeaders":
				c.Request().Header.VisitAll(func(k, v []byte) {
					fields = append(fields, slog.Any(string(k), v))
				})
			}

		}

		cfg.Logger.LogAttrs(c.UserContext(), logLevel, logMessage, fields...)

		return err
	}
}
