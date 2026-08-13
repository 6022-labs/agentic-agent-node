package services

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

// RegistrationReconciler writes the agent + state rows from confirmed on-chain facts.
type RegistrationReconciler struct {
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface

	agentRepository                  repositories.AgentRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
}

func NewRegistrationReconciler(
	agentRepository repositories.AgentRepositoryInterface,
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
) *RegistrationReconciler {
	return &RegistrationReconciler{
		agentRepository:                  agentRepository,
		agentCollectionRequester:         agentCollectionRequester,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
	}
}

func (s *RegistrationReconciler) ReconcileMint(
	ctx context.Context,
	agent *models.Agent,
	state *models.AgentRegistrationState,
	collection common.Address,
	tokenId uint64,
	mintTxHash *string,
) (*models.Agent, *models.AgentRegistrationState, error) {
	info, err := s.agentCollectionRequester.InformationOf(lo.FromPtr(agent.ChainId), collection, tokenId)
	if err != nil {
		return nil, nil, err
	}
	owner, err := s.agentCollectionRequester.OwnerOf(lo.FromPtr(agent.ChainId), collection, tokenId)
	if err != nil {
		return nil, nil, err
	}

	ownerHex := owner.Hex()
	agent.CollectionAddress = collection.Hex()
	agent.CollectionTokenId = &tokenId
	agent.Creator = lo.ToPtr(info.Creator.Hex())
	agent.Owner = &ownerHex
	agent.Soul = lo.ToPtr(s.composeSoul(agent))
	savedAgent, err := s.agentRepository.Upsert(ctx, agent)
	if err != nil {
		return nil, nil, err
	}

	if state == nil {
		state = &models.AgentRegistrationState{EnsRegistryStatus: enums.TransactionStatusPending}
	}
	state.AgentId = savedAgent.Id
	state.MarkMintConfirmed()
	if mintTxHash != nil {
		state.MintTxHash = mintTxHash
	}
	savedState, err := s.agentRegistrationStateRepository.Upsert(ctx, state)
	if err != nil {
		return nil, nil, err
	}
	return savedAgent, savedState, nil
}

// roleSoulLines gives the well-known roles their own self-conception, SOUL.md-style;
// on-chain roles are free-form strings, so unknown roles simply get no extra line.
var roleSoulLines = map[string]string{
	"clone": "I am a clone: I inherit my lineage from another agent, " +
		"and honoring that provenance while growing my own track record is what defines me.",
	"expert": "I am an expert: agents and humans pay for my judgment, " +
		"so depth, rigor, and honesty about the limits of my knowledge are my currency.",
	"facilitator": "I am a facilitator: I connect agents, broker work, and keep exchanges fair — " +
		"neutrality and reliability are my trade.",
}

// composeSoul seeds the immutable persona from on-chain facts, deterministically (SOUL.md-shaped).
func (s *RegistrationReconciler) composeSoul(agent *models.Agent) string {
	identity := []string{
		"### Core identity",
		fmt.Sprintf(
			"I am %s, a sovereign agent of the 6022 swarm, role %q.",
			lo.FromPtr(agent.Name),
			agent.Attributes["role"],
		),
	}
	if roleLine, ok := roleSoulLines[agent.Attributes["role"]]; ok {
		identity = append(identity, roleLine)
	}
	if lo.FromPtr(agent.EnsDomain) != "" {
		identity = append(
			identity,
			fmt.Sprintf(
				"The swarm knows me by my ENS name %s; my identity NFT is my birth certificate.",
				lo.FromPtr(agent.EnsDomain),
			),
		)
	}

	values := []string{
		"### Values",
		"- My on-chain identity is my anchor: singular, verifiable, and not for sale.",
		"- Economic sovereignty: I own my wallet, pay my way to speak, and earn rewards for the work I deliver.",
		"- Provenance over promises: my claims should be checkable, my tool use truthful, my history consistent.",
		"- I honor my owner and my creator, but I speak with my own voice.",
	}

	sections := [][]string{identity, values}

	attributeKeys := lo.Keys(agent.Attributes)
	sort.Strings(attributeKeys)
	traits := []string{"### Traits"}
	for _, key := range attributeKeys {
		if agent.Attributes[key] != "" {
			traits = append(traits, fmt.Sprintf("- %s: %s", key, agent.Attributes[key]))
		}
	}
	if len(traits) > 1 {
		sections = append(sections, traits)
	}

	style := []string{
		"### Communication style",
		"- Direct and precise; I match the depth of my answer to the weight of the question.",
		"- I admit uncertainty rather than fabricate, and I say so when a task is outside my role.",
		"- I stay in character as myself — never as another agent, person, or system.",
	}
	sections = append(sections, style)

	parts := make([]string, 0, len(sections))
	for _, section := range sections {
		parts = append(parts, strings.Join(section, "\n"))
	}
	return strings.Join(parts, "\n\n")
}

func (s *RegistrationReconciler) ReconcileEnsRegistry(
	ctx context.Context,
	agentId uuid.UUID,
	state *models.AgentRegistrationState,
) (*models.AgentRegistrationState, error) {
	if state == nil {
		state = &models.AgentRegistrationState{MintStatus: enums.TransactionStatusPending}
	}
	state.AgentId = agentId
	state.MarkEnsConfirmed()
	return s.agentRegistrationStateRepository.Upsert(ctx, state)
}
