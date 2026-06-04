package use_cases

import (
	"context"
	"errors"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
)

type PublishEnsRuntimeRecords struct {
	publisher identity_interfaces.EnsRuntimeRecordsPublisherInterface
}

func NewPublishEnsRuntimeRecords(publisher identity_interfaces.EnsRuntimeRecordsPublisherInterface) *PublishEnsRuntimeRecords {
	return &PublishEnsRuntimeRecords{publisher: publisher}
}

func (u *PublishEnsRuntimeRecords) Execute(ctx context.Context) error {
	if err := u.publisher.PublishRuntimeRecords(ctx); err != nil {
		if errors.Is(err, common_interfaces.ErrIdentityNotFound) {
			return apperrors.NewConflictError("identity_not_found", "agent NFT must be minted before ENS records can be published")
		}
		return apperrors.NewInternalError("ens_publish_failed", err.Error())
	}
	return nil
}
