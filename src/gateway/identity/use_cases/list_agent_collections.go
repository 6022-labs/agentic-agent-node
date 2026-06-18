package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/responses"
)

type ListAgentCollections struct {
	agentCollectionsClient gateway_interfaces.AgentCollectionsClientInterface
}

func NewListAgentCollections(
	agentCollectionsClient gateway_interfaces.AgentCollectionsClientInterface,
) *ListAgentCollections {
	return &ListAgentCollections{agentCollectionsClient: agentCollectionsClient}
}

func (u *ListAgentCollections) Execute(ctx context.Context, req *requests.ListAgentCollectionsRequest) (*responses.AgentCollectionsResponse, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	payload, err := u.agentCollectionsClient.List(ctx, req.ChainId)
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
}

func (u *ListAgentCollections) toResponse(p *models.ExternalAgentCollectionsPayload) *responses.AgentCollectionsResponse {
	collections := make([]responses.AgentCollectionResponse, 0, len(p.Collections))
	for _, c := range p.Collections {
		collections = append(collections, responses.AgentCollectionResponse{
			Address:        c.Address,
			Name:           c.Name,
			Description:    c.Description,
			ModeratorCount: c.ModeratorCount,
			NextTokenId:    c.NextTokenId,
		})
	}
	return &responses.AgentCollectionsResponse{Collections: collections}
}
