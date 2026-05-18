package runtime_mvc

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc_result"
	"github.com/gofiber/fiber/v2"
)

const InternalConversationMessagesEndpoint = "/internal/conversations/:conversationId/messages"

type ConversationMessagesController struct {
	appender use_cases.AppendConversationMessagesInterface
	lister   use_cases.ListConversationMessagesInterface
}

func NewConversationMessagesController(
	appender use_cases.AppendConversationMessagesInterface,
	lister use_cases.ListConversationMessagesInterface,
) *ConversationMessagesController {
	return &ConversationMessagesController{appender: appender, lister: lister}
}

func (c *ConversationMessagesController) RegisterRoutes(app fiber.Router) {
	app.Post(InternalConversationMessagesEndpoint, c.Append)
	app.Get(InternalConversationMessagesEndpoint, c.List)
}

// Append adds messages to a conversation transcript.
// @Summary Internal append messages
// @Description Idempotent batch append to the local conversation transcript. Dedup by SHA-512 hash of (role, content, conversationId).
// @Tags Internal
// @Accept json
// @Produce json
// @Param conversationId path string true "Conversation id"
// @Param request body requests.AppendConversationMessagesRequest true "Messages"
// @Success 200 {object} responses.AppendConversationMessagesResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 500 {object} fiber.Map "append_failed"
// @Router /internal/conversations/{conversationId}/messages [post]
func (c *ConversationMessagesController) Append(ctx *fiber.Ctx) error {
	conversationId := strings.TrimSpace(ctx.Params("conversationId"))
	if len(conversationId) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.NewValidationError("conversationId", "cannot be empty"))
	}

	body := new(requests.AppendConversationMessagesRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}

	return mvc_result.Write(ctx, c.appender.Execute(conversationId, body))
}

// List returns the conversation transcript ordered by timestamp ascending.
// @Summary Internal list messages
// @Tags Internal
// @Produce json
// @Param conversationId path string true "Conversation id"
// @Success 200 {object} responses.ListConversationMessagesResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 500 {object} fiber.Map "list_failed"
// @Router /internal/conversations/{conversationId}/messages [get]
func (c *ConversationMessagesController) List(ctx *fiber.Ctx) error {
	conversationId := strings.TrimSpace(ctx.Params("conversationId"))
	if len(conversationId) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.NewValidationError("conversationId", "cannot be empty"))
	}

	return mvc_result.Write(ctx, c.lister.Execute(conversationId))
}
