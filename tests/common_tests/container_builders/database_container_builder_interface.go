package container_builders

// DatabaseContainerBuilderInterface lets integration tests treat each
// supported engine the same way: spin up a scratch DB, get a connection
// string, drop it on cleanup.
type DatabaseContainerBuilderInterface interface {
	Create() (string, string)
	DropDatabase(dbName string)
}
