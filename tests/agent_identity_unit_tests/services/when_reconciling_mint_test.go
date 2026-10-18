package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReconcilingMintTestingSuite struct {
	sut *services.RegistrationReconciler

	mockAgentCollectionRequester         *interface_mocks.MockAgentCollectionRequesterInterface
	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
}

func WhenReconcilingMintBeforeEach(t *testing.T) *WhenReconcilingMintTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentCollectionRequester := interface_mocks.NewMockAgentCollectionRequesterInterface(mockController)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(
		mockController,
	)

	sut := services.NewRegistrationReconciler(
		mockAgentRepository,
		mockAgentCollectionRequester,
		mockAgentRegistrationStateRepository,
	)

	return &WhenReconcilingMintTestingSuite{
		sut:                                  sut,
		mockAgentCollectionRequester:         mockAgentCollectionRequester,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
	}
}

func TestWhenReconcilingMint(t *testing.T) {
	t.Parallel()

	t.Run("Given a freshly minted token", func(t *testing.T) {
		t.Parallel()

		creator := common_tests.RandomEthAddress()
		nftOwner := common_tests.RandomEthAddress()

		initSuite := func(suite *WhenReconcilingMintTestingSuite) {
			suite.mockAgentCollectionRequester.EXPECT().
				InformationOf(gomock.Any(), gomock.Any(), uint64(5)).
				Return(identity_interfaces.OnChainAgentInfo{Creator: creator}, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				OwnerOf(gomock.Any(), gomock.Any(), uint64(5)).
				Return(nftOwner, nil)
			suite.mockAgentRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, a *models.Agent) (*models.Agent, error) { return a, nil })
			suite.mockAgentRegistrationStateRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, s *models.AgentRegistrationState) (*models.AgentRegistrationState, error) {
					return s, nil
				})
		}

		t.Run("Should write the tokenId + confirmed mint status onto the rows", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingMintBeforeEach(t)
			initSuite(suite)

			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			savedAgent, savedState, err := suite.sut.ReconcileMint(
				context.Background(),
				agent,
				nil,
				common_tests.RandomEthAddress(),
				5,
				nil,
			)

			assert.NoError(t, err)
			assert.NotNil(t, savedAgent.CollectionTokenId)
			assert.Equal(t, uint64(5), *savedAgent.CollectionTokenId)
			assert.Equal(t, creator.Hex(), lo.FromPtr(savedAgent.Creator))
			assert.Equal(t, identity_enums.TransactionStatusConfirmed, savedState.MintStatus)
			assert.NotNil(t, savedState.MintedAt)
		})
	})

	initSoulSuite := func(suite *WhenReconcilingMintTestingSuite) {
		suite.mockAgentCollectionRequester.EXPECT().
			InformationOf(gomock.Any(), gomock.Any(), uint64(5)).
			Return(identity_interfaces.OnChainAgentInfo{Creator: common_tests.RandomEthAddress()}, nil)
		suite.mockAgentCollectionRequester.EXPECT().
			OwnerOf(gomock.Any(), gomock.Any(), uint64(5)).
			Return(common_tests.RandomEthAddress(), nil)
		suite.mockAgentRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, a *models.Agent) (*models.Agent, error) { return a, nil })
		suite.mockAgentRegistrationStateRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, s *models.AgentRegistrationState) (*models.AgentRegistrationState, error) {
				return s, nil
			})
	}

	t.Run("Given no soul exists", func(t *testing.T) {
		t.Parallel()

		t.Run("Should compose the soul from the identity facts", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingMintBeforeEach(t)
			initSoulSuite(suite)

			agent := model_builders.NewAgentBuilder().
				WithName("Alice").
				WithEnsDomain("alice.agent.eth").
				WithAttributes(map[string]string{"specialty": "defi"}).
				Build()
			savedAgent, _, err := suite.sut.ReconcileMint(
				context.Background(),
				agent,
				nil,
				common_tests.RandomEthAddress(),
				5,
				nil,
			)

			assert.NoError(t, err)
			assert.Contains(t, lo.FromPtr(savedAgent.Soul), "### Core identity")
			assert.Contains(t, lo.FromPtr(savedAgent.Soul), "I am Alice")
			assert.Contains(t, lo.FromPtr(savedAgent.Soul), "alice.agent.eth")
			assert.Contains(t, lo.FromPtr(savedAgent.Soul), "### Values")
			assert.Contains(t, lo.FromPtr(savedAgent.Soul), "### Traits")
			assert.Contains(t, lo.FromPtr(savedAgent.Soul), "- specialty: defi")
			assert.Contains(t, lo.FromPtr(savedAgent.Soul), "### Communication style")
		})
	})

	t.Run("Given a soul already exists", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recompute it from the current facts", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingMintBeforeEach(t)
			initSoulSuite(suite)

			agent := model_builders.NewAgentBuilder().
				WithName("Alice").
				WithSoul("stale persona").
				Build()

			savedAgent, _, err := suite.
				sut.
				ReconcileMint(context.Background(), agent, nil, common_tests.RandomEthAddress(), 5, nil)

			assert.NoError(t, err)
			assert.NotEqual(t, "stale persona", lo.FromPtr(savedAgent.Soul))
			assert.Contains(t, lo.FromPtr(savedAgent.Soul), "I am Alice")
		})
	})
}
