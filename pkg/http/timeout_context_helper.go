package http

import (
	"context"
	"time"

	"github.com/mystaline/clefinport-be/pkg/entity"
	"github.com/mystaline/clefinport-be/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type UseCaseFunc[T any] func(ctx context.Context) (T, *entity.HttpError)

// RunWithTimeout runs a use case with timeout and handles Fiber response to avoid rewrite in every usecase calls in controller causing controller bloat
// Also properly pass context with timeout and handle early timeout before finish
// Below are more detailed descriptions with example usage for onboarding purpose to this helper
//
// Parameters:
//   - ctx: *fiber.Ctx – the current Fiber context
//   - timeout: time.Duration – the maximum time to wait for the use case before timing out
//   - useCase: UseCaseFunc[T] – the function to run with timeout enforcement
//   - successMessage: string – the success message to include in the response if successful
//
// Example:
//
//	func (c *MyController) GetData(ctx *fiber.Ctx) error {
//	   someCompanyCode := ctx.Locals(some_util.someKey).(string)
//
//	   return RunWithTimeout(ctx, 5*time.Second, func(ctx context.Context) (*[]MyResponseDTO, *HttpError) {
//	       param := MyUseCaseParam{Ctx: ctx, ID: id}
//	       res, err := myUseCase.Invoke(param)
//
//	       if err != nil {
//	           return nil, some_util.ToHttpError(err)
//	       }
//
//	       return res, nil
//	   }, "Successfully fetched data")
//	}
func RunWithTimeout[T any](
	ctx *fiber.Ctx,
	timeout time.Duration,
	useCase UseCaseFunc[T],
	successMessage string,
	successCode int,
) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.UserContext(), timeout)
	defer cancel()

	resultChan := make(chan T)
	errorChan := make(chan entity.HttpError)

	go func() {
		res, err := useCase(ctxWithTimeout)
		if err != nil {
			select {
			case errorChan <- *entity.ToHttpError(err):
			case <-ctxWithTimeout.Done():
			}
			return
		}
		select {
		case <-ctxWithTimeout.Done():
		case resultChan <- res:
		}
	}()

	select {
	case <-ctxWithTimeout.Done():
		return response.SendResponse(ctx, fiber.StatusRequestTimeout, nil, "Timeout")
	case err := <-errorChan:
		return response.SendResponse(ctx, err.Code, err.Data, err.Message)
	case res := <-resultChan:
		return response.SendResponse(ctx, successCode, res, successMessage)
	}
}
