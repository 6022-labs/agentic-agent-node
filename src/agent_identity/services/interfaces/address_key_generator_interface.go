package interfaces

import "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"

// AddressKeyGeneratorInterface generates a fresh keypair for one chain family;
// the credential store dispatches to it by AddressType.
type AddressKeyGeneratorInterface interface {
	Type() enums.AddressType
	Generate() (*WalletCredentials, error)
}
