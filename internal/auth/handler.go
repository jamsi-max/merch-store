package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jamsi-max/merch-store/internal/db"
)

type AuthHandler struct {
	db        *db.Database
	jwtSecret string
}

func NewAuthHandler(db *db.Database, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Auth(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Invalid request"})
		return
	}

	var user User

	err := h.db.DB.Get(&user, "SELECT * FROM users WHERE name=$1", req.Username)

	if err != nil {
		hashedPassword, err := HashPassword(req.Password)
		if err != nil {
			log.Printf("[ERR] failed to hash password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "Failed to hash password"})
			return
		}

		err = h.db.DB.QueryRow(`
			INSERT INTO users (name, pass, coins) VALUES ($1, $2, 1000) RETURNING id`,
			req.Username, hashedPassword).Scan(&user.ID)

		if err != nil {
			log.Printf("[ERR] failed to create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "Failed to create user"})
			return
		}

		user.Name = req.Username
		user.Coins = 1000
	} else {

		if !CheckPassword(user.Pass, req.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"errors": "Invalid username or password"})
			return
		}
	}

	token, err := GenerateToken(user.ID, user.Name, h.jwtSecret)
	if err != nil {
		log.Printf("[ERR] failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
