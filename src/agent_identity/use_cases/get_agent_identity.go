package use_cases

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
)

type GetAgentIdentity struct {
	resolver common_interfaces.AgentIdentityResolverInterface
}

func NewGetAgentIdentity(resolver common_interfaces.AgentIdentityResolverInterface) *GetAgentIdentity {
	return &GetAgentIdentity{resolver: resolver}
}

func (u *GetAgentIdentity) Execute(ctx context.Context) *usecase_result.Result {
	identity, err := u.resolver.Resolve(ctx)
	if err != nil {
		if errors.Is(err, common_interfaces.ErrIdentityNotFound) {
			return usecase_result.NotFound("identity_not_found", err.Error())
		}
		return usecase_result.InternalError("identity_lookup_failed", err.Error())
	}
	return usecase_result.Ok(&responses.GetAgentIdentityResponse{
		CollectionAddress: identity.CollectionAddress,
		TokenId:           identity.TokenId,
		Name:              identity.Name,
	})
}
