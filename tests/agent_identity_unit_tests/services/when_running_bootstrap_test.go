package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenRunningBootstrapTestingSuite struct {
	sut *services.BootstrapRunner

	mockAgentRepository          *repository_mocks.MockAgentRepositoryInterface
	mockWalletCredentialsStorage *interface_mocks.MockWalletCredentialsStorageInterface
	mockEnsPublisher             *interface_mocks.MockEnsPublisherInterface
}

func WhenRunningBootstrapBeforeEach(t *testing.T) *WhenRunningBootstrapTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockWalletCredentialsStorage := interface_mocks.NewMockWalletCredentialsStorageInterface(mockController)
	mockEnsPublisher := interface_mocks.NewMockEnsPublisherInterface(mockController)

	sut := services.NewBootstrapRunner(
		zap.NewNop(),
		mockAgentRepository,
		mockWalletCredentialsStorage,
		mockEnsPublisher,
	)

	return &WhenRunningBootstrapTestingSuite{
		sut:                          sut,
		mockAgentRepository:          mockAgentRepository,
		mockWalletCredentialsStorage: mockWalletCredentialsStorage,
		mockEnsPublisher:             mockEnsPublisher,
	}
}

// expectEnsRefresh returns a channel closed when the background ENS refresh runs.
func expectEnsRefresh(suite *WhenRunningBootstrapTestingSuite) chan struct{} {
	done := make(chan struct{})
	suite.mockEnsPublisher.EXPECT().PublishRuntimeRecords(gomock.Any()).DoAndReturn(
		func(_ context.Context) error {
			close(done)
			return nil
		})
	return done
}

func waitForEnsRefresh(t *testing.T, done chan struct{}) {
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("ENS runtime records refresh was not triggered")
	}
}

func TestWhenRunningBootstrap(t *testing.T) {
	t.Parallel()

	cred := &identity_interfaces.WalletCredentials{PrivateKeyHex: "0xpk", Address: "0xaddr"}

	t.Run("Given no agent record exists yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRunningBootstrapTestingSuite) chan struct{} {
			created := model_builders.NewAgentBuilder().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
			suite.mockAgentRepository.EXPECT().Upsert(gomock.Any(), &models.Agent{}).Return(created, nil)
			suite.mockWalletCredentialsStorage.EXPECT().GetOrGenerate(gomock.Any()).Return(cred, nil)
			return expectEnsRefresh(suite)
		}

		t.Run("Should create an empty agent draft before readying the wallet", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningBootstrapBeforeEach(t)
			done := initSuite(suite)

			err := suite.sut.Run(context.Background())

			assert.NoError(t, err)
			waitForEnsRefresh(t, done)
		})
	})

	t.Run("Given an agent record already exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRunningBootstrapTestingSuite) chan struct{} {
			agent := model_builders.NewAgentBuilder().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockWalletCredentialsStorage.EXPECT().GetOrGenerate(gomock.Any()).Return(cred, nil)
			return expectEnsRefresh(suite)
		}

		t.Run("Should not create another agent", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningBootstrapBeforeEach(t)
			done := initSuite(suite)

			err := suite.sut.Run(context.Background())

			assert.NoError(t, err)
			waitForEnsRefresh(t, done)
		})
	})

	t.Run("Given ensuring the agent fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRunningBootstrapTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should error without touching the wallet", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningBootstrapBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Run(context.Background())

			assert.ErrorContains(t, err, "db down")
		})
	})

	t.Run("Given the wallet cannot be readied", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRunningBootstrapTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockWalletCredentialsStorage.EXPECT().GetOrGenerate(gomock.Any()).Return(nil, errors.New("vault down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningBootstrapBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Run(context.Background())

			assert.ErrorContains(t, err, "vault down")
		})
	})
}
