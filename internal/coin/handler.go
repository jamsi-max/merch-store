package coin

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jamsi-max/merch-store/internal/db"
)

type CoinHandler struct {
	db *db.Database
}

func NewCoinHandler(db *db.Database) *CoinHandler {
	return &CoinHandler{db: db}
}

func (h *CoinHandler) SendCoin(c *gin.Context) {
	var req struct {
		ToUser string `json:"toUser" binding:"required"`
		Amount int    `json:"amount" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Invalid request"})
		return
	}

	fromUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "Unauthorized"})
		return
	}

	var toUserID int
	err := h.db.DB.Get(&toUserID, "SELECT id FROM users WHERE name=$1", req.ToUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Recipient not found"})
		return
	}

	var senderCoins int
	err = h.db.DB.Get(&senderCoins, "SELECT coins FROM users WHERE id=$1", fromUserID)
	if err != nil || senderCoins < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Insufficient funds"})
		return
	}

	tx, err := h.db.DB.Begin()
	if err != nil {
		log.Printf("[ERR] transaction coin failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Transaction coin failed"})
		return
	}

	_, err = tx.Exec("UPDATE users SET coins = coins - $1 WHERE id = $2", req.Amount, fromUserID)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("[ERR] failed to rollback transaction: %v", rbErr)
		}
		log.Printf("[ERR] failed to update sender balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Failed to update sender balance"})
		return
	}

	_, err = tx.Exec("UPDATE users SET coins = coins + $1 WHERE id = $2", req.Amount, toUserID)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("[ERR] failed to rollback transaction: %v", rbErr)
		}
		log.Printf("[ERR] failed to update recipient balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Failed to update recipient balance"})
		return
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (sender_id, receiver_id, amount) VALUES ($1, $2, $3)`,
		fromUserID, toUserID, req.Amount)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("[ERR] failed to rollback transaction: %v", rbErr)
		}
		log.Printf("[ERR] failed to record transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Failed to record transaction"})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[ERR] failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Failed to commit transaction"})
		return
	}

	c.Status(http.StatusOK)
}
