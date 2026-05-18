package use_cases

import (
	"context"
	"errors"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
)

type PublishEnsRuntimeRecords struct {
	publisher identity_interfaces.EnsRuntimeRecordsPublisherInterface
}

func NewPublishEnsRuntimeRecords(publisher identity_interfaces.EnsRuntimeRecordsPublisherInterface) *PublishEnsRuntimeRecords {
	return &PublishEnsRuntimeRecords{publisher: publisher}
}

func (u *PublishEnsRuntimeRecords) Execute(ctx context.Context) *usecase_result.Result {
	if err := u.publisher.PublishRuntimeRecords(ctx); err != nil {
		if errors.Is(err, common_interfaces.ErrIdentityNotFound) {
			return usecase_result.Conflict("identity_not_found", "agent NFT must be minted before ENS records can be published")
		}
		return usecase_result.InternalError("ens_publish_failed", err.Error())
	}
	return usecase_result.NoContent()
}
