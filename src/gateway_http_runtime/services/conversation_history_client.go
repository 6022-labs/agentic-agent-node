package services

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpConversationHistoryClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpConversationHistoryClient(client clients.ClientWithResponsesInterface) *HttpConversationHistoryClient {
	return &HttpConversationHistoryClient{client: client}
}

func (c *HttpConversationHistoryClient) ListConversations(ctx context.Context) (*models.ExternalListConversationsPayload, error) {
	resp, err := c.client.GetInternalConversationsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	var payload models.ExternalListConversationsPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpConversationHistoryClient) ListMessages(ctx context.Context, conversationId string) (*models.ExternalListConversationMessagesPayload, error) {
	resp, err := c.client.GetInternalConversationsConversationIdMessagesWithResponse(ctx, conversationId)
	if err != nil {
		return nil, err
	}
	var payload models.ExternalListConversationMessagesPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}
