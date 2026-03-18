package auth

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, handler *AuthHandler, authMiddleware *AuthMiddleware) {
	auth := e.Group("/auth")

	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)
	auth.POST("/logout", handler.Logout)

	// Protected routes
	protected := e.Group("")
	protected.Use(authMiddleware.Middleware())
	protected.GET("/profile", handler.GetProfile)
}
