package container_builders

// DatabaseContainerBuilderInterface gives every engine the same scratch-DB flow: create, connect, drop on cleanup.
type DatabaseContainerBuilderInterface interface {
	Create() (string, string)
	DropDatabase(dbName string)
}
