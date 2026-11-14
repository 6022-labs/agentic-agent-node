package container_builders

// DatabaseContainerBuilderInterface gives every engine the same scratch-DB flow: create, connect, drop on cleanup.
type DatabaseContainerBuilderInterface interface {
	Create() (string, string)
	DropDatabase(dbName string)
}

// TlsServerName is the only name the TLS builders' server certificates are valid for.
const TlsServerName = "db.test.internal"

// TlsDatabaseContainerBuilderInterface is a scratch-DB builder whose server presents a certificate signed by its CA.
type TlsDatabaseContainerBuilderInterface interface {
	DatabaseContainerBuilderInterface
	CaCertificatePem() []byte
}
