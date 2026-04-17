package presenter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInternalServerErrorResponseWithoutRequestId(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return InternalServerErrorResponse(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), result["error"])
	assert.Equal(t, ServerErrorMessage, result["message"])
	assert.Equal(t, fmt.Sprintf(InternalServerError, "unknown"), result["details"])
}

func TestInternalServerErrorResponseWithRequestId(t *testing.T) {
	const testRequestId = "550e8400-e29b-41d4-a716-446655440000"

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("requestid", testRequestId)
		return InternalServerErrorResponse(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), result["error"])
	assert.Equal(t, ServerErrorMessage, result["message"])
	assert.Equal(t, fmt.Sprintf(InternalServerError, testRequestId), result["details"])
}
