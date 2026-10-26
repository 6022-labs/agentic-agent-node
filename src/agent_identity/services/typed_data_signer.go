package services

import (
	"context"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const digestLength = 32

// TypedDataSigner signs EIP-712 digests; AgentCardSigner commits to a SHA-256 JWS digest instead.
type TypedDataSigner struct {
	*AbstractWalletSigner
}

func NewTypedDataSigner(
	agentRepository repositories.AgentRepositoryInterface,
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface,
) *TypedDataSigner {
	return &TypedDataSigner{
		AbstractWalletSigner: NewAbstractWalletSigner(agentRepository, walletCredentialsStorage),
	}
}

func (s *TypedDataSigner) SignDigest(ctx context.Context, digest []byte) ([]byte, error) {
	if len(digest) != digestLength {
		return nil, fmt.Errorf("digest must be %d bytes", digestLength)
	}

	key, err := s.load(ctx)
	if err != nil {
		return nil, err
	}

	signature, err := crypto.Sign(digest, key)
	if err != nil {
		return nil, fmt.Errorf("sign digest: %w", err)
	}
	// go-ethereum yields v as 0/1; on-chain ecrecover and the x402 verifier read 27/28.
	signature[crypto.RecoveryIDOffset] += 27
	return signature, nil
}

func (s *TypedDataSigner) SignTypedData(ctx context.Context, typedData *apitypes.TypedData) (string, error) {
	digest, _, err := apitypes.TypedDataAndHash(*typedData)
	if err != nil {
		return "", fmt.Errorf("hash typed data: %w", err)
	}

	signature, err := s.SignDigest(ctx, digest)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(signature), nil
}

func (s *TypedDataSigner) WalletAddress(ctx context.Context) (common.Address, error) {
	key, err := s.load(ctx)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(key.PublicKey), nil
}
