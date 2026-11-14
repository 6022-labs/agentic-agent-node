package services_test

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSigningTypedDataTestingSuite struct {
	sut *services.TypedDataSigner

	mockAgentResolver            *interfaces_mocks.MockAgentResolverInterface
	mockWalletCredentialsStorage *interfaces_mocks.MockWalletCredentialsStorageInterface
	walletPrivateKeyHex          string
}

func WhenSigningTypedDataBeforeEach(t *testing.T) *WhenSigningTypedDataTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentResolver := interfaces_mocks.NewMockAgentResolverInterface(mockController)
	mockWalletCredentialsStorage := interfaces_mocks.NewMockWalletCredentialsStorageInterface(mockController)

	key, err := crypto.GenerateKey()
	assert.NoError(t, err)

	return &WhenSigningTypedDataTestingSuite{
		sut: services.NewTypedDataSigner(
			mockAgentResolver,
			mockWalletCredentialsStorage,
		),
		mockAgentResolver:            mockAgentResolver,
		mockWalletCredentialsStorage: mockWalletCredentialsStorage,
		walletPrivateKeyHex:          hex.EncodeToString(crypto.FromECDSA(key)),
	}
}

// expectWallet makes the Vault-backed lookup answer with the suite's generated key.
func (s *WhenSigningTypedDataTestingSuite) expectWallet() {
	agent := model_builders.NewAgentBuilder().Build()
	s.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(agent, nil).AnyTimes()
	s.mockWalletCredentialsStorage.EXPECT().
		GetForId(gomock.Any(), agent.Id.String(), enums.AddressTypeEvm).
		Return(&identity_interfaces.WalletCredentials{PrivateKeyHex: s.walletPrivateKeyHex}, true, nil).
		AnyTimes()
}

func testTypedData() *apitypes.TypedData {
	return &apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"Mail": {{Name: "contents", Type: "string"}},
		},
		PrimaryType: "Mail",
		Domain: apitypes.TypedDataDomain{
			Name:    "Test",
			ChainId: math.NewHexOrDecimal256(80002),
		},
		Message: apitypes.TypedDataMessage{"contents": "hello"},
	}
}

func TestWhenSigningTypedData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a typed data payload and a bootstrapped wallet", func(t *testing.T) {
		t.Parallel()

		t.Run("Should produce a signature that recovers to the wallet address", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			suite.expectWallet()

			signature, err := suite.sut.SignTypedData(ctx, testTypedData())
			assert.NoError(t, err)

			address, err := suite.sut.WalletAddress(ctx)
			assert.NoError(t, err)

			digest, _, err := apitypes.TypedDataAndHash(*testTypedData())
			assert.NoError(t, err)

			raw, err := hex.DecodeString(signature[2:])
			assert.NoError(t, err)
			// v comes back as 27/28, which SigToPub does not accept.
			assert.Contains(t, []byte{27, 28}, raw[crypto.RecoveryIDOffset])
			raw[crypto.RecoveryIDOffset] -= 27

			publicKey, err := crypto.SigToPub(digest, raw)
			assert.NoError(t, err)
			assert.Equal(t, address, crypto.PubkeyToAddress(*publicKey))
		})
	})

	t.Run("Given a digest of the wrong length", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse to sign it", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)

			signature, err := suite.sut.SignDigest(ctx, []byte("short"))

			assert.Nil(t, signature)
			assert.ErrorContains(t, err, "32 bytes")
		})
	})

	t.Run("Given the wallet is not bootstrapped", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail rather than sign with a generated key", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			agent := model_builders.NewAgentBuilder().Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockWalletCredentialsStorage.EXPECT().
				GetForId(gomock.Any(), agent.Id.String(), enums.AddressTypeEvm).
				Return(nil, false, nil)

			signature, err := suite.sut.SignTypedData(ctx, testTypedData())

			assert.Empty(t, signature)
			assert.ErrorContains(t, err, "not bootstrapped")
		})
	})

	t.Run("Given the agent row cannot be read", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the failure", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db down"))

			address, err := suite.sut.WalletAddress(ctx)

			assert.Equal(t, "0x0000000000000000000000000000000000000000", address.Hex())
			assert.ErrorContains(t, err, "db down")
		})
	})
}
