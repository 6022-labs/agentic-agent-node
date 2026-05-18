package use_cases

import (
	"context"

	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
)

type GetBootstrapStatus struct {
	stateReader      common_interfaces.BootstrapStateReaderInterface
	identityResolver common_interfaces.AgentIdentityResolverInterface
}

func NewGetBootstrapStatus(
	stateReader common_interfaces.BootstrapStateReaderInterface,
	identityResolver common_interfaces.AgentIdentityResolverInterface,
) *GetBootstrapStatus {
	return &GetBootstrapStatus{
		stateReader:      stateReader,
		identityResolver: identityResolver,
	}
}

func (u *GetBootstrapStatus) Execute(ctx context.Context) *usecase_result.Result {
	state := u.stateReader.Get()

	resp := &responses.GetBootstrapStatusResponse{
		Status:       state.Status,
		AgentAddress: state.AgentAddress,
		LastError:    state.LastError,
	}

	if state.Status == enums.BootstrapStatusReady {
		identity, err := u.identityResolver.Resolve(ctx)
		if err != nil {
			resp.Identity = &responses.AgentIdentitySnapshot{
				Resolved: false,
				Error:    err.Error(),
			}
		} else {
			resp.Identity = &responses.AgentIdentitySnapshot{
				Resolved:          true,
				CollectionAddress: identity.CollectionAddress,
				TokenId:           identity.TokenId,
				Name:              identity.Name,
			}
		}
	}

	return usecase_result.Ok(resp)
}
