package models

import "mime/multipart"

// ExternalKeyValue mirrors the runtime KeyValue tuple.
type ExternalKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ExternalSaveAgentRegistrationPayload is the runtime save-endpoint input; owner
// injected by the gateway. The client serializes it as multipart/form-data,
// streaming Image as the file part.
type ExternalSaveAgentRegistrationPayload struct {
	CollectionAddress string
	Owner             string
	Name              string
	Role              string
	CloneOf           *uint64
	Image             *multipart.FileHeader
	Attributes        map[string]string
}

// ExternalAgentRegistrationPayload mirrors the runtime draft + status response.
type ExternalAgentRegistrationPayload struct {
	ChainId           string            `json:"chainId"`
	CollectionAddress string            `json:"collectionAddress"`
	TokenId           *uint64           `json:"tokenId"`
	ProposalId        *uint64           `json:"proposalId"`
	Name              string            `json:"name"`
	Role              string            `json:"role"`
	CloneOf           *uint64           `json:"cloneOf"`
	Owner             string            `json:"owner"`
	WalletAddress     string            `json:"walletAddress"`
	EnsDomain         string            `json:"ensDomain"`
	Images            map[string]string `json:"images"`
	Attributes        map[string]string `json:"attributes"`
	MintStatus        string            `json:"mintStatus"`
	EnsStatus         string            `json:"ensStatus"`
}

// ExternalMintInfoPayload mirrors the runtime mint() calldata response.
type ExternalMintInfoPayload struct {
	CollectionAddress string             `json:"collectionAddress"`
	To                string             `json:"to"`
	Role              string             `json:"role"`
	Name              string             `json:"name"`
	Wallets           []string           `json:"wallets"`
	Images            []ExternalKeyValue `json:"images"`
	Attributes        []ExternalKeyValue `json:"attributes"`
	CloneOf           *uint64            `json:"cloneOf"`
	ModeratorCount    uint64             `json:"moderatorCount"`
	OwnerIsModerator  bool               `json:"ownerIsModerator"`
}

// ExternalEnsInfoPayload mirrors the runtime initAgentTexts calldata response.
type ExternalEnsInfoPayload struct {
	CollectionAddress string             `json:"collectionAddress"`
	TokenId           uint64             `json:"tokenId"`
	Records           []ExternalKeyValue `json:"records"`
}

// ExternalRegistrationStatusPayload mirrors the runtime check/self status response.
type ExternalRegistrationStatusPayload struct {
	MintStatus string  `json:"mintStatus"`
	EnsStatus  string  `json:"ensStatus"`
	TokenId    *uint64 `json:"tokenId"`
	ProposalId *uint64 `json:"proposalId"`
	MintTxHash *string `json:"mintTxHash"`
	EnsTxHash  *string `json:"ensTxHash"`
}
