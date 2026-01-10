package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"

	postgresTestContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	dbInstance *postgresTestContainer.PostgresContainer
	connStr    string
	dbOnce     sync.Once
)

func GetTestDB(t *testing.T) string {
	t.Helper()

	dbOnce.Do(func() {
		container := SetupPostgresContainer(t)

		cs, err := container.ConnectionString(context.Background(), "sslmode=disable")
		if err != nil {
			t.Fatalf("failed to get connection string: %s", err)
		}

		dbInstance = container
		connStr = cs
	})

	return connStr
}

func SetupPostgresContainer(t *testing.T) *postgresTestContainer.PostgresContainer {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	newNetwork, err := network.New(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %s", err)
	}
	netName := newNetwork.Name
	pgContainer, err := postgresTestContainer.Run(ctx, "postgres:15.3-alpine",
		postgresTestContainer.WithDatabase("testdb"),
		postgresTestContainer.WithUsername("postgres"),
		postgresTestContainer.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(1*time.Minute),
		),
		testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
			req.Networks = []string{netName}
			req.NetworkAliases = map[string][]string{
				netName: {"db-host"},
			}
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	runMigrations(t, ctx, "db-host", "5432", netName)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	seedData(t, connStr)

	return pgContainer
}

func runMigrations(t *testing.T, ctx context.Context, dbHost string, dbPort, netName string) {
	t.Helper()

	migrationPath := os.Getenv("MIGRATION_PATH")
	if migrationPath == "" {
		absPath, err := filepath.Abs("../../../../ctc-db")
		if err != nil {
			t.Fatalf("failed to get absolute path: %s", err)
		}
		migrationPath = absPath
	}

	req := testcontainers.ContainerRequest{
		Image: "liquibase/liquibase:4.30",
		Cmd: []string{
			"--search-path=/migration-data",
			"--changelog-file=changelog.xml",
			"--url=jdbc:postgresql://" + dbHost + ":" + dbPort + "/testdb",
			"--username=postgres",
			"--password=postgres",
			"update",
		},
		Mounts: testcontainers.Mounts(
			testcontainers.ContainerMount{
				Source: testcontainers.GenericBindMountSource{
					HostPath: migrationPath,
				},
				Target: "/migration-data",
			},
		),
		Networks:   []string{netName},
		WaitingFor: wait.ForExit(),
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&StdoutLogConsumer{}},
		},
	}

	liquibaseContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Liquibase failed to start: %s", err)
	}

	defer liquibaseContainer.Terminate(context.Background())

	state, err := liquibaseContainer.State(ctx)
	if err != nil {
		t.Fatalf("failed to get liquibase container state: %s", err)
	}

	if state.ExitCode != 0 {
		t.Fatalf("Liquibase migrations failed with exit code %d. Check logs for details.", state.ExitCode)
	}
}

type StdoutLogConsumer struct{}

func (lc *StdoutLogConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}

func seedData(t *testing.T, connStr string) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	seedFilePath := filepath.Join("../../..", "testdata", "db-seed.sql")
	content, err := os.ReadFile(seedFilePath)
	if err != nil {
		t.Fatalf("failed to read seed file: %s", err)
	}

	_, err = db.Exec(string(content))
	if err != nil {
		t.Fatalf("failed to seed database: %s", err)
	}
}
