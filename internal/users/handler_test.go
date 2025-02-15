package users

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jamsi-max/merch-store/internal/db"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(_ *testing.T, mockDB *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	database := &db.Database{DB: sqlxDB}

	userHandler := NewUserHandler(database)
	r.GET("/api/info", func(c *gin.Context) {
		c.Set("userID", 1) // Устанавливаем userID в контекст
		userHandler.GetUserInfo(c)
	})

	return r
}

func TestGetUserInfo(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	mock.ExpectQuery("SELECT coins FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(1000))

	mock.ExpectQuery("SELECT item, quantity FROM user_merch WHERE user_id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"item", "quantity"}))

	mock.ExpectQuery("SELECT u.name AS sender_id, t.amount FROM transactions t JOIN users u ON t.sender_id = u.id WHERE t.receiver_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"sender_id", "amount"}))

	mock.ExpectQuery("SELECT u.name AS receiver_id, t.amount FROM transactions t JOIN users u ON t.receiver_id = u.id WHERE t.sender_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"receiver_id", "amount"}))

	server := setupTestServer(t, mockDB)

	req, err := http.NewRequest(http.MethodGet, "/api/info", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response InfoResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 1000, response.Coins)
	assert.Empty(t, response.Inventory)
	assert.Empty(t, response.CoinHistory.Received)
	assert.Empty(t, response.CoinHistory.Sent)
}

func TestGetUserInfo_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	userHandler := NewUserHandler(&db.Database{DB: sqlxDB})
	r.GET("/api/info", userHandler.GetUserInfo)

	req, err := http.NewRequest(http.MethodGet, "/api/info", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)

	assert.Equal(t, "Unauthorized", res["errors"])
}

func TestGetUserInfo_MethodNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	userHandler := NewUserHandler(&db.Database{DB: sqlxDB})
	r.GET("/api/info", userHandler.GetUserInfo)

	req, err := http.NewRequest(http.MethodPost, "/api/info", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	req, err = http.NewRequest(http.MethodPut, "/api/info", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	req, err = http.NewRequest(http.MethodTrace, "/api/info", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserInfo_DatabaseError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	mock.ExpectQuery("SELECT coins FROM users WHERE id=\\$1").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	server := setupTestServer(t, mockDB)

	req, err := http.NewRequest(http.MethodGet, "/api/info", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var res map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)

	assert.Equal(t, "Failed to get balance", res["errors"])
}
