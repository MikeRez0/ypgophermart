package testdb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

const (
	testDBName       = "test"
	testUserName     = "test"
	testUserPassword = "test"
)

type TestDBInstance struct {
	dockerPool *dockertest.Pool
	dbResource *dockertest.Resource
	DSN        string
}

func NewTestDBInstance() (*TestDBInstance, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a pool: %w", err)
	}

	pg, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "15.3",
			Name:       "database-test",
			Env: []string{
				"POSTGRES_USER=postgres",
				"POSTGRES_PASSWORD=postgres",
			},
			ExposedPorts: []string{"5432"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				docker.Port("5432"): {docker.PortBinding{HostPort: "50000"}},
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run the postgres container: %w", err)
	}
	defer func(err *error) {
		if *err != nil {
			if err1 := pool.Purge(pg); err1 != nil {
				log.Printf("failed to purge the postgres container: %v", err1)
			}
		}
	}(&err)

	hostPort := pg.GetHostPort("5432/tcp")

	pool.MaxWait = 10 * time.Second
	var connSU *pgx.Conn
	err = pool.Retry(func() error {
		connSU, err = pgx.Connect(context.Background(),
			fmt.Sprintf("postgres://postgres:postgres@%s/postgres?sslmode=disable", hostPort))
		if err != nil {
			return fmt.Errorf("failed to get a super user connection: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	err = createTestDB(connSU)
	if err != nil {
		return nil, fmt.Errorf("failed to create db: %w", err)
	}

	t := TestDBInstance{
		dockerPool: pool,
		dbResource: pg,
		DSN: fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
			testUserName, testUserPassword, hostPort, testDBName),
	}
	return &t, nil
}

func (t *TestDBInstance) Down() {
	if err1 := t.dockerPool.Purge(t.dbResource); err1 != nil {
		log.Printf("failed to purge the postgres container: %v", err1)
	}
}

func createTestDB(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(),
		fmt.Sprintf(
			`CREATE USER %s PASSWORD '%s'`,
			testUserName,
			testUserPassword,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create a test user: %w", err)
	}

	_, err = conn.Exec(context.Background(),
		fmt.Sprintf(`
			CREATE DATABASE %s
				OWNER '%s'
				ENCODING 'UTF8'
				LC_COLLATE = 'en_US.utf8'
				LC_CTYPE = 'en_US.utf8'
			`, testDBName, testUserName,
		),
	)

	if err != nil {
		return fmt.Errorf("failed to create a test DB: %w", err)
	}

	return nil
}
