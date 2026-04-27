package container_builders

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var mysqlContainer *MySQLContainerBuilder

type MySQLContainerBuilder struct {
	container       testcontainers.Container
	connStrTemplate string
	containerMu     sync.Mutex
	databasesMu     sync.Mutex
	databases       []string
}

func GetMySQLContainerBuilder() *MySQLContainerBuilder {
	if mysqlContainer == nil {
		mysqlContainer = &MySQLContainerBuilder{}
	}
	return mysqlContainer
}

func (b *MySQLContainerBuilder) start() string {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container == nil {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "mysql:8.0",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "root",
				"MYSQL_DATABASE":      "mysql",
			},
			WaitingFor: wait.ForListeningPort("3306/tcp"),
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
		b.connStrTemplate = fmt.Sprintf("root:root@tcp(%s:%s)/%%s?parseTime=true", host, port.Port())
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
	mysqlContainer = nil
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
