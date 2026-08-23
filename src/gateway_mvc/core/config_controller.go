package core

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	"github.com/gofiber/fiber/v3"
)

const ConfigEndpoint = "/config"

// ConfigController serves the public chain config the owner UI boots from —
// unauthenticated because the UI needs the chain id before wallet login.
type ConfigController struct {
	getBlockchainConfig *use_cases.GetBlockchainConfig
}

func NewConfigController(
	getBlockchainConfig *use_cases.GetBlockchainConfig,
) *ConfigController {
	return &ConfigController{getBlockchainConfig: getBlockchainConfig}
}

func (c *ConfigController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Get(ConfigEndpoint, c.Get)
}

func (c *ConfigController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

// Get returns the chain id and contract addresses the node runs against.
// @Summary Public chain config
// @Tags Config
// @Produce json
// @Success 200 {object} responses.BlockchainConfigResponse
// @Router /config [get]
func (c *ConfigController) Get(ctx fiber.Ctx) error {
	resp, err := c.getBlockchainConfig.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
