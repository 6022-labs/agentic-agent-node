package gateway_mvc

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/gofiber/fiber/v2"
)

const (
	ConversationsEndpoint        = "/conversations"
	ConversationMessagesEndpoint = "/conversations/:conversationId/messages"
)

type ConversationsController struct {
	listConversations        *use_cases.ListConversations
	listConversationMessages *use_cases.ListConversationMessages
}

func NewConversationsController(
	listConversations *use_cases.ListConversations,
	listConversationMessages *use_cases.ListConversationMessages,
) *ConversationsController {
	return &ConversationsController{
		listConversations:        listConversations,
		listConversationMessages: listConversationMessages,
	}
}

func (c *ConversationsController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *ConversationsController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(ConversationsEndpoint, c.List)
	app.Get(ConversationMessagesEndpoint, c.ListMessages)
}

// List returns the agent's past conversations (most recent first).
// @Summary List conversations
// @Tags Conversations
// @Produce json
// @Success 200 {object} responses.ListConversationsResponse
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /conversations [get]
func (c *ConversationsController) List(ctx *fiber.Ctx) error {
	resp, err := c.listConversations.Execute(ctx.UserContext())
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(resp)
}

// ListMessages returns the transcript for one conversation (timestamp ascending).
// @Summary List conversation messages
// @Tags Conversations
// @Produce json
// @Param conversationId path string true "Conversation id"
// @Success 200 {object} responses.ListConversationMessagesResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /conversations/{conversationId}/messages [get]
func (c *ConversationsController) ListMessages(ctx *fiber.Ctx) error {
	conversationId := strings.TrimSpace(ctx.Params("conversationId"))
	if len(conversationId) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.NewValidationError("conversationId", "cannot be empty"))
	}

	resp, err := c.listConversationMessages.Execute(ctx.UserContext(), conversationId)
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(resp)
}
