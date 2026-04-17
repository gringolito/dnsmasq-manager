package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gringolito/dnsmasq-manager/api/presenter"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

const (
	NotAuthorizedMessage = "You do not have permission to access this resource."
	MalformedJwt         = "The user provided a malformed JWT."
	MissingRole          = "The user does not have the required role to access this resource."
)

func authorizationHandler(jwtContextKey string, roles []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals(jwtContextKey).(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		name, ok := claims["name"].(string)
		if !ok {
			return presenter.ForbiddenResponse(c, NotAuthorizedMessage, MalformedJwt)
		}

		scopes, err := parseClaimScope(claims)
		if err != nil {
			slog.Debug("Authorization denied: malformed JWT claim 'scope'",
				slog.String("user", name),
				slog.String("error", err.Error()),
			)
			return presenter.ForbiddenResponse(c, NotAuthorizedMessage, MalformedJwt)
		}

		authorized := false
		for _, role := range roles {
			if slices.Contains[string](scopes, role) {
				authorized = true
				slog.Debug("Authorization granted",
					slog.String("user", name),
					slog.String("role", role),
				)
				break
			}
		}

		if !authorized {
			slog.Debug("Authorization denied: required role not found",
				slog.String("user", name),
			)
			return presenter.ForbiddenResponse(c, NotAuthorizedMessage, MissingRole)
		}

		return c.Next()
	}
}

func parseClaimScope(claims jwt.MapClaims) ([]string, error) {
	var scopes []string

	switch v := claims["scope"].(type) {
	case string:
		scopes = strings.Fields(v)
	case []string:
		scopes = v
	case []interface{}:
		for _, a := range v {
			scope, ok := a.(string)
			if !ok {
				return nil, jwt.ErrInvalidType
			}
			scopes = append(scopes, scope)
		}
	default:
		return nil, jwt.ErrInvalidType
	}

	return scopes, nil
}
