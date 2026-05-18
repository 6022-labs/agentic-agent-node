// Package mvc_result translates a usecase_result.Result into a fiber HTTP
// response. It is the only place the outcome -> HTTP status mapping lives.
package mvc_result

import (
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	"github.com/gofiber/fiber/v2"
)

// Write serialises a use-case Result onto the fiber context. Returns the
// fiber error so callers can `return mvc_result.Write(ctx, r)` directly.
func Write(ctx *fiber.Ctx, r *usecase_result.Result) error {
	if r == nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "nil_result",
			"message": "use case returned a nil result",
		})
	}

	status := toHttpStatus(r.Outcome)

	if r.Body != nil {
		return ctx.Status(status).JSON(r.Body)
	}
	if r.Code != "" {
		return ctx.Status(status).JSON(fiber.Map{
			"code":    r.Code,
			"message": r.Message,
		})
	}
	return ctx.SendStatus(status)
}

func toHttpStatus(o usecase_result.Outcome) int {
	switch o {
	case usecase_result.OutcomeOk:
		return fiber.StatusOK
	case usecase_result.OutcomeCreated:
		return fiber.StatusCreated
	case usecase_result.OutcomeNoContent:
		return fiber.StatusNoContent
	case usecase_result.OutcomeInvalidInput:
		return fiber.StatusBadRequest
	case usecase_result.OutcomeUnauthorized:
		return fiber.StatusUnauthorized
	case usecase_result.OutcomeForbidden:
		return fiber.StatusForbidden
	case usecase_result.OutcomeNotFound:
		return fiber.StatusNotFound
	case usecase_result.OutcomeConflict:
		return fiber.StatusConflict
	case usecase_result.OutcomeUnavailable:
		return fiber.StatusBadGateway
	default:
		return fiber.StatusInternalServerError
	}
}
