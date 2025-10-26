package entity

import (
	"github.com/mystaline/clefinport-be/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type HttpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     any    `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func (e *HttpError) Error() string {
	return e.Message
}

func (e *HttpError) SendResponse(ctx *fiber.Ctx) error {
	return response.SendResponse(ctx, e.Code, nil, e.Message)
}

func (e *HttpError) SendResponseWithError(ctx *fiber.Ctx) error {
	return response.SendResponseWithError(ctx, e.Code, nil, e.Message, e.Err)
}

func InternalServerError(message string) *HttpError {
	return &HttpError{
		Code:    fiber.StatusInternalServerError,
		Message: message,
	}
}

func BadRequest(message string) *HttpError {
	return &HttpError{
		Code:    fiber.StatusBadRequest,
		Message: message,
	}
}

func Unauthorized(message string) *HttpError {
	return &HttpError{
		Code:    fiber.StatusUnauthorized,
		Message: message,
	}
}

func Forbidden(message string) *HttpError {
	return &HttpError{
		Code:    fiber.StatusForbidden,
		Message: message,
	}
}

func NotFound(message string) *HttpError {
	return &HttpError{
		Code:    fiber.StatusNotFound,
		Message: message,
	}
}

func Conflict(message string) *HttpError {
	return &HttpError{
		Code:    fiber.StatusConflict,
		Message: message,
	}
}

func ToHttpError(err error) *HttpError {
	if httpErr, ok := err.(*HttpError); ok {
		return httpErr
	}

	return InternalServerError(err.Error())
}

// ─────── Specific error messages ───────

// Auth
