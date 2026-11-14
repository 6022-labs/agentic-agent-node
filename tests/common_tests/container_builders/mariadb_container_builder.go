package container_builders

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/certificate_builders"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var mariadbContainer *MariaDBContainerBuilder

type MariaDBContainerBuilder struct {
	container       testcontainers.Container
	connStrTemplate string
	containerMu     sync.Mutex

	tlsContainer       testcontainers.Container
	tlsConnStrTemplate string
	caCertificatePem   []byte
	tlsContainerMu     sync.Mutex

	databasesMu sync.Mutex
	databases   []string
}

func GetMariaDBContainerBuilder() *MariaDBContainerBuilder {
	if mariadbContainer == nil {
		mariadbContainer = &MariaDBContainerBuilder{}
	}
	return mariadbContainer
}

func (b *MariaDBContainerBuilder) start() string {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container == nil {
		req := b.request()
		// The image enables TLS by default; plaintext lets tests prove there is no fallback to it.
		req.Cmd = []string{"--skip-ssl"}

		b.container, b.connStrTemplate = b.run(req)
		log.Println("MariaDB container ready")
	}
	return b.connStrTemplate
}

func (b *MariaDBContainerBuilder) startTls() (string, []byte) {
	b.tlsContainerMu.Lock()
	defer b.tlsContainerMu.Unlock()

	if b.tlsContainer == nil {
		certificateAuthority := certificate_builders.NewCertificateAuthority()
		certificatePem, keyPem := certificateAuthority.IssueServerCertificate(TlsServerName)

		req := b.request()
		req.Files = []testcontainers.ContainerFile{
			{Reader: bytes.NewReader(certificatePem), ContainerFilePath: "/certs/server.crt", FileMode: 0o644},
			{Reader: bytes.NewReader(keyPem), ContainerFilePath: "/certs/server.key", FileMode: 0o644},
		}
		req.Entrypoint = []string{"sh", "-c", "install -o mysql -g mysql -m 600 /certs/server.key /tmp/server.key" +
			" && exec docker-entrypoint.sh mariadbd --ssl-cert=/certs/server.crt --ssl-key=/tmp/server.key"}

		var connStrTemplate string
		b.tlsContainer, connStrTemplate = b.run(req)
		// skip-verify lets the builder create databases without trusting the CA.
		b.tlsConnStrTemplate = connStrTemplate + "&tls=skip-verify"
		b.caCertificatePem = certificateAuthority.CertificatePem
		log.Println("MariaDB TLS container ready")
	}
	return b.tlsConnStrTemplate, b.caCertificatePem
}

func (b *MariaDBContainerBuilder) request() testcontainers.ContainerRequest {
	return testcontainers.ContainerRequest{
		Image:        "mariadb:12.3",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MARIADB_ROOT_PASSWORD": "root",
			"MARIADB_DATABASE":      "mariadb",
		},
		WaitingFor: wait.ForListeningPort("3306/tcp"),
	}
}

func (b *MariaDBContainerBuilder) run(req testcontainers.ContainerRequest) (testcontainers.Container, string) {
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start mariadb container: %v", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "3306")
	return container, fmt.Sprintf("root:root@tcp(%s:%s)/%%s?parseTime=true", host, port.Port())
}

func (b *MariaDBContainerBuilder) stop() {
	b.containerMu.Lock()
	if b.container != nil {
		_ = b.container.Terminate(context.Background())
		b.container = nil
		log.Println("MariaDB container stopped")
	}
	b.containerMu.Unlock()

	b.tlsContainerMu.Lock()
	if b.tlsContainer != nil {
		_ = b.tlsContainer.Terminate(context.Background())
		b.tlsContainer = nil
		log.Println("MariaDB TLS container stopped")
	}
	b.tlsContainerMu.Unlock()

	mariadbContainer = nil
}

func (b *MariaDBContainerBuilder) Create() (string, string) {
	dnsTemplate := b.start()
	dbName := b.createDatabase(dnsTemplate)
	return dbName, fmt.Sprintf(dnsTemplate, dbName)
}

func (b *MariaDBContainerBuilder) CreateTls() (string, string, []byte) {
	dnsTemplate, caCertificatePem := b.startTls()
	dbName := b.createDatabase(dnsTemplate)
	return dbName, fmt.Sprintf(dnsTemplate, dbName), caCertificatePem
}

func (b *MariaDBContainerBuilder) createDatabase(dnsTemplate string) string {
	conn, err := sql.Open("mysql", fmt.Sprintf(dnsTemplate, "mariadb"))
	if err != nil {
		log.Fatalf("failed to open mariadb: %v", err)
	}
	defer conn.Close()

	dbName := "test_" + uuid.New().String()[:8]
	if _, err := conn.Exec(fmt.Sprintf("CREATE DATABASE `%s`;", dbName)); err != nil {
		log.Fatalf("failed to create db %s: %v", dbName, err)
	}

	b.databasesMu.Lock()
	b.databases = append(b.databases, dbName)
	b.databasesMu.Unlock()

	return dbName
}

func (b *MariaDBContainerBuilder) DropDatabase(dbName string) {
	b.databasesMu.Lock()
	defer b.databasesMu.Unlock()
	for i, db := range b.databases {
		if db == dbName {
			b.databases = append(b.databases[:i], b.databases[i+1:]...)
			break
		}
	}
}
