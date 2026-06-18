package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	identity_settings "github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const recoveryChainId = "80002"

type WhenRecoveringAgentFromChainTestingSuite struct {
	sut *services.AgentChainRecoverer

	mockWalletCredentialsStorage         *interface_mocks.MockWalletCredentialsStorageInterface
	mockAgentCollectionsManagerRequester *interface_mocks.MockAgentCollectionsManagerRequesterInterface
	mockAgentCollectionRequester         *interface_mocks.MockAgentCollectionRequesterInterface
	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
	mockRegistrationReconciler           *interface_mocks.MockRegistrationReconcilerInterface
	mockEnsRegistrar                     *interface_mocks.MockEnsRegistrarInterface
}

func WhenRecoveringAgentFromChainBeforeEach(t *testing.T) *WhenRecoveringAgentFromChainTestingSuite {
	mockController := gomock.NewController(t)

	mockWalletCredentialsStorage := interface_mocks.NewMockWalletCredentialsStorageInterface(mockController)
	mockAgentCollectionsManagerRequester := interface_mocks.NewMockAgentCollectionsManagerRequesterInterface(mockController)
	mockAgentCollectionRequester := interface_mocks.NewMockAgentCollectionRequesterInterface(mockController)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(mockController)
	mockRegistrationReconciler := interface_mocks.NewMockRegistrationReconcilerInterface(mockController)
	mockEnsRegistrar := interface_mocks.NewMockEnsRegistrarInterface(mockController)

	managersSettings := runtime_common_settings.AgentCollectionsManagersSettings{recoveryChainId: common_tests.RandomEthAddress()}

	sut := services.NewAgentChainRecoverer(
		zap.NewNop(),
		mockWalletCredentialsStorage,
		mockAgentCollectionsManagerRequester,
		mockAgentCollectionRequester,
		mockAgentRepository,
		mockAgentRegistrationStateRepository,
		mockRegistrationReconciler,
		mockEnsRegistrar,
		services.NewEnsDomainBuilder(&identity_settings.AgentIdentitySettings{EnsBaseDomain: "6022.eth"}),
		&managersSettings,
	)

	return &WhenRecoveringAgentFromChainTestingSuite{
		sut:                                  sut,
		mockWalletCredentialsStorage:         mockWalletCredentialsStorage,
		mockAgentCollectionsManagerRequester: mockAgentCollectionsManagerRequester,
		mockAgentCollectionRequester:         mockAgentCollectionRequester,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
		mockRegistrationReconciler:           mockRegistrationReconciler,
		mockEnsRegistrar:                     mockEnsRegistrar,
	}
}

