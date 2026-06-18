package mvc

import "github.com/gofiber/fiber/v2"

// ControllerInterface splits unauthenticated routes (/health, /auth/*) from
// JWT-gated ones (/responses, /config/*).
type AuthAwareControllerInterface interface {
	RegisterUnauthenticatedRoutes(app fiber.Router)
	RegisterAuthenticatedRoutes(app fiber.Router)
}
