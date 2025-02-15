package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jamsi-max/merch-store/internal/auth"
	"github.com/jamsi-max/merch-store/internal/coin"
	"github.com/jamsi-max/merch-store/internal/db"
	"github.com/jamsi-max/merch-store/internal/store"
	"github.com/jamsi-max/merch-store/internal/users"
)

func SetupRouter(db *db.Database, jwtSecret string) *gin.Engine {
	r := gin.Default()

	authHandler := auth.NewAuthHandler(db, jwtSecret)
	r.POST("/api/auth", authHandler.Auth)

	coinHandler := coin.NewCoinHandler(db)
	storeHandler := store.NewStoreHandler(db)
	userHandler := users.NewUserHandler(db)

	protected := r.Group("/api")
	protected.Use(auth.AuthMiddleware(jwtSecret))

	protected.POST("/sendCoin", coinHandler.SendCoin)
	protected.GET("/buy/:item", storeHandler.BuyItem)
	protected.GET("/info", userHandler.GetUserInfo)

	return r
}
