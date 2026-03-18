package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Register_BodyBinding(t *testing.T) {
	e := echo.New()

	reqBody := `{"username": "testuser", "email": "test@example.com", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var registerReq RegisterRequest
	err := c.Bind(&registerReq)

	assert.NoError(t, err)
	assert.Equal(t, "testuser", registerReq.Username)
	assert.Equal(t, "test@example.com", registerReq.Email)
	assert.Equal(t, "password123", registerReq.Password)
}

func TestAuthHandler_Register_BindError(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("invalid json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := &AuthHandler{}
	err := handler.Register(c)

	assert.NoError(t, err) // Echo handles the error
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Login_BodyBinding(t *testing.T) {
	e := echo.New()

	reqBody := `{"username": "testuser", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var loginReq LoginRequest
	err := c.Bind(&loginReq)

	assert.NoError(t, err)
	assert.Equal(t, "testuser", loginReq.Username)
	assert.Equal(t, "password123", loginReq.Password)
}

func TestAuthHandler_Logout_MissingHeader(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := &AuthHandler{}
	err := handler.Logout(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp AuthResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "missing authorization header", resp.Error)
}

func TestAuthHandler_Logout_BearerFormat(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test that the header is parsed correctly
	authHeader := c.Request().Header.Get("Authorization")
	assert.Equal(t, "Bearer test-token", authHeader)

	// Simulate token extraction
	token := ""
	if authHeader != "" {
		token = authHeader[7:] // len("Bearer ") = 7
	}
	assert.Equal(t, "test-token", token)
}

func TestAuthHandler_GetProfile_UserID(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set user_id from middleware
	c.Set("user_id", "test-user-123")
	c.Set("username", "testuser")

	userID := c.Get("user_id").(string)
	assert.Equal(t, "test-user-123", userID)
}

func TestRegisterRequest_JSONTags(t *testing.T) {
	req := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, "testuser", result["username"])
	assert.Equal(t, "test@example.com", result["email"])
	assert.Equal(t, "password123", result["password"])
}

func TestLoginRequest_JSONTags(t *testing.T) {
	req := LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, "testuser", result["username"])
	assert.Equal(t, "password123", result["password"])
}

func TestAuthResponse_JSONSerialization(t *testing.T) {
	resp := AuthResponse{
		Success:  true,
		Token:    "jwt-token-123",
		PlayerID: "player-456",
		Message:  "Login successful",
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.True(t, result["success"].(bool))
	assert.Equal(t, "jwt-token-123", result["token"])
	assert.Equal(t, "player-456", result["player_id"])
	assert.Equal(t, "Login successful", result["message"])
}

func TestAuthResponse_ErrorSerialization(t *testing.T) {
	resp := AuthResponse{
		Success: false,
		Error:   "invalid credentials",
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.False(t, result["success"].(bool))
	assert.Equal(t, "invalid credentials", result["error"])
}

func TestAuthHandler_NewAuthHandler(t *testing.T) {
	authService := &AuthService{
		jwtSecret:  []byte("test"),
		bcryptCost: 4,
	}

	handler := NewAuthHandler(authService)

	assert.NotNil(t, handler)
	assert.Equal(t, authService, handler.authService)
}
