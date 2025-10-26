package response

import "github.com/gofiber/fiber/v2"

type HttpResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Err     interface{} `json:"error,omitempty"`
}

// SendResponse is a helper to send JSON responses in Fiber
func SendResponse(c *fiber.Ctx, statusCode int, data interface{}, message string) error {
	response := HttpResponse{
		Status:  statusCode,
		Message: message,
		Data:    data,
	}
	return c.Status(statusCode).JSON(response)
}

// SendResponseWithError is a helper to send JSON responses in Fiber, with additional error field
func SendResponseWithError(c *fiber.Ctx, statusCode int, data interface{}, message string, err interface{}) error {
	response := HttpResponse{
		Status:  statusCode,
		Message: message,
		Data:    data,
		Err:     err,
	}
	return c.Status(statusCode).JSON(response)
}

// SUCCESS
func Success(c *fiber.Ctx, message string, data interface{}) error {
	return SendResponse(c, fiber.StatusOK, data, message)
}

func Created(c *fiber.Ctx, message string, data interface{}) error {
	return SendResponse(c, fiber.StatusCreated, data, message)
}

// ERROR
func BadRequest(c *fiber.Ctx, message string, data interface{}) error {
	return SendResponse(c, fiber.StatusBadRequest, data, message)
}

func Unauthorized(c *fiber.Ctx, message string, data interface{}) error {
	return SendResponse(c, fiber.StatusUnauthorized, data, message)
}

func Forbidden(c *fiber.Ctx, message string, data interface{}) error {
	return SendResponse(c, fiber.StatusForbidden, data, message)
}

func NotFound(c *fiber.Ctx, message string, data interface{}) error {
	return SendResponse(c, fiber.StatusNotFound, data, message)
}

func InternalServerError(c *fiber.Ctx, message string, data interface{}) error {
	return SendResponse(c, fiber.StatusInternalServerError, data, message)
}
