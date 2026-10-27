package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/ethereum/go-ethereum/crypto"
)

// AbstractWalletSigner carries the wallet-key load every signer needs; load is unexported so no port exposes the key.
type AbstractWalletSigner struct {
	agentRepository          repositories.AgentRepositoryInterface
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface
}

func NewAbstractWalletSigner(
	agentRepository repositories.AgentRepositoryInterface,
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface,
) *AbstractWalletSigner {
	return &AbstractWalletSigner{
		agentRepository:          agentRepository,
		walletCredentialsStorage: walletCredentialsStorage,
	}
}

func (s *AbstractWalletSigner) load(ctx context.Context) (*ecdsa.PrivateKey, error) {
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
