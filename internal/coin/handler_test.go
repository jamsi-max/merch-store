package coin

import (
	"bytes"
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
)

func setupTestServer(mockDB *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	database := &db.Database{DB: sqlxDB}

	coinHandler := NewCoinHandler(database)

	r.POST("/api/sendCoin", setUserIDMiddleware(1), coinHandler.SendCoin)
	return r
}

func setUserIDMiddleware(userID int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
}

func TestSendCoin(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	server := setupTestServer(mockDB)

	tests := []struct {
		name           string
		body           map[string]interface{}
		setupMock      func()
		expectedStatus int
	}{
		{
			name: "Successful transaction",
			body: map[string]interface{}{"toUser": "receiver", "amount": 100},
			setupMock: func() {

				recipientQuery := mock.ExpectQuery(`SELECT id FROM users WHERE name=\$1`)
				recipientQuery.WithArgs("receiver").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

				balanceQuery := mock.ExpectQuery(`SELECT coins FROM users WHERE id=\$1`)
				balanceQuery.WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(1000))

				mock.ExpectBegin()

				mock.ExpectExec(`UPDATE users SET coins = coins - \$1 WHERE id = \$2`).
					WithArgs(100, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(`UPDATE users SET coins = coins \+ \$1 WHERE id = \$2`).
					WithArgs(100, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(`INSERT INTO transactions \(sender_id, receiver_id, amount\) VALUES \(\$1, \$2, \$3\)`).
					WithArgs(1, 2, 100).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Insufficient funds",
			body: map[string]interface{}{"toUser": "receiver", "amount": 100},
			setupMock: func() {
				mock.ExpectQuery(`SELECT id FROM users WHERE name=\$1`).
					WithArgs("receiver").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

				mock.ExpectQuery(`SELECT coins FROM users WHERE id=\$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(50))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Recipient not found",
			body: map[string]interface{}{"toUser": "unknown", "amount": 100},
			setupMock: func() {
				mock.ExpectQuery(`SELECT id FROM users WHERE name=\$1`).
					WithArgs("unknown").
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid request body",
			body: map[string]interface{}{"toUser": "", "amount": 0},
			setupMock: func() {
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, err := json.Marshal(tt.body)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