func TestWhenRecoveringAgentFromChain(t *testing.T) {
	t.Parallel()

	agentId := uuid.New()
	walletAddress := common_tests.RandomEthAddress().Hex()
	collection := common_tests.RandomEthAddress()
	creator := common_tests.RandomEthAddress()

	onChainInfo := identity_interfaces.OnChainAgentInfo{
		Name:       "alice",
		Role:       enums.AgentRoleExpert.String(),
		Creator:    creator,
		Images:     map[string]string{"default": "cid-default"},
		Attributes: map[string]string{"trait": "value"},
	}

	t.Run("Given no agent exists in the DB", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRecoveringAgentFromChainTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should report nothing recovered", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringAgentFromChainBeforeEach(t)
			initSuite(suite)

			recovered, err := suite.sut.Recover(context.Background())

			assert.NoError(t, err)
			assert.False(t, recovered)
		})
	})

	t.Run("Given the DB agent is already minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRecoveringAgentFromChainTestingSuite) {
			tokenId := uint64(9)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(&models.Agent{Id: agentId, CollectionTokenId: &tokenId}, nil)
		}

		t.Run("Should skip recovery without touching the wallet or chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringAgentFromChainBeforeEach(t)
			initSuite(suite)

			recovered, err := suite.sut.Recover(context.Background())

			assert.NoError(t, err)
			assert.False(t, recovered)
		})
	})

	t.Run("Given the unminted agent's wallet owns a minted, ENS-provisioned NFT", func(t *testing.T) {
		t.Parallel()

		var capturedAgent *models.Agent
		savedState := &models.AgentRegistrationState{}

		initSuite := func(suite *WhenRecoveringAgentFromChainTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(&models.Agent{Id: agentId}, nil)
			suite.mockWalletCredentialsStorage.EXPECT().GetForId(gomock.Any(), agentId.String()).Return(&identity_interfaces.WalletCredentials{Address: walletAddress}, true, nil)
			suite.mockAgentCollectionsManagerRequester.EXPECT().FindAgentByWallet(recoveryChainId, gomock.Any()).Return(collection, uint64(9), "alice", nil)
			suite.mockAgentCollectionRequester.EXPECT().InformationOf(recoveryChainId, collection, uint64(9)).Return(onChainInfo, nil)
			suite.mockAgentCollectionRequester.EXPECT().CollectionName(recoveryChainId, collection).Return("mycollection", nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
			suite.mockRegistrationReconciler.EXPECT().ReconcileMint(gomock.Any(), gomock.Any(), gomock.Any(), collection, uint64(9), gomock.Any()).DoAndReturn(
				func(_ context.Context, agent *models.Agent, _ *models.AgentRegistrationState, _ common.Address, _ uint64, _ *string) (*models.Agent, *models.AgentRegistrationState, error) {
					capturedAgent = agent
					return agent, savedState, nil
				})
			suite.mockEnsRegistrar.EXPECT().Provisioned(gomock.Any(), recoveryChainId, collection, "alice").Return(true, nil)
			suite.mockRegistrationReconciler.EXPECT().ReconcileEnsRegistry(gomock.Any(), savedState).Return(savedState, nil)
		}

		t.Run("Should recover the full identity in place and confirm ENS", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringAgentFromChainBeforeEach(t)
			initSuite(suite)

			recovered, err := suite.sut.Recover(context.Background())

			assert.NoError(t, err)
			assert.True(t, recovered)

			assert.Equal(t, agentId, capturedAgent.Id)
			assert.Equal(t, recoveryChainId, capturedAgent.ChainId)
			assert.Equal(t, collection.Hex(), capturedAgent.CollectionAddress)
			assert.Equal(t, walletAddress, capturedAgent.WalletAddress)
			assert.Equal(t, "alice", capturedAgent.Name)
			assert.Equal(t, enums.AgentRoleExpert, capturedAgent.Role)
			assert.Equal(t, "alice.mycollection.80002.6022.eth", capturedAgent.EnsDomain)
			assert.Equal(t, onChainInfo.Images, capturedAgent.Images)
			assert.Equal(t, onChainInfo.Attributes, capturedAgent.Attributes)
		})
	})

	t.Run("Given the unminted agent's wallet owns an NFT whose ENS is not provisioned", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRecoveringAgentFromChainTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(&models.Agent{Id: agentId}, nil)
			suite.mockWalletCredentialsStorage.EXPECT().GetForId(gomock.Any(), agentId.String()).Return(&identity_interfaces.WalletCredentials{Address: walletAddress}, true, nil)
			suite.mockAgentCollectionsManagerRequester.EXPECT().FindAgentByWallet(recoveryChainId, gomock.Any()).Return(collection, uint64(9), "alice", nil)
			suite.mockAgentCollectionRequester.EXPECT().InformationOf(recoveryChainId, collection, uint64(9)).Return(onChainInfo, nil)
			suite.mockAgentCollectionRequester.EXPECT().CollectionName(recoveryChainId, collection).Return("mycollection", nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
			suite.mockRegistrationReconciler.EXPECT().ReconcileMint(gomock.Any(), gomock.Any(), gomock.Any(), collection, uint64(9), gomock.Any()).Return(&models.Agent{}, &models.AgentRegistrationState{}, nil)
			suite.mockEnsRegistrar.EXPECT().Provisioned(gomock.Any(), recoveryChainId, collection, "alice").Return(false, nil)
		}

		t.Run("Should recover without confirming ENS", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringAgentFromChainBeforeEach(t)
			initSuite(suite)

			recovered, err := suite.sut.Recover(context.Background())

			assert.NoError(t, err)
			assert.True(t, recovered)
		})
	})

	t.Run("Given the unminted agent's wallet owns no NFT on any chain", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRecoveringAgentFromChainTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(&models.Agent{Id: agentId}, nil)
			suite.mockWalletCredentialsStorage.EXPECT().GetForId(gomock.Any(), agentId.String()).Return(&identity_interfaces.WalletCredentials{Address: walletAddress}, true, nil)
			suite.mockAgentCollectionsManagerRequester.EXPECT().FindAgentByWallet(recoveryChainId, gomock.Any()).Return(common.Address{}, uint64(0), "", nil)
		}

		t.Run("Should report nothing recovered without writing the agent", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringAgentFromChainBeforeEach(t)
			initSuite(suite)

			recovered, err := suite.sut.Recover(context.Background())

			assert.NoError(t, err)
			assert.False(t, recovered)
		})
	})
}
