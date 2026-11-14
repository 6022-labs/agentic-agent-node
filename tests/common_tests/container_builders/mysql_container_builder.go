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

var (
	mysqlContainer   *MySQLContainerBuilder
	mysqlContainerMu sync.Mutex
)

type MySQLContainerBuilder struct {
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

func GetMySQLContainerBuilder() *MySQLContainerBuilder {
	mysqlContainerMu.Lock()
	defer mysqlContainerMu.Unlock()

	if mysqlContainer == nil {
		mysqlContainer = &MySQLContainerBuilder{}
	}
	return mysqlContainer
}

func (b *MySQLContainerBuilder) start() string {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container == nil {
		req := b.request()
		// The image enables TLS by default; plaintext lets tests prove there is no fallback to it.
		req.Cmd = []string{"--tls-version="}

		b.container, b.connStrTemplate = b.run(req)
		log.Println("MySQL container ready")
	}
	return b.connStrTemplate
}

func (b *MySQLContainerBuilder) startTls() (string, []byte) {
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
			" && exec docker-entrypoint.sh mysqld --ssl-cert=/certs/server.crt --ssl-key=/tmp/server.key"}

		var connStrTemplate string
		b.tlsContainer, connStrTemplate = b.run(req)
		// skip-verify lets the builder create databases without trusting the CA.
		b.tlsConnStrTemplate = connStrTemplate + "&tls=skip-verify"
		b.caCertificatePem = certificateAuthority.CertificatePem
		log.Println("MySQL TLS container ready")
	}
	return b.tlsConnStrTemplate, b.caCertificatePem
}

func (b *MySQLContainerBuilder) request() testcontainers.ContainerRequest {
	return testcontainers.ContainerRequest{
		Image:        "mysql:9.7",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "root",
			"MYSQL_DATABASE":      "mysql",
		},
		WaitingFor: wait.ForListeningPort("3306/tcp"),
	}
}

func (b *MySQLContainerBuilder) run(req testcontainers.ContainerRequest) (testcontainers.Container, string) {
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start mysql container: %v", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "3306")
	return container, fmt.Sprintf("root:root@tcp(%s:%s)/%%s?parseTime=true", host, port.Port())
}

func (b *MySQLContainerBuilder) stop() {
	b.containerMu.Lock()
	if b.container != nil {
		_ = b.container.Terminate(context.Background())
		b.container = nil
		log.Println("MySQL container stopped")
	}
	b.containerMu.Unlock()

	b.tlsContainerMu.Lock()
	if b.tlsContainer != nil {
		_ = b.tlsContainer.Terminate(context.Background())
		b.tlsContainer = nil
		log.Println("MySQL TLS container stopped")
	}
	b.tlsContainerMu.Unlock()

	mysqlContainerMu.Lock()
	mysqlContainer = nil
	mysqlContainerMu.Unlock()
}

func (b *MySQLContainerBuilder) Create() (string, string) {
	dnsTemplate := b.start()
	dbName := b.createDatabase(dnsTemplate)
	return dbName, fmt.Sprintf(dnsTemplate, dbName)
}

func (b *MySQLContainerBuilder) CreateTls() (string, string, []byte) {
	dnsTemplate, caCertificatePem := b.startTls()
	dbName := b.createDatabase(dnsTemplate)
	return dbName, fmt.Sprintf(dnsTemplate, dbName), caCertificatePem
}

func (b *MySQLContainerBuilder) createDatabase(dnsTemplate string) string {
	conn, err := sql.Open("mysql", fmt.Sprintf(dnsTemplate, "mysql"))
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

	return dbName
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
