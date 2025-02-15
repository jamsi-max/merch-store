package store

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jamsi-max/merch-store/internal/db"
)

const (
	green = "\033[32m"
	red   = "\033[31m"
	reset = "\033[0m"
)

type StoreHandler struct {
	db           *db.Database
	MerchCatalog map[string]int
}

func NewStoreHandler(db *db.Database) *StoreHandler {
	handler := &StoreHandler{
		db:           db,
		MerchCatalog: make(map[string]int),
	}

	if err := handler.loadMerchCatalog(); err != nil {
		log.Fatalf(red+"[ERR]"+reset+"couldn't load the merch catalog: %v", err)
	}

	log.Println(green + "[INFO] merch catalog upload successfully" + reset)
	return handler
}

func (h *StoreHandler) loadMerchCatalog() error {
	rows, err := h.db.DB.Query("SELECT name, price FROM merch")
	if err != nil {
		return err
	}
	defer rows.Close()

	var name string
	var price int
	for rows.Next() {
		if err := rows.Scan(&name, &price); err != nil {
			return err
		}
		h.MerchCatalog[name] = price
	}

	return rows.Err()
}

func (h *StoreHandler) BuyItem(c *gin.Context) {
	item := c.Param("item")

	price, ok := h.MerchCatalog[item]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Item not found"})
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "Unauthorized"})
		return
	}

	var userCoins int
	err := h.db.DB.Get(&userCoins, "SELECT coins FROM users WHERE id=$1", userID)
	if err != nil || price > userCoins {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds"})
		return
	}

	tx, err := h.db.DB.Begin()
	if err != nil {
		log.Printf("[ERR] transaction store failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction store failed"})
		return
	}

	_, err = tx.Exec("UPDATE users SET coins = coins - $1 WHERE id = $2", price, userID)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("[ERR] failed to rollback transaction: %v", rbErr)
		}
		log.Printf("[ERR] failed to update balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	_, err = tx.Exec(`
		INSERT INTO user_merch (user_id, item, quantity) 
		VALUES ($1, $2, 1) 
		ON CONFLICT (user_id, item) 
		DO UPDATE SET quantity = user_merch.quantity + 1`, userID, item)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("[ERR] failed to rollback transaction: %v", rbErr)
		}
		log.Printf("[ERR] failed to update user merch: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user merch"})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[ERR] failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.Status(http.StatusOK)
}
