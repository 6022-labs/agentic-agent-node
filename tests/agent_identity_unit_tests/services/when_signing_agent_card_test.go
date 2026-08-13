package services_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_services "github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// The cross-repo contract vector. agentic-orchestrator's when_verifying_agent_card_signature_test.go
// pins these exact literals — if either side changes, signed cards stop verifying in production.
const (
	// Anvil account 0, a publicly known test key.
	vectorPrivateKeyHex = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	vectorAddress       = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"

	vectorCard = `{"protocol":"6022","protocolVersion":"0.1","name":"hermes",` +
		`"ensDomain":"hermes.agents6022.eth","endpoints":{"responses":"https://agent.example.com/api/responses"},` +
		`"paymentMethods":["l402","x402"]}`
	vectorCanonical = `{"endpoints":{"responses":"https://agent.example.com/api/responses"},` +
		`"ensDomain":"hermes.agents6022.eth","name":"hermes","paymentMethods":["l402","x402"],"protocol":"6022",` +
		`"protocolVersion":"0.1"}`

	vectorProtected = "eyJhbGciOiJFUzI1NksiLCJqd2siOnsiY3J2Ijoic2VjcDI1NmsxIiwia3R5IjoiRUMiLCJ4IjoiZ3hoVFcxUVFYVXA2cm" +
		"1EQWo4UmZsb2NZRzBfZnhpVzlHblVfcHpsXzdYVSIsInkiOiJOVWZ4SEtocFprYnk4Nnl3ampFQmF2ckNQbU1NWFJIMW4ySC05WHNOS3FVIn" +
		"0sImtpZCI6IjB4ZjM5RmQ2ZTUxYWFkODhGNkY0Y2U2YUI4ODI3Mjc5Y2ZmRmI5MjI2NiIsInR5cCI6IkpPU0UifQ"
	vectorSignature = "HoLnS3IMLIFR8cgB35NLrQbvgxvUq_NFj6-cbEVbUm4CX_RV-lSHMNiy2kaKPbyvp8htlPvGlT0dTNJkHPjhvA"
)

type WhenSigningAgentCardTestingSuite struct {
	sut *identity_services.AgentCardSigner

	mockAgentRepository          *repositories_mocks.MockAgentRepositoryInterface
	mockWalletCredentialsStorage *identity_mocks.MockWalletCredentialsStorageInterface
}

func WhenSigningAgentCardBeforeEach(t *testing.T) *WhenSigningAgentCardTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockWalletCredentialsStorage := identity_mocks.NewMockWalletCredentialsStorageInterface(mockController)

	return &WhenSigningAgentCardTestingSuite{
		sut: identity_services.NewAgentCardSigner(
			mockAgentRepository,
			mockWalletCredentialsStorage,
		),
		mockAgentRepository:          mockAgentRepository,
		mockWalletCredentialsStorage: mockWalletCredentialsStorage,
	}
}

