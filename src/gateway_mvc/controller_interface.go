package gateway_mvc

import "github.com/gofiber/fiber/v2"

// ControllerInterface separates unauthenticated routes (no JWT) from
// authenticated ones (JWT-gated).
//
// Unauthenticated: /health, /auth/*.
// Authenticated:   /responses, /config/*.
type ControllerInterface interface {
	RegisterUnauthenticatedRoutes(app fiber.Router)
	RegisterAuthenticatedRoutes(app fiber.Router)
}
