package models

// AgentCollection is one collection governed by the on-chain manager, as
// returned by its paginated listing. Not persisted — always read from chain.
type AgentCollection struct {
	Address        string
	Name           string
	Description    string
	ModeratorCount uint64
	NextTokenId    uint64
}
