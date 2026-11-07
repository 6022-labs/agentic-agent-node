package middlewares

import (
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/gofiber/fiber/v3"
)

type JwtAuthMiddleware struct {
	validator   gateway_interfaces.JwtValidatorInterface
	errorWriter *mvc.ErrorWriter
}

func NewJwtAuthMiddleware(
	validator gateway_interfaces.JwtValidatorInterface,
	errorWriter *mvc.ErrorWriter,
) *JwtAuthMiddleware {
	return &JwtAuthMiddleware{validator: validator, errorWriter: errorWriter}
}

func (m *JwtAuthMiddleware) Handle(ctx fiber.Ctx) error {
	header := ctx.Get(fiber.HeaderAuthorization)
	if len(header) == 0 {
		return m.errorWriter.Write(
			ctx,
			apperrors.NewUnauthorizedError("missing_authorization", "Authorization header is required"),
		)
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return m.errorWriter.Write(
			ctx,
			apperrors.NewUnauthorizedError("invalid_authorization", "Authorization must use Bearer scheme"),
		)
	}
	token := strings.TrimSpace(header[len(prefix):])
	if len(token) == 0 {
		return m.errorWriter.Write(ctx, apperrors.NewUnauthorizedError("missing_token", "bearer token is empty"))
	}

	address, err := m.validator.Verify(token)
	if err != nil {
		return m.errorWriter.Write(ctx, apperrors.NewUnauthorizedError("invalid_token", err.Error()))
	}

	ctx.SetContext(helpers.ContextWithCallerId(ctx.Context(), address.Hex()))
	return ctx.Next()
}
