package container_builders

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/certificate_builders"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var postgresContainer *PostgresContainerBuilder
var (
	postgresTlsContainer   *PostgresContainerBuilder
	postgresTlsContainerMu sync.Mutex
)

type PostgresContainerBuilder struct {
	certificateAuthority *certificate_builders.CertificateAuthority
	container            testcontainers.Container
	connStrTemplate      string
	containerMu          sync.Mutex
	databasesMu          sync.Mutex
	databases            []string
}

func GetPostgresContainerBuilder() *PostgresContainerBuilder {
	if postgresContainer == nil {
		postgresContainer = &PostgresContainerBuilder{}
	}
	return postgresContainer
}

// GetPostgresTlsContainerBuilder presents a certificate for TlsServerName signed by its own CA.
func GetPostgresTlsContainerBuilder() *PostgresContainerBuilder {
	postgresTlsContainerMu.Lock()
	defer postgresTlsContainerMu.Unlock()

	if postgresTlsContainer == nil {
		postgresTlsContainer = &PostgresContainerBuilder{
			certificateAuthority: certificate_builders.NewCertificateAuthority(),
		}
	}
	return postgresTlsContainer
}

func (b *PostgresContainerBuilder) CaCertificatePem() []byte {
	return b.certificateAuthority.CertificatePem
}

func (b *PostgresContainerBuilder) start() string {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container == nil {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "postgres:18-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "postgres",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		}
		sslMode := "disable"
		if b.certificateAuthority != nil {
			certificatePem, keyPem := b.certificateAuthority.IssueServerCertificate(TlsServerName)
			req.Files = []testcontainers.ContainerFile{
				{Reader: bytes.NewReader(certificatePem), ContainerFilePath: "/certs/server.crt", FileMode: 0o644},
				{Reader: bytes.NewReader(keyPem), ContainerFilePath: "/certs/server.key", FileMode: 0o644},
			}
			// postgres refuses a key file it does not own with group/other access.
			req.Entrypoint = []string{
				"sh",
				"-c",
				"install -o postgres -g postgres -m 600 /certs/server.key /tmp/server.key" +
					" && exec docker-entrypoint.sh postgres -c ssl=on" +
					" -c ssl_cert_file=/certs/server.crt -c ssl_key_file=/tmp/server.key",
			}
			sslMode = "require"
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			log.Fatalf("failed to start postgres container: %v", err)
		}
		b.container = container

		host, _ := container.Host(ctx)
		port, _ := container.MappedPort(ctx, "5432")

		b.connStrTemplate = fmt.Sprintf(
			"host=%s port=%s dbname=%s user=postgres password=postgres sslmode=%s",
			host, port.Port(), "%s", sslMode,
		)
		log.Println("PostgreSQL container ready")
	}
	return b.connStrTemplate
}

func (b *PostgresContainerBuilder) stop() {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container != nil {
		_ = b.container.Terminate(context.Background())
		b.container = nil
		log.Println("PostgreSQL container stopped")
	}
	if postgresContainer == b {
		postgresContainer = nil
	}
	postgresTlsContainerMu.Lock()
	if postgresTlsContainer == b {
		postgresTlsContainer = nil
	}
	postgresTlsContainerMu.Unlock()
}

// Create returns (dbName, connectionString) for a scratch DB in the shared container, so tests run in parallel.
func (b *PostgresContainerBuilder) Create() (string, string) {
	dnsTemplate := b.start()
	mainDsn := fmt.Sprintf(dnsTemplate, "postgres")

	conn, err := sql.Open("postgres", mainDsn)
	if err != nil {
		log.Fatalf("failed to open postgres: %v", err)
	}
	defer conn.Close()

	dbName := "test_" + uuid.New().String()[:8]
	if _, err := conn.Exec(fmt.Sprintf(`CREATE DATABASE "%s";`, dbName)); err != nil {
		log.Fatalf("failed to create db %s: %v", dbName, err)
	}

	b.databasesMu.Lock()
	b.databases = append(b.databases, dbName)
	b.databasesMu.Unlock()

	return dbName, fmt.Sprintf(dnsTemplate, dbName)
}

func (b *PostgresContainerBuilder) DropDatabase(dbName string) {
	b.databasesMu.Lock()
	defer b.databasesMu.Unlock()
	for i, db := range b.databases {
		if db == dbName {
			b.databases = append(b.databases[:i], b.databases[i+1:]...)
			break
		}
	}
}
