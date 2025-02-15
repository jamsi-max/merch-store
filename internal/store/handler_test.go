package store_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jamsi-max/merch-store/internal/auth"
	"github.com/jamsi-max/merch-store/internal/db"
	"github.com/jamsi-max/merch-store/internal/router"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testJWTSecret = "testsecret"

func setupTestDB(t *testing.T) *db.Database {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "admin",
			"POSTGRES_PASSWORD": "secrets",
			"POSTGRES_DB":       "merch_store",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithStartupTimeout(30 * time.Second),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start test database container: %v", err)
	}

	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	})

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}
	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	connStr := "user=admin password=secrets dbname=merch_store sslmode=disable host=" + host + " port=" + port.Port()
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	time.Sleep(5 * time.Second)

	_, err = dbConn.Exec(`
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            pass TEXT NOT NULL,
            coins INT NOT NULL
        );

        CREATE TABLE merch (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            price INT NOT NULL
        );

        CREATE TABLE user_merch (
            user_id INT REFERENCES users(id),
            item TEXT NOT NULL,
            quantity INT NOT NULL,
            PRIMARY KEY (user_id, item)
        );

        CREATE TABLE transactions (
            id SERIAL PRIMARY KEY,
            user_id INT REFERENCES users(id),
            item TEXT NOT NULL,
            quantity INT NOT NULL,
            total_price INT NOT NULL
        );
    `)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	_, err = dbConn.Exec(`
        INSERT INTO users (name, pass, coins) VALUES ('testuser', 'password', 1000);
        INSERT INTO merch (name, price) VALUES ('t-shirt', 500);
    `)
	if err != nil {
		t.Fatalf("Failed to prepare test database: %v", err)
	}

	sqlxDB := sqlx.NewDb(dbConn, "postgres")
	return &db.Database{DB: sqlxDB}
}

func TestBuyItemE2E(t *testing.T) {
	db := setupTestDB(t)
	r := router.SetupRouter(db, testJWTSecret)

	token, err := auth.GenerateToken(1, "testuser", testJWTSecret)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/buy/t-shirt", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var newBalance int
	err = db.DB.QueryRow("SELECT coins FROM users WHERE id=1").Scan(&newBalance)
	assert.NoError(t, err)
	assert.Equal(t, 500, newBalance)

	var quantity int
	err = db.DB.QueryRow("SELECT quantity FROM user_merch WHERE user_id=1 AND item='t-shirt'").Scan(&quantity)
	assert.NoError(t, err)
	assert.Equal(t, 1, quantity)
}
