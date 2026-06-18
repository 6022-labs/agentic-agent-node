package interfaces

// EnsDomainBuilderInterface composes an agent's display domain, appending the
// configured base zone so callers never pass it.
type EnsDomainBuilderInterface interface {
	// Build returns <name>.<collection>.<chainId>.<baseDomain>, all lowercased.
	Build(name, collectionName, chainId string) string
}
