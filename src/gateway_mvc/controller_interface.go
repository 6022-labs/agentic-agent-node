package gateway_mvc

import "github.com/gofiber/fiber/v2"

// ControllerInterface separates unauthenticated routes (no JWT, available
// pre-bootstrap) from authenticated ones (JWT-gated, only served once the
// runtime is `ready`).
//
// Unauthenticated: /health, /bootstrap/status, /auth/*.
// Authenticated:   /responses, /config/*.
type ControllerInterface interface {
	RegisterUnauthenticatedRoutes(app fiber.Router)
	RegisterAuthenticatedRoutes(app fiber.Router)
}
