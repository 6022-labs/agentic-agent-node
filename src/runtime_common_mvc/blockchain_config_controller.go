package runtime_common_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/use_cases"
	"github.com/gofiber/fiber/v3"
)

const InternalBlockchainConfigEndpoint = "/internal/config/blockchain"

// BlockchainConfigController exposes the node's static chain config; the
// gateway serves it unauthenticated so the owner UI boots from the API.
type BlockchainConfigController struct {
	getBlockchainConfig *use_cases.GetBlockchainConfig
}

func NewBlockchainConfigController(
	getBlockchainConfig *use_cases.GetBlockchainConfig,
) *BlockchainConfigController {
	return &BlockchainConfigController{getBlockchainConfig: getBlockchainConfig}
}

func (c *BlockchainConfigController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalBlockchainConfigEndpoint, c.Get)
}

// Get returns the chain id and contract addresses bound at startup.
// @Summary Get blockchain config
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.BlockchainConfigResponse
// @Router /internal/config/blockchain [get]
func (c *BlockchainConfigController) Get(ctx fiber.Ctx) error {
	resp, err := c.getBlockchainConfig.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
