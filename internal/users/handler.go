package users

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jamsi-max/merch-store/internal/db"
	"github.com/jamsi-max/merch-store/utils"
)

type UserHandler struct {
	db   *db.Database
	Info InfoResponse
}

func NewUserHandler(db *db.Database) *UserHandler {
	info := InfoResponse{
		Inventory: []InventoryItem{},
		CoinHistory: CoinHistory{
			Received: []CoinTransaction{},
			Sent:     []CoinTransaction{},
		},
	}

	return &UserHandler{db: db, Info: info}
}

func (u *UserHandler) GetUserInfo(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := u.db.DB.Get(&u.Info.Coins, "SELECT coins FROM users WHERE id=$1", userID)
	if err != nil {
		log.Printf("[ERR] failed to get balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Failed to get balance"})
		return
	}

	err = u.db.DB.Select(&u.Info.Inventory, "SELECT item, quantity FROM user_merch WHERE user_id=$1", userID)
	if err != nil {
		log.Printf("[ERR] failed to get user_merch: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Failed to get user merch"})
		return
	}

	err = u.db.DB.Select(&u.Info.CoinHistory.Received, `
		SELECT u.name AS sender_id, t.amount 
		FROM transactions t 
		JOIN users u ON t.sender_id = u.id 
		WHERE t.receiver_id = $1`, userID)
	if err != nil {
		log.Printf("[ERR] failed to get received transactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Failed to get received transactions"})
		return
	}

	err = u.db.DB.Select(&u.Info.CoinHistory.Sent, `
		SELECT u.name AS receiver_id, t.amount 
		FROM transactions t 
		JOIN users u ON t.receiver_id = u.id 
		WHERE t.sender_id = $1`, userID)
	if err != nil {
		log.Printf("[ERR] failed to get sent transactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Failed to get sent transactions"})
		return
	}

	c.JSON(http.StatusOK, u.Info)
}
