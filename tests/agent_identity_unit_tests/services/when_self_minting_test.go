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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSelfMintingTestingSuite struct {
	sut *services.MintPublisher

	mockAgentCollectionRequester         *interface_mocks.MockAgentCollectionRequesterInterface
	mockAgentCollectionsManagerRequester *interface_mocks.MockAgentCollectionsManagerRequesterInterface
	mockAgentMinter                      *interface_mocks.MockAgentMinterInterface
	mockWalletCredentialsStorage         *interface_mocks.MockWalletCredentialsStorageInterface
	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
	mockRegistrationReconciler           *interface_mocks.MockRegistrationReconcilerInterface
}

func WhenSelfMintingBeforeEach(t *testing.T) *WhenSelfMintingTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentCollectionRequester := interface_mocks.NewMockAgentCollectionRequesterInterface(mockController)
	mockAgentCollectionsManagerRequester := interface_mocks.NewMockAgentCollectionsManagerRequesterInterface(mockController)
	mockAgentMinter := interface_mocks.NewMockAgentMinterInterface(mockController)
	mockWalletCredentialsStorage := interface_mocks.NewMockWalletCredentialsStorageInterface(mockController)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(mockController)
	mockRegistrationReconciler := interface_mocks.NewMockRegistrationReconcilerInterface(mockController)

	sut := services.NewMintPublisher(
		mockAgentCollectionRequester,
		mockAgentCollectionsManagerRequester,
		mockAgentMinter,
		mockWalletCredentialsStorage,
		mockAgentRepository,
		mockAgentRegistrationStateRepository,
		mockRegistrationReconciler,
	)

	return &WhenSelfMintingTestingSuite{
		sut:                                  sut,
		mockAgentCollectionRequester:         mockAgentCollectionRequester,
		mockAgentCollectionsManagerRequester: mockAgentCollectionsManagerRequester,
		mockAgentMinter:                      mockAgentMinter,
		mockWalletCredentialsStorage:         mockWalletCredentialsStorage,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
		mockRegistrationReconciler:           mockRegistrationReconciler,
	}
}

