package router

import (
	"ewallet/internal/config"
	"ewallet/internal/handlers"
	"ewallet/internal/middleware"
	"ewallet/internal/repository/gorm"
	"ewallet/internal/service"
	"ewallet/internal/validator"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"time"
)

func Register(e *echo.Echo, database *gorm.DB, cfg config.Config) {
	e.Validator = validator.New()
	e.HideBanner = true
	e.HTTPErrorHandler = handlers.HTTPErrorHandler

	userRepo := gormrepo.NewUserRepository(database)
	walletRepo := gormrepo.NewWalletRepository(database)
	transactionRepo := gormrepo.NewTransactionRepository(database)
	revokedTokenRepo := gormrepo.NewRevokedTokenRepository(database)

	authService := service.NewAuthService(userRepo, revokedTokenRepo, cfg.JWTSecret)
	userService := service.NewUserService(userRepo)
	walletService := service.NewWalletService(walletRepo)
	historyCache := service.NewInMemoryHistoryCache(30 * time.Second)
	transactionService := service.NewTransactionService(userRepo, transactionRepo, historyCache)

	authHandler := handlers.NewAuthHandler(authService, cfg)
	userHandler := handlers.NewUserHandler(userService)
	walletHandler := handlers.NewWalletHandler(walletService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	api := e.Group("/api")
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/logout", authHandler.Logout)

	protected := api.Group("")
	protected.Use(middleware.JWT(cfg.JWTSecret, authService))
	protected.GET("/users/profile", userHandler.Profile)
	protected.POST("/wallets/topup", walletHandler.TopUp)
	protected.GET("/wallets/balance", walletHandler.Balance)
	protected.POST("/transactions/transfer", transactionHandler.Transfer)
	protected.GET("/transactions/history", transactionHandler.History)
}
