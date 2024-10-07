package presenter

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type errorMessage struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details"`
}

const ServerErrorMessage = "An error occurred on the server."
const InternalServerError = "An internal server error occurred. Please contact the administrator and provide the following request ID: %s."

func ErrorResponse(c *fiber.Ctx, httpStatus int, message string, details interface{}) error {
	return c.Status(httpStatus).JSON(errorMessage{
		Error:   http.StatusText(httpStatus),
		Message: message,
		Details: details,
	})
}

func InternalServerErrorResponse(c *fiber.Ctx) error {
	requestId := c.Locals("requestid").(string)
	return ErrorResponse(c, http.StatusInternalServerError, ServerErrorMessage, fmt.Sprintf(InternalServerError, requestId))
}

func UnprocessableEntityResponse(c *fiber.Ctx, message string, details interface{}) error {
	return ErrorResponse(c, http.StatusUnprocessableEntity, message, details)
}

func ConflictResponse(c *fiber.Ctx, message string, details string) error {
	return ErrorResponse(c, http.StatusConflict, message, details)
}

func BadRequestResponse(c *fiber.Ctx, message string, details string) error {
	return ErrorResponse(c, http.StatusBadRequest, message, details)
}

func NotFoundResponse(c *fiber.Ctx, message string, details string) error {
	return ErrorResponse(c, http.StatusNotFound, message, details)
}

func ForbiddenResponse(c *fiber.Ctx, message string, details string) error {
	return ErrorResponse(c, http.StatusForbidden, message, details)
}

func UnauthorizedResponse(c *fiber.Ctx, message string, details string) error {
	return ErrorResponse(c, http.StatusUnauthorized, message, details)
}
