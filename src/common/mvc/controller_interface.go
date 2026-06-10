package mvc

import "github.com/gofiber/fiber/v2"

// ControllerInterface registers internal-only routes mounted under /internal.
// Runtime has no auth split — gateway is the only caller and reaches it via
// the loopback interface. The runtime mvc adapters implement it and register into the shared
// "runtime_controllers" dig group that runtime_host collects.
type ControllerInterface interface {
	RegisterRoutes(app fiber.Router)
}
