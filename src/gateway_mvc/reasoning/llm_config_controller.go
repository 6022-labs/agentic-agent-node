package reasoning

import (
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const (
	LlmConfigEndpoint         = "/config/llm"
	LlmConfigValidateEndpoint = "/config/llm/validate"
)

type LlmConfigController struct {
	getLlmConfig      *use_cases.GetLlmConfig
	setLlmConfig      *use_cases.SetLlmConfig
	deleteLlmConfig   *use_cases.DeleteLlmConfig
	validateLlmConfig *use_cases.ValidateLlmConfig
	errorWriter       *mvc.ErrorWriter
}

func NewLlmConfigController(
	getLlmConfig *use_cases.GetLlmConfig,
	setLlmConfig *use_cases.SetLlmConfig,
	deleteLlmConfig *use_cases.DeleteLlmConfig,
	validateLlmConfig *use_cases.ValidateLlmConfig,
	errorWriter *mvc.ErrorWriter,
) *LlmConfigController {
	return &LlmConfigController{
		getLlmConfig:      getLlmConfig,
		setLlmConfig:      setLlmConfig,
		deleteLlmConfig:   deleteLlmConfig,
		validateLlmConfig: validateLlmConfig,
		errorWriter:       errorWriter,
	}
}

func (c *LlmConfigController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *LlmConfigController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(LlmConfigEndpoint, c.Get)
	app.Post(LlmConfigEndpoint, c.Set)
	app.Delete(LlmConfigEndpoint, c.Delete)
	app.Post(LlmConfigValidateEndpoint, c.Validate)
}

// Validate checks the LLM provider configuration without saving it.
// @Summary Validate LLM config
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.SetLlmConfigRequest true "LLM config"
// @Success 204 ""
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/llm/validate [post]
func (c *LlmConfigController) Validate(ctx fiber.Ctx) error {
	body := new(requests.SetLlmConfigRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return c.errorWriter.Write(ctx, apperrors.NewBadRequestError("invalid_body", err.Error()))
	}
	if err := c.validateLlmConfig.Execute(ctx.Context(), body); err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// Get returns the LLM provider configuration.
// @Summary Get LLM config
// @Tags Config
// @Produce json
// @Success 200 {object} responses.GetLlmConfigResponse
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 404 {object} errors.NotFoundError "not_configured"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/llm [get]
func (c *LlmConfigController) Get(ctx fiber.Ctx) error {
	resp, err := c.getLlmConfig.Execute(ctx.Context())
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.JSON(resp)
}

// Set upserts the LLM provider configuration.
// @Summary Set LLM config
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.SetLlmConfigRequest true "LLM config"
// @Success 200 {object} responses.GetLlmConfigResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/llm [post]
func (c *LlmConfigController) Set(ctx fiber.Ctx) error {
	body := new(requests.SetLlmConfigRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return c.errorWriter.Write(ctx, apperrors.NewBadRequestError("invalid_body", err.Error()))
	}
	resp, err := c.setLlmConfig.Execute(ctx.Context(), body)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete clears the LLM provider configuration.
// @Summary Delete LLM config
// @Tags Config
// @Produce json
// @Success 204 ""
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/llm [delete]
func (c *LlmConfigController) Delete(ctx fiber.Ctx) error {
	if err := c.deleteLlmConfig.Execute(ctx.Context()); err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
