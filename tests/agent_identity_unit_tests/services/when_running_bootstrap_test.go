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

func TestWhenRunningBootstrap(t *testing.T) {
	t.Parallel()

	t.Run("Given the wallet store returns an existing wallet", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		wallet := identity_mocks.NewMockWalletCredentialsStorageInterface(ctrl)
		wallet.EXPECT().GetOrGenerate(gomock.Any()).Return(&identity_interfaces.WalletCredentials{
			PrivateKeyHex: "0xpk",
			Address:       "0xaddr",
		}, nil)
		state := services.NewBootstrapStateHolder()

		runner := services.NewBootstrapRunner(zap.NewNop(), wallet, state)

		t.Run("Should mark the holder as ready with the wallet address", func(t *testing.T) {
			err := runner.Run(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, enums.BootstrapStatusReady, state.Get().Status)
			assert.Equal(t, "0xaddr", state.Get().AgentAddress)
			assert.Empty(t, state.Get().LastError)
		})
	})

	t.Run("Given the wallet store fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		wallet := identity_mocks.NewMockWalletCredentialsStorageInterface(ctrl)
		wallet.EXPECT().GetOrGenerate(gomock.Any()).Return(nil, errors.New("vault locked"))
		state := services.NewBootstrapStateHolder()

		runner := services.NewBootstrapRunner(zap.NewNop(), wallet, state)

		t.Run("Should record the error and leave status non-ready", func(t *testing.T) {
			err := runner.Run(context.Background())
			assert.ErrorContains(t, err, "vault locked")
			assert.NotEqual(t, enums.BootstrapStatusReady, state.Get().Status)
			assert.Contains(t, state.Get().LastError, "vault locked")
		})
	})
}
