package interfaces

import "github.com/gofiber/fiber/v3"

// AuthAwareControllerInterface splits unauthenticated routes (/health, /auth/*) from
// JWT-gated ones (/responses, /config/*).
type AuthAwareControllerInterface interface {
	RegisterUnauthenticatedRoutes(app fiber.Router)
	RegisterAuthenticatedRoutes(app fiber.Router)
}
