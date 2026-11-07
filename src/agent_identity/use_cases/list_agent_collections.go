package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// ListAgentCollections returns the collections the manager governs so the
// owner can pick one when drafting the registration.
type ListAgentCollections struct {
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface
}

func NewListAgentCollections(
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface,
) *ListAgentCollections {
	return &ListAgentCollections{agentCollectionsManagerRequester: agentCollectionsManagerRequester}
}

func (u *ListAgentCollections) Execute(
	ctx context.Context,
	req *requests.ListAgentCollectionsRequest,
) (*responses.AgentCollectionsResponse, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	collections, err := u.agentCollectionsManagerRequester.ListCollections(req.ChainId)
	if err != nil {
		return nil, errors.NewUnavailableError("collections_read_failed", "failed to list the collections", err)
	}

	items := make([]responses.AgentCollectionResponse, 0, len(collections))
	for _, collection := range collections {
		items = append(items, responses.AgentCollectionResponse{
			Address:        collection.Address,
			Name:           collection.Name,
			Description:    collection.Description,
			ModeratorCount: collection.ModeratorCount,
			NextTokenId:    collection.NextTokenId,
		})
	}
	return &responses.AgentCollectionsResponse{Collections: items}, nil
}
