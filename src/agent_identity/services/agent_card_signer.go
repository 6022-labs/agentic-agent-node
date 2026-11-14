package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

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
	*AbstractWalletSigner
}

func NewAgentCardSigner(
	agentResolver identity_interfaces.AgentResolverInterface,
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface,
) *AgentCardSigner {
	return &AgentCardSigner{
		AbstractWalletSigner: NewAbstractWalletSigner(agentResolver, walletCredentialsStorage),
	}
}

// Fields are alphabetical so the marshalled header is already JCS-ordered.
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

	key, err := s.load(ctx)
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
