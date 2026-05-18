package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/models/enums"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenRunningBootstrapTestingSuite struct {
	sut   *services.BootstrapRunner
	state *services.BootstrapStateHolder

	mockWallet       *identity_mocks.MockWalletCredentialsStorageInterface
	mockEnsPublisher *identity_mocks.MockEnsRuntimeRecordsPublisherInterface
}

func WhenRunningBootstrapBeforeEach(t *testing.T) *WhenRunningBootstrapTestingSuite {
	mockController := gomock.NewController(t)

	mockWallet := identity_mocks.NewMockWalletCredentialsStorageInterface(mockController)
	mockEnsPublisher := identity_mocks.NewMockEnsRuntimeRecordsPublisherInterface(mockController)
	state := services.NewBootstrapStateHolder()
	sut := services.NewBootstrapRunner(zap.NewNop(), mockWallet, state, mockEnsPublisher)

	return &WhenRunningBootstrapTestingSuite{
		sut:   sut,
		state: state,

		mockWallet:       mockWallet,
		mockEnsPublisher: mockEnsPublisher,
	}
}

func TestWhenRunningBootstrap(t *testing.T) {
	t.Parallel()

	t.Run("Given the wallet store returns an existing wallet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRunningBootstrapTestingSuite) {
			suite.mockWallet.EXPECT().GetOrGenerate(gomock.Any()).Return(&identity_interfaces.WalletCredentials{
				PrivateKeyHex: "0xpk",
				Address:       "0xaddr",
			}, nil)
			suite.mockEnsPublisher.EXPECT().PublishRuntimeRecords(gomock.Any()).Return(nil)
		}

		t.Run("Should mark the holder as ready with the wallet address", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningBootstrapBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Run(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, enums.BootstrapStatusReady, suite.state.Get().Status)
			assert.Equal(t, "0xaddr", suite.state.Get().AgentAddress)
			assert.Empty(t, suite.state.Get().LastError)
		})
	})

	t.Run("Given the wallet store fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRunningBootstrapTestingSuite) {
			suite.mockWallet.EXPECT().GetOrGenerate(gomock.Any()).Return(nil, errors.New("vault locked"))
		}

		t.Run("Should record the error and leave status non-ready", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningBootstrapBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Run(context.Background())

			assert.ErrorContains(t, err, "vault locked")
			assert.NotEqual(t, enums.BootstrapStatusReady, suite.state.Get().Status)
			assert.Contains(t, suite.state.Get().LastError, "vault locked")
		})
	})
}
