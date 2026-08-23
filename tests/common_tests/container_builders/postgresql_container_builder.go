package container_builders

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var postgresContainer *PostgresContainerBuilder

type PostgresContainerBuilder struct {
	container       testcontainers.Container
	connStrTemplate string
	containerMu     sync.Mutex
	databasesMu     sync.Mutex
	databases       []string
}

func GetPostgresContainerBuilder() *PostgresContainerBuilder {
	if postgresContainer == nil {
		postgresContainer = &PostgresContainerBuilder{}
	}
	return postgresContainer
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
			"host=%s port=%s dbname=%s user=postgres password=postgres sslmode=disable",
			host, port.Port(), "%s",
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
	postgresContainer = nil
}

// Create returns (dbName, connectionString) for a freshly-created scratch DB
// inside the shared container. Each test gets its own DB so they can run in
// parallel without stepping on each other's schema.
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
