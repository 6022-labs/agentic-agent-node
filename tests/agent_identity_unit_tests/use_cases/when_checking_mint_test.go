package use_cases_test

import (
	"context"
	"errors"
	"testing"

	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/metrics_mocks/interfaces_mocks"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCheckingMintTestingSuite struct {
	sut *use_cases.CheckAgentMint

	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
	mockMintPublisher                    *interface_mocks.MockMintPublisherInterface
	mockRegistrationReconciler           *interface_mocks.MockRegistrationReconcilerInterface
}

func WhenCheckingMintBeforeEach(t *testing.T) *WhenCheckingMintTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(
		mockController,
	)
	mockMintPublisher := interface_mocks.NewMockMintPublisherInterface(mockController)
	mockRegistrationReconciler := interface_mocks.NewMockRegistrationReconcilerInterface(mockController)

	mockMintMetrics := metrics_mocks.NewMockMintMetricsInterface(mockController)
	mockMintMetrics.EXPECT().RecordMintSucceeded(gomock.Any()).AnyTimes()
	mockMintMetrics.EXPECT().RecordMintFailed(gomock.Any()).AnyTimes()

	sut := use_cases.NewCheckAgentMint(
		mockAgentRepository,
		mockAgentRegistrationStateRepository,
		mockMintPublisher,
		mockRegistrationReconciler,
		mockMintMetrics,
	)

	return &WhenCheckingMintTestingSuite{
		sut:                                  sut,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
		mockMintPublisher:                    mockMintPublisher,
		mockRegistrationReconciler:           mockRegistrationReconciler,
	}
}

func TestWhenCheckingMint(t *testing.T) {
	t.Parallel()

	t.Run("Given the mint is already recorded in the database", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingMintTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(7).Build()
			state := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusConfirmed).
				Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(state, nil)
		}

		t.Run("Should return confirmed without reading the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingMintBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, string(identity_enums.TransactionStatusConfirmed), resp.MintStatus)
			assert.NotNil(t, resp.TokenId)
			assert.Equal(t, uint64(7), *resp.TokenId)
		})
	})

	t.Run("Given the chain read fails", func(t *testing.T) {
		t.Parallel()

		rpcErr := errors.New("rpc down")
		initSuite := func(suite *WhenCheckingMintTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				Get(gomock.Any()).
				Return(model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build(), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewAgentRegistrationStateBuilder().Build(), nil)
			suite.mockMintPublisher.EXPECT().
				ResolveMintedToken(gomock.Any(), gomock.Any()).
				Return(common.Address{}, uint64(0), rpcErr)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingMintBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "chain_read_failed", target.Code)
				assert.Equal(t, "failed to read the minted token from chain", target.Message)
			}
			assert.ErrorIs(t, err, rpcErr)
		})
	})

	t.Run("Given no token is found on chain yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingMintTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			state := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusPending).
				Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockMintPublisher.EXPECT().
				ResolveMintedToken(gomock.Any(), gomock.Any()).
				Return(common.Address{}, uint64(0), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(state, nil)
		}

		t.Run("Should report mint still pending", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingMintBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, string(identity_enums.TransactionStatusPending), resp.MintStatus)
		})
	})

	t.Run("Given the node wallet now controls a token", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingMintTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			savedAgent := model_builders.NewAgentBuilder().WithCollectionTokenId(7).Build()
			savedState := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusConfirmed).
				Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockMintPublisher.EXPECT().
				ResolveMintedToken(gomock.Any(), gomock.Any()).
				Return(common_tests.RandomEthAddress(), uint64(7), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(nil, nil)
			suite.mockRegistrationReconciler.EXPECT().
				ReconcileMint(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), uint64(7), gomock.Any()).
				Return(savedAgent, savedState, nil)
		}

		t.Run("Should confirm the mint and return the tokenId", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingMintBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, string(identity_enums.TransactionStatusConfirmed), resp.MintStatus)
			assert.NotNil(t, resp.TokenId)
			assert.Equal(t, uint64(7), *resp.TokenId)
		})
	})
}
