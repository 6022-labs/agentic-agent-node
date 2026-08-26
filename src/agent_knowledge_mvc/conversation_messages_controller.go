package agent_knowledge_mvc

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/mvc"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/gofiber/fiber/v3"
)

const InternalConversationMessagesEndpoint = "/internal/conversations/:conversationId/messages"

type ConversationMessagesController struct {
	appendConversationMessages *use_cases.AppendConversationMessages
	listConversationMessages   *use_cases.ListConversationMessages
}

func NewConversationMessagesController(
	appendConversationMessages *use_cases.AppendConversationMessages,
	listConversationMessages *use_cases.ListConversationMessages,
) *ConversationMessagesController {
	return &ConversationMessagesController{
		appendConversationMessages: appendConversationMessages,
		listConversationMessages:   listConversationMessages,
	}
}

func (c *ConversationMessagesController) RegisterRoutes(app fiber.Router) {
	app.Post(InternalConversationMessagesEndpoint, c.Append)
	app.Get(InternalConversationMessagesEndpoint, c.List)
}

// Append adds messages to a conversation transcript.
// @Summary Internal append messages
// @Description Idempotent batch append to the transcript. Dedup by SHA-512 hash of (role, content, conversationId).
// @Tags Internal
// @Accept json
// @Produce json
// @Param conversationId path string true "Conversation id"
// @Param request body requests.AppendConversationMessagesRequest true "Messages"
// @Success 200 {object} responses.AppendConversationMessagesResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 500 {object} fiber.Map "append_failed"
// @Router /internal/conversations/{conversationId}/messages [post]
func (c *ConversationMessagesController) Append(ctx fiber.Ctx) error {
	conversationId := strings.TrimSpace(ctx.Params("conversationId"))
	if len(conversationId) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.NewValidationError("conversationId", "cannot be empty"))
	}

	body := new(requests.AppendConversationMessagesRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}

	resp, err := c.appendConversationMessages.Execute(ctx.Context(), conversationId, body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
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
func (c *ConversationMessagesController) List(ctx fiber.Ctx) error {
	conversationId := strings.TrimSpace(ctx.Params("conversationId"))
	if len(conversationId) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.NewValidationError("conversationId", "cannot be empty"))
	}

	resp, err := c.listConversationMessages.Execute(ctx.Context(), conversationId)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
