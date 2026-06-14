package middlewares

import (
	"strings"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/gofiber/fiber/v2"
)

const AuthAddressLocalKey = "auth.address"

type JwtAuthMiddleware struct {
	validator gateway_interfaces.JwtValidatorInterface
}

func NewJwtAuthMiddleware(validator gateway_interfaces.JwtValidatorInterface) *JwtAuthMiddleware {
	return &JwtAuthMiddleware{validator: validator}
}

func (m *JwtAuthMiddleware) Handle(ctx *fiber.Ctx) error {
	header := ctx.Get(fiber.HeaderAuthorization)
	if len(header) == 0 {
		return unauthorized(ctx, "missing_authorization", "Authorization header is required")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return unauthorized(ctx, "invalid_authorization", "Authorization must use Bearer scheme")
	}
	token := strings.TrimSpace(header[len(prefix):])
	if len(token) == 0 {
		return unauthorized(ctx, "missing_token", "bearer token is empty")
	}

	address, err := m.validator.Verify(token)
	if err != nil {
		return unauthorized(ctx, "invalid_token", err.Error())
	}

	ctx.Locals(AuthAddressLocalKey, address.Hex())
	return ctx.Next()
}

func unauthorized(ctx *fiber.Ctx, code string, message string) error {
	return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"code":    code,
		"message": message,
	})
}
