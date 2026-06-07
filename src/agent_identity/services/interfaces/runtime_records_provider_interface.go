package interfaces

// RuntimeRecordsProviderInterface builds the canonical set of ENS text records
// the agent publishes (the contract-fixed keys filled from runtime config).
type RuntimeRecordsProviderInterface interface {
	RuntimeRecords() []EnsRecord
}
