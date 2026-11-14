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

var mysqlContainer *MySQLContainerBuilder
var (
	mysqlTlsContainer   *MySQLContainerBuilder
	mysqlTlsContainerMu sync.Mutex
)

type MySQLContainerBuilder struct {
	certificateAuthority *certificate_builders.CertificateAuthority
	container            testcontainers.Container
	connStrTemplate      string
	containerMu          sync.Mutex
	databasesMu          sync.Mutex
	databases            []string
}

func GetMySQLContainerBuilder() *MySQLContainerBuilder {
	if mysqlContainer == nil {
		mysqlContainer = &MySQLContainerBuilder{}
	}
	return mysqlContainer
}

// GetMySQLTlsContainerBuilder presents a certificate for TlsServerName signed by its own CA.
func GetMySQLTlsContainerBuilder() *MySQLContainerBuilder {
	mysqlTlsContainerMu.Lock()
	defer mysqlTlsContainerMu.Unlock()

	if mysqlTlsContainer == nil {
		mysqlTlsContainer = &MySQLContainerBuilder{certificateAuthority: certificate_builders.NewCertificateAuthority()}
	}
	return mysqlTlsContainer
}

func (b *MySQLContainerBuilder) CaCertificatePem() []byte {
	return b.certificateAuthority.CertificatePem
}

func (b *MySQLContainerBuilder) start() string {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container == nil {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "mysql:9.7",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "root",
				"MYSQL_DATABASE":      "mysql",
			},
			WaitingFor: wait.ForListeningPort("3306/tcp"),
		}
		tlsParam := ""
		if b.certificateAuthority != nil {
			certificatePem, keyPem := b.certificateAuthority.IssueServerCertificate(TlsServerName)
			req.Files = []testcontainers.ContainerFile{
				{Reader: bytes.NewReader(certificatePem), ContainerFilePath: "/certs/server.crt", FileMode: 0o644},
				{Reader: bytes.NewReader(keyPem), ContainerFilePath: "/certs/server.key", FileMode: 0o644},
			}
			req.Entrypoint = []string{"sh", "-c", "install -o mysql -g mysql -m 600 /certs/server.key /tmp/server.key" +
				" && exec docker-entrypoint.sh mysqld --ssl-cert=/certs/server.crt --ssl-key=/tmp/server.key"}
			// skip-verify lets the builder create databases without trusting the CA.
			tlsParam = "&tls=skip-verify"
		} else {
			// The image enables TLS by default; plaintext lets tests prove there is no fallback to it.
			req.Cmd = []string{"--tls-version="}
		}
		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			log.Fatalf("failed to start mysql container: %v", err)
		}
		b.container = container

		host, _ := container.Host(ctx)
		port, _ := container.MappedPort(ctx, "3306")
		b.connStrTemplate = fmt.Sprintf("root:root@tcp(%s:%s)/%%s?parseTime=true%s", host, port.Port(), tlsParam)
		log.Println("MySQL container ready")
	}
	return b.connStrTemplate
}

func (b *MySQLContainerBuilder) stop() {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container != nil {
		_ = b.container.Terminate(context.Background())
		b.container = nil
		log.Println("MySQL container stopped")
	}
	if mysqlContainer == b {
		mysqlContainer = nil
	}
	mysqlTlsContainerMu.Lock()
	if mysqlTlsContainer == b {
		mysqlTlsContainer = nil
	}
	mysqlTlsContainerMu.Unlock()
}

func (b *MySQLContainerBuilder) Create() (string, string) {
	dnsTemplate := b.start()
	mainDsn := fmt.Sprintf(dnsTemplate, "mysql")

	conn, err := sql.Open("mysql", mainDsn)
	if err != nil {
		log.Fatalf("failed to open mysql: %v", err)
	}
	defer conn.Close()

	dbName := "test_" + uuid.New().String()[:8]
	if _, err := conn.Exec(fmt.Sprintf("CREATE DATABASE `%s`;", dbName)); err != nil {
		log.Fatalf("failed to create db %s: %v", dbName, err)
	}

	b.databasesMu.Lock()
	b.databases = append(b.databases, dbName)
	b.databasesMu.Unlock()

	return dbName, fmt.Sprintf(dnsTemplate, dbName)
}

func (b *MySQLContainerBuilder) DropDatabase(dbName string) {
	b.databasesMu.Lock()
	defer b.databasesMu.Unlock()
	for i, db := range b.databases {
		if db == dbName {
			b.databases = append(b.databases[:i], b.databases[i+1:]...)
			break
		}
	}
}