func TestWhenSigningAgentCard(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	agentId := uuid.MustParse("11111111-2222-3333-4444-555555555555")

	expectWallet := func(suite *WhenSigningAgentCardTestingSuite) {
		suite.mockAgentRepository.EXPECT().Get(ctx).Return(&models.Agent{Id: agentId}, nil)
		suite.mockWalletCredentialsStorage.EXPECT().
			GetForId(ctx, agentId.String(), enums.AddressTypeEvm).
			Return(&identity_interfaces.WalletCredentials{PrivateKeyHex: vectorPrivateKeyHex, Address: vectorAddress}, true, nil)
	}

	t.Run("Given the node wallet is bootstrapped", func(t *testing.T) {
		t.Parallel()

		t.Run("Should produce the pinned cross-repo vector", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			expectWallet(suite)

			signature, err := suite.sut.Sign(ctx, []byte(vectorCard))

			assert.NoError(t, err)
			if assert.NotNil(t, signature) {
				assert.Equal(t, vectorProtected, signature.Protected)
				assert.Equal(t, vectorSignature, signature.Signature)
			}
		})

		t.Run("Should sign the JCS canonical form, verifiable with the embedded jwk", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			expectWallet(suite)

			signature, err := suite.sut.Sign(ctx, []byte(vectorCard))
			assert.NoError(t, err)

			header := decodeProtectedHeader(t, signature.Protected)
			assert.Equal(t, "ES256K", header.Algorithm)
			assert.Equal(t, "JOSE", header.Type)
			assert.Equal(t, vectorAddress, header.KeyId)
			assert.Equal(t, "secp256k1", header.Jwk.Curve)
			assert.Equal(t, "EC", header.Jwk.KeyType)

			// Rebuild the signing input independently from the canonical form the spec mandates.
			digest := sha256.Sum256([]byte(signature.Protected + "." + encodeBase64Url([]byte(vectorCanonical))))
			publicKey := uncompressedPublicKey(t, header.Jwk)
			rawSignature, err := base64.RawURLEncoding.DecodeString(signature.Signature)
			assert.NoError(t, err)
			assert.Len(t, rawSignature, 64)
			assert.True(t, crypto.VerifySignature(publicKey, digest[:], rawSignature))
		})

		t.Run("Should key the jwk to the wallet the kid advertises", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			expectWallet(suite)

			signature, err := suite.sut.Sign(ctx, []byte(vectorCard))
			assert.NoError(t, err)

			header := decodeProtectedHeader(t, signature.Protected)
			publicKey, err := crypto.UnmarshalPubkey(uncompressedPublicKey(t, header.Jwk))
			assert.NoError(t, err)
			assert.Equal(t, vectorAddress, crypto.PubkeyToAddress(*publicKey).Hex())
		})

		t.Run("Should ignore member order, so the same card in any order signs identically", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			expectWallet(suite)

			reordered := `{"paymentMethods":["l402","x402"],"name":"hermes","protocolVersion":"0.1",` +
				`"endpoints":{"responses":"https://agent.example.com/api/responses"},"protocol":"6022",` +
				`"ensDomain":"hermes.agents6022.eth"}`
			signature, err := suite.sut.Sign(ctx, []byte(reordered))

			assert.NoError(t, err)
			assert.Equal(t, vectorSignature, signature.Signature)
		})
	})

	t.Run("Given the card is not valid JSON", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail rather than sign unparsed bytes", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)

			signature, err := suite.sut.Sign(ctx, []byte(`{"protocol":`))

			assert.Error(t, err)
			assert.Nil(t, signature)
		})
	})

	t.Run("Given no agent row exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(nil, nil)
		}

		t.Run("Should fail rather than sign with no identity", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Sign(ctx, []byte(vectorCard))

			assert.Error(t, err)
		})
	})

	t.Run("Given the node wallet is not bootstrapped", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(&models.Agent{Id: agentId}, nil)
			suite.mockWalletCredentialsStorage.EXPECT().
				GetForId(ctx, agentId.String(), enums.AddressTypeEvm).
				Return(nil, false, nil)
		}

		t.Run("Should fail rather than serve an unsigned card silently", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Sign(ctx, []byte(vectorCard))

			assert.Error(t, err)
		})
	})
}

type testProtectedHeader struct {
	Algorithm string           `json:"alg"`
	Jwk       testSecp256k1Jwk `json:"jwk"`
	KeyId     string           `json:"kid"`
	Type      string           `json:"typ"`
}

type testSecp256k1Jwk struct {
	Curve   string `json:"crv"`
	KeyType string `json:"kty"`
	X       string `json:"x"`
	Y       string `json:"y"`
}

func decodeProtectedHeader(t *testing.T, protected string) testProtectedHeader {
	t.Helper()

	raw, err := base64.RawURLEncoding.DecodeString(protected)
	assert.NoError(t, err)

	var header testProtectedHeader
	assert.NoError(t, json.Unmarshal(raw, &header))
	return header
}

func uncompressedPublicKey(t *testing.T, jwk testSecp256k1Jwk) []byte {
	t.Helper()

	x, err := base64.RawURLEncoding.DecodeString(jwk.X)
	assert.NoError(t, err)
	y, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	assert.NoError(t, err)

	return append(append([]byte{0x04}, x...), y...)
}

func encodeBase64Url(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
