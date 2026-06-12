package agent_knowledge_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/gofiber/fiber/v2"
)

const (
	InternalConversationMessagesEndpoint = "/internal/conversations/:conversationId/messages"
	InternalConversationsEndpoint        = "/internal/conversations"
)

type ConversationMessagesController struct {
	appendConversationMessages *use_cases.AppendConversationMessages
	listConversationMessages   *use_cases.ListConversationMessages
	listConversations          *use_cases.ListConversations
}

func NewConversationMessagesController(
	appendConversationMessages *use_cases.AppendConversationMessages,
	listConversationMessages *use_cases.ListConversationMessages,
	listConversations *use_cases.ListConversations,
) *ConversationMessagesController {
	return &ConversationMessagesController{
		appendConversationMessages: appendConversationMessages,
		listConversationMessages:   listConversationMessages,
		listConversations:          listConversations,
	}
}

func (c *ConversationMessagesController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalConversationsEndpoint, c.ListConversations)
	app.Post(InternalConversationMessagesEndpoint, c.Append)
	app.Get(InternalConversationMessagesEndpoint, c.List)
}

// ListConversations returns one summary row per distinct conversation id.
// @Summary Internal list conversations
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.ListConversationsResponse
// @Failure 500 {object} fiber.Map "list_failed"
// @Router /internal/conversations [get]
func (c *ConversationMessagesController) ListConversations(ctx *fiber.Ctx) error {
	resp, err := c.listConversations.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
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
func (c *ConversationMessagesController) List(ctx *fiber.Ctx) error {
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
