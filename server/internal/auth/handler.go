package auth

import (
	"net/http"
	"strings"

	"github.com/ffreville/mmo-team-test/server/pkg/models"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *AuthService
}

func NewAuthHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Success  bool   `json:"success"`
	Token    string `json:"token,omitempty"`
	PlayerID string `json:"player_id,omitempty"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Error:   "invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	user, err := h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusConflict, AuthResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, AuthResponse{
		Success:  true,
		Message:  "User registered successfully",
		PlayerID: user.UserID,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Error:   "invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Get user_id from token claims for response
	claims, _ := h.authService.ValidateToken(token)

	return c.JSON(http.StatusOK, AuthResponse{
		Success:  true,
		Token:    token,
		PlayerID: claims.UserID,
		Message:  "Login successful",
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Error:   "missing authorization header",
		})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if err := h.authService.Logout(token); err != nil {
		return c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

func (h *AuthHandler) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var user models.User
	err := h.authService.db.Pool().QueryRow(c.Request().Context(),
		"SELECT user_id, username, email, created_at, last_login, is_banned FROM users WHERE user_id = $1",
		userID).Scan(&user.UserID, &user.Username, &user.Email, &user.CreatedAt, &user.LastLogin, &user.IsBanned)

	if err != nil {
		return c.JSON(http.StatusNotFound, AuthResponse{
			Success: false,
			Error:   "user not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":    true,
		"user_id":    user.UserID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"last_login": user.LastLogin,
		"is_banned":  user.IsBanned,
	})
}
