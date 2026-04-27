package runtime_mvc

import "github.com/gofiber/fiber/v2"

// ControllerInterface registers internal-only routes mounted under /internal.
// Runtime has no auth split — gateway is the only caller and reaches it via
// the loopback interface.
type ControllerInterface interface {
	RegisterRoutes(app fiber.Router)
}
