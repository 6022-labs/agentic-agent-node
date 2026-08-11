package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gowebpki/jcs"
)

const (
	cardSignatureAlgorithm = "ES256K"
	cardSignatureType      = "JOSE"
	jwkKeyType             = "EC"
	jwkCurve               = "secp256k1"
	secp256k1CoordinateLen = 32
)

// AgentCardSigner signs agent cards as A2A detached JWS (ES256K, RFC 8812) with the node's EVM wallet key.
type AgentCardSigner struct {
	agentRepository          repositories.AgentRepositoryInterface
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface
}

func NewAgentCardSigner(
	agentRepository repositories.AgentRepositoryInterface,
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface,
) *AgentCardSigner {
	return &AgentCardSigner{
		agentRepository:          agentRepository,
		walletCredentialsStorage: walletCredentialsStorage,
	}
}

// jwsProtectedHeader fields are declared alphabetically so the marshalled header is already JCS-ordered.
type jwsProtectedHeader struct {
	Algorithm string       `json:"alg"`
	Jwk       secp256k1Jwk `json:"jwk"`
	KeyId     string       `json:"kid"`
	Type      string       `json:"typ"`
}

type secp256k1Jwk struct {
	Curve   string `json:"crv"`
	KeyType string `json:"kty"`
	X       string `json:"x"`
	Y       string `json:"y"`
}

func (s *AgentCardSigner) Sign(ctx context.Context, card []byte) (*identity_interfaces.AgentCardSignature, error) {
	canonical, err := jcs.Transform(card)
	if err != nil {
		return nil, fmt.Errorf("canonicalize agent card: %w", err)
	}

	key, err := s.loadWalletKey(ctx)
	if err != nil {
		return nil, err
	}

	headerBytes, err := json.Marshal(s.buildProtectedHeader(key))
	if err != nil {
		return nil, fmt.Errorf("encode protected header: %w", err)
	}
	protected := s.encodeBase64Url(headerBytes)

	// SHA-256 is what keeps this key path apart from transactions and personal_sign, which commit to Keccak-256.
	digest := sha256.Sum256([]byte(protected + "." + s.encodeBase64Url(canonical)))
	signature, err := crypto.Sign(digest[:], key)
	if err != nil {
		return nil, fmt.Errorf("sign agent card: %w", err)
	}

	// ES256K is r||s only; go-ethereum appends a recovery byte.
	return &identity_interfaces.AgentCardSignature{
		Protected: protected,
		Signature: s.encodeBase64Url(signature[:crypto.SignatureLength-1]),
	}, nil
}

func (s *AgentCardSigner) encodeBase64Url(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func (s *AgentCardSigner) buildProtectedHeader(key *ecdsa.PrivateKey) jwsProtectedHeader {
	// Uncompressed public key is 0x04 || X || Y.
	publicKey := crypto.FromECDSAPub(&key.PublicKey)

	return jwsProtectedHeader{
		Algorithm: cardSignatureAlgorithm,
		Jwk: secp256k1Jwk{
			Curve:   jwkCurve,
			KeyType: jwkKeyType,
			X:       s.encodeBase64Url(publicKey[1 : 1+secp256k1CoordinateLen]),
			Y:       s.encodeBase64Url(publicKey[1+secp256k1CoordinateLen:]),
		},
		KeyId: crypto.PubkeyToAddress(key.PublicKey).Hex(),
		Type:  cardSignatureType,
	}
}

func (s *AgentCardSigner) loadWalletKey(ctx context.Context) (*ecdsa.PrivateKey, error) {
	agent, err := s.agentRepository.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("load agent: %w", err)
	}
	if agent == nil {
		return nil, fmt.Errorf("agent not registered")
	}

	cred, found, err := s.walletCredentialsStorage.GetForId(ctx, agent.Id.String(), enums.AddressTypeEvm)
	if err != nil {
		return nil, fmt.Errorf("load node wallet: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("node wallet not bootstrapped")
	}

	key, err := crypto.HexToECDSA(
		strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(cred.PrivateKeyHex), "0x"), "0X"),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet private key: %w", err)
	}
	return key, nil
}