func TestWhenSelfMinting(t *testing.T) {
	t.Parallel()

	owner := common_tests.RandomEthAddress().Hex()
	nodeWallet := common_tests.RandomEthAddress().Hex()

	draft := func() *models.Agent {
		return model_builders.NewAgentBuilder().WithoutCollectionTokenId().WithOwner(owner).Build()
	}

	t.Run("Given a non-moderated collection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSelfMintingTestingSuite) {
			mintedAgent := model_builders.NewAgentBuilder().WithCollectionTokenId(5).WithOwner(owner).Build()
			mintedState := model_builders.NewAgentRegistrationStateBuilder().WithMintStatus(identity_enums.TransactionStatusConfirmed).Build()

			suite.mockWalletCredentialsStorage.EXPECT().GetForId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&identity_interfaces.WalletCredentials{Address: nodeWallet}, true, nil)
			suite.mockAgentCollectionRequester.EXPECT().ModeratorCount(gomock.Any(), gomock.Any()).Return(uint64(0), nil)
			suite.mockAgentMinter.EXPECT().Mint(gomock.Any(), gomock.Any(), gomock.Any()).Return("0xtx", nil)
			suite.mockAgentCollectionRequester.EXPECT().AddressToTokenId(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm, gomock.Any()).Return(uint64(5), nil)
			suite.mockRegistrationReconciler.EXPECT().ReconcileMint(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), uint64(5), gomock.Any()).Return(mintedAgent, mintedState, nil)
		}

		t.Run("Should mint directly and reconcile the rows", func(t *testing.T) {
			t.Parallel()

			suite := WhenSelfMintingBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.SelfMint(context.Background(), draft(), nil)

			assert.NoError(t, err)
			assert.False(t, result.ProposalPending)
			assert.NotNil(t, result.Agent.CollectionTokenId)
			assert.Equal(t, uint64(5), *result.Agent.CollectionTokenId)
		})
	})

	t.Run("Given the agent carries evm and btc addresses", func(t *testing.T) {
		t.Parallel()

		staleEvm := common_tests.RandomEthAddress().Hex()
		btcAddress := "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4"

		agentWithAddresses := func() *models.Agent {
			return model_builders.NewAgentBuilder().
				WithoutCollectionTokenId().
				WithOwner(owner).
				WithAddress(identity_enums.AddressTypeEvm, staleEvm).
				WithAddress(identity_enums.AddressTypeBtc, btcAddress).
				Build()
		}

		initSuite := func(suite *WhenSelfMintingTestingSuite, captured *identity_interfaces.MintParams) {
			mintedAgent := model_builders.NewAgentBuilder().WithCollectionTokenId(5).WithOwner(owner).Build()
			mintedState := model_builders.NewAgentRegistrationStateBuilder().WithMintStatus(identity_enums.TransactionStatusConfirmed).Build()

			suite.mockWalletCredentialsStorage.EXPECT().GetForId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&identity_interfaces.WalletCredentials{Address: nodeWallet}, true, nil)
			suite.mockAgentCollectionRequester.EXPECT().ModeratorCount(gomock.Any(), gomock.Any()).Return(uint64(0), nil)
			suite.mockAgentMinter.EXPECT().Mint(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, params identity_interfaces.MintParams) (string, error) {
					*captured = params
					return "0xtx", nil
				})
			suite.mockAgentCollectionRequester.EXPECT().AddressToTokenId(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm, gomock.Any()).Return(uint64(5), nil)
			suite.mockRegistrationReconciler.EXPECT().ReconcileMint(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), uint64(5), gomock.Any()).Return(mintedAgent, mintedState, nil)
		}

		t.Run("Should send the evm node wallet first and the btc address along", func(t *testing.T) {
			t.Parallel()

			suite := WhenSelfMintingBeforeEach(t)
			var captured identity_interfaces.MintParams
			initSuite(suite, &captured)

			_, err := suite.sut.SelfMint(context.Background(), agentWithAddresses(), nil)

			assert.NoError(t, err)
			assert.Equal(t, []identity_interfaces.MintAddress{
				{AddressType: identity_enums.AddressTypeEvm, Value: nodeWallet},
				{AddressType: identity_enums.AddressTypeBtc, Value: btcAddress},
			}, captured.Addresses)
		})

		t.Run("Should use the Vault credential when agent_addresses disagrees", func(t *testing.T) {
			t.Parallel()

			suite := WhenSelfMintingBeforeEach(t)
			var captured identity_interfaces.MintParams
			initSuite(suite, &captured)

			agent := agentWithAddresses()
			_, err := suite.sut.SelfMint(context.Background(), agent, nil)

			assert.NoError(t, err)
			assert.Equal(t, nodeWallet, captured.Addresses[0].Value)
			assert.NotEqual(t, staleEvm, captured.Addresses[0].Value)
			assert.Equal(t, staleEvm, agent.Addresses[identity_enums.AddressTypeEvm])
		})
	})

	t.Run("Given a moderated collection where the node is a moderator", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSelfMintingTestingSuite) {
			mintedAgent := model_builders.NewAgentBuilder().WithCollectionTokenId(5).WithOwner(owner).Build()
			mintedState := model_builders.NewAgentRegistrationStateBuilder().WithMintStatus(identity_enums.TransactionStatusConfirmed).Build()

			suite.mockWalletCredentialsStorage.EXPECT().GetForId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&identity_interfaces.WalletCredentials{Address: nodeWallet}, true, nil)
			suite.mockAgentCollectionRequester.EXPECT().ModeratorCount(gomock.Any(), gomock.Any()).Return(uint64(2), nil)
			suite.mockAgentCollectionRequester.EXPECT().IsModerator(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			suite.mockAgentMinter.EXPECT().Mint(gomock.Any(), gomock.Any(), gomock.Any()).Return("0xtx", nil)
			suite.mockAgentCollectionRequester.EXPECT().AddressToTokenId(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm, gomock.Any()).Return(uint64(5), nil)
			suite.mockRegistrationReconciler.EXPECT().ReconcileMint(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), uint64(5), gomock.Any()).Return(mintedAgent, mintedState, nil)
		}

		t.Run("Should mint directly", func(t *testing.T) {
			t.Parallel()

			suite := WhenSelfMintingBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.SelfMint(context.Background(), draft(), nil)

			assert.NoError(t, err)
			assert.False(t, result.ProposalPending)
		})
	})

	t.Run("Given a moderated collection where the node is not a moderator", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSelfMintingTestingSuite) {
			suite.mockWalletCredentialsStorage.EXPECT().GetForId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&identity_interfaces.WalletCredentials{Address: nodeWallet}, true, nil)
			suite.mockAgentCollectionRequester.EXPECT().ModeratorCount(gomock.Any(), gomock.Any()).Return(uint64(2), nil)
			suite.mockAgentCollectionRequester.EXPECT().IsModerator(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
			suite.mockAgentMinter.EXPECT().CreateMintProposal(gomock.Any(), gomock.Any(), gomock.Any()).Return("0xtx", uint64(9), nil)
			suite.mockAgentRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, a *models.Agent) (*models.Agent, error) { return a, nil })
			suite.mockAgentRegistrationStateRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, s *models.AgentRegistrationState) (*models.AgentRegistrationState, error) {
					return s, nil
				})
		}

		t.Run("Should create a mint proposal and leave the mint submitted", func(t *testing.T) {
			t.Parallel()

			suite := WhenSelfMintingBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.SelfMint(context.Background(), draft(), nil)

			assert.NoError(t, err)
			assert.True(t, result.ProposalPending)
			assert.NotNil(t, result.Agent.ProposalId)
			assert.Equal(t, uint64(9), *result.Agent.ProposalId)
			assert.Equal(t, identity_enums.TransactionStatusSubmitted, result.State.MintStatus)
		})
	})

	t.Run("Given the mint lands but the token cannot be read back", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSelfMintingTestingSuite) {
			suite.mockWalletCredentialsStorage.EXPECT().GetForId(gomock.Any(), gomock.Any(), gomock.Any()).Return(&identity_interfaces.WalletCredentials{Address: nodeWallet}, true, nil)
			suite.mockAgentCollectionRequester.EXPECT().ModeratorCount(gomock.Any(), gomock.Any()).Return(uint64(0), nil)
			suite.mockAgentMinter.EXPECT().Mint(gomock.Any(), gomock.Any(), gomock.Any()).Return("0xtx", nil)
			suite.mockAgentCollectionRequester.EXPECT().AddressToTokenId(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm, gomock.Any()).Return(uint64(0), nil)
		}

		t.Run("Should return an error without reconciling", func(t *testing.T) {
			t.Parallel()

			suite := WhenSelfMintingBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.SelfMint(context.Background(), draft(), nil)

			assert.Error(t, err)
		})
	})
}
