package use_cases

import (
	"context"
	"fmt"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/common"
)

// SaveAgentRegistration upserts the agent draft the connected wallet filled in,
// binding it to the node's bootstrap wallet, and seeds the registration state.
type SaveAgentRegistration struct {
	agentRepository                  repositories.AgentRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface
	agentCollectionRequester         identity_interfaces.AgentCollectionRequesterInterface
	walletCredentialsStorage         identity_interfaces.WalletCredentialsStorageInterface
}

func NewSaveAgentRegistration(
	agentRepository repositories.AgentRepositoryInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface,
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface,
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface,
) *SaveAgentRegistration {
	return &SaveAgentRegistration{
		agentRepository:                  agentRepository,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
		agentCollectionsManagerRequester: agentCollectionsManagerRequester,
		agentCollectionRequester:         agentCollectionRequester,
		walletCredentialsStorage:         walletCredentialsStorage,
	}
}

func (u *SaveAgentRegistration) Execute(ctx context.Context, req *requests.SaveAgentRegistrationRequest) (*responses.AgentRegistrationResponse, error) {
	collection := common.HexToAddress(req.CollectionAddress)

	known, err := u.agentCollectionsManagerRequester.IsKnownCollection(collection)
	if err != nil {
		return nil, errors.NewUnavailableError("collection_check_failed", err.Error())
	}
	if !known {
		return nil, errors.NewValidationError("collectionAddress", "collection is not governed by the manager")
	}

	cred, found, err := u.walletCredentialsStorage.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("wallet_lookup_failed", err.Error())
	}
	if !found {
		return nil, errors.NewUnavailableError("wallet_not_ready", "node wallet has not been bootstrapped yet")
	}

	collectionName, err := u.agentCollectionRequester.CollectionName(collection)
	if err != nil {
		return nil, errors.NewUnavailableError("collection_check_failed", err.Error())
	}

	existing, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("agent_lookup_failed", err.Error())
	}

	owner := req.Owner
	var role enums.AgentRole
	_ = role.Set(req.Role)

	agent := &models.Agent{
		CollectionAddress: collection.Hex(),
		Name:              req.Name,
		CloneOf:           req.CloneOf,
		EnsDomain:         ensDomain(req.Name, collectionName),
		Role:              role,
		Creator:           owner,
		WalletAddress:     cred.Address,
		Owner:             &owner,
		Images:            req.Images,
		Attributes:        req.Attributes,
	}
	// Preserve an already-minted identity if the owner re-edits the form.
	if existing != nil {
		agent.CollectionTokenId = existing.CollectionTokenId
		agent.ProposalId = existing.ProposalId
	}

	saved, err := u.agentRepository.Upsert(ctx, agent)
	if err != nil {
		return nil, errors.NewInternalError("agent_save_failed", err.Error())
	}

	state, err := u.agentRegistrationStateRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("state_lookup_failed", err.Error())
	}
	if state == nil {
		state, err = u.agentRegistrationStateRepository.Upsert(ctx, &models.AgentRegistrationState{
			MintStatus:        enums.TransactionStatusPending,
			EnsRegistryStatus: enums.TransactionStatusPending,
		})
		if err != nil {
			return nil, errors.NewInternalError("state_save_failed", err.Error())
		}
	}

	resp := &responses.AgentRegistrationResponse{}
	resp.FromModel(saved, state)
	return resp, nil
}

// ensDomain builds the human-readable agent domain label path (agent.collection).
func ensDomain(name, collectionName string) string {
	return fmt.Sprintf("%s.%s", strings.ToLower(name), strings.ToLower(collectionName))
}
