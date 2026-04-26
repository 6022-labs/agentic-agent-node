package gateway_mvc

import "github.com/gofiber/fiber/v2"

func notImplementedBody(useCase string) fiber.Map {
	return fiber.Map{
		"code":    "not_implemented",
		"message": "use case '" + useCase + "' is not implemented yet",
	}
}
