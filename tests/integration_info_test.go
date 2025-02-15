package users_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/jamsi-max/merch-store/internal/db"
	"github.com/jamsi-max/merch-store/internal/router"
)

var jwtToken = generateTestJWT("testsecret", 1)

func setupTestDB(t *testing.T) *db.Database {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		t.Fatalf("Failed to start test database container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	})

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	connStr := fmt.Sprintf("host=localhost port=%s user=testuser password=testpass dbname=testdb sslmode=disable", port.Port())

	dbConn, err := db.NewDatabase(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := applyMigrations(dbConn); err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	seedTestData(dbConn)

	return dbConn
}

func seedTestData(db *db.Database) {
	db.DB.MustExec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		pass TEXT NOT NULL,
		coins INT NOT NULL
	)`)

	db.DB.MustExec(`CREATE TABLE IF NOT EXISTS merch (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		price INT NOT NULL
	)`)

	db.DB.MustExec(`CREATE TABLE IF NOT EXISTS user_merch (
		user_id INT REFERENCES users(id),
		item TEXT NOT NULL,
		quantity INT NOT NULL
	)`)

	db.DB.MustExec(`CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		sender_id INT REFERENCES users(id),
		receiver_id INT REFERENCES users(id),
		amount INT NOT NULL
	)`)

	db.DB.MustExec("INSERT INTO users (id, name, pass, coins) VALUES (1, 'testuser', 'password', 500)")
	db.DB.MustExec("INSERT INTO users (id, name, pass, coins) VALUES (2, 'sender', 'password', 300)")
	db.DB.MustExec("INSERT INTO user_merch (user_id, item, quantity) VALUES (1, 'sword', 2), (1, 'shield', 1)")
	db.DB.MustExec("INSERT INTO transactions (sender_id, receiver_id, amount) VALUES (2, 1, 100)")
}

func TestGetUserInfo(t *testing.T) {
	db := setupTestDB(t)
	r := router.SetupRouter(db, "testsecret")

	req, _ := http.NewRequest("GET", "/api/info", nil)
	req.Header.Set("Authorization", jwtToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 500, int(response["coins"].(float64)))
}

func applyMigrations(db *db.Database) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		pass TEXT NOT NULL,
		coins INT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS user_merch (
		user_id INT REFERENCES users(id),
		item TEXT NOT NULL,
		quantity INT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		sender_id INT REFERENCES users(id),
		receiver_id INT REFERENCES users(id),
		amount INT NOT NULL
	);
	`
	_, err := db.DB.Exec(schema)
	return err
}

func generateTestJWT(secret string, userID int) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString([]byte(secret))
	return "Bearer " + signedToken
}
