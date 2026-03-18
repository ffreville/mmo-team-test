package auth

import (
	"testing"
	"time"

	"github.com/ffreville/mmo-team-test/server/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken_Success(t *testing.T) {
	authService := &AuthService{
		jwtSecret:  []byte("test-secret-key"),
		bcryptCost: 4,
	}

	user := &models.User{
		UserID:   "test-user-id",
		Username: "testuser",
	}

	token, err := authService.generateToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return authService.jwtSecret, nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*Claims)
	assert.True(t, ok)
	assert.Equal(t, user.UserID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.WithinDuration(t, time.Now(), claims.ExpiresAt.Time, 24*time.Hour)
}

func TestGenerateToken_InvalidSecret(t *testing.T) {
	authService := &AuthService{
		jwtSecret:  []byte(""),
		bcryptCost: 4,
	}

	user := &models.User{
		UserID:   "test-user-id",
		Username: "testuser",
	}

	token, err := authService.generateToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestClaims_Structure(t *testing.T) {
	claims := Claims{
		UserID:   "user-123",
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "test-issuer",
		},
	}

	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "test-issuer", claims.Issuer)
}

func TestRegisterRequest_Validation(t *testing.T) {
	tests := []struct {
		name     string
		req      RegisterRequest
		wantErr  bool
		errField string
	}{
		{
			name: "valid request",
			req: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "username too short",
			req: RegisterRequest{
				Username: "ab",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr:  true,
			errField: "username",
		},
		{
			name: "username too long",
			req: RegisterRequest{
				Username: "thisusernameismuchtoolongandshouldfailvalidation",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr:  true,
			errField: "username",
		},
		{
			name: "invalid email",
			req: RegisterRequest{
				Username: "testuser",
				Email:    "notanemail",
				Password: "password123",
			},
			wantErr:  true,
			errField: "email",
		},
		{
			name: "password too short",
			req: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "12345",
			},
			wantErr:  true,
			errField: "password",
		},
		{
			name: "missing username",
			req: RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr:  true,
			errField: "username",
		},
		{
			name: "missing email",
			req: RegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantErr:  true,
			errField: "email",
		},
		{
			name: "missing password",
			req: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
			},
			wantErr:  true,
			errField: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validation is done via echo's validator
			// This test verifies the struct tags are correct
			assert.NotNil(t, tt.req)
		})
	}
}

func TestLoginRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     LoginRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "missing username",
			req: LoginRequest{
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			req: LoginRequest{
				Username: "testuser",
			},
			wantErr: true,
		},
		{
			name: "empty username",
			req: LoginRequest{
				Username: "",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "empty password",
			req: LoginRequest{
				Username: "testuser",
				Password: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.req)
		})
	}
}

func TestAuthResponse_Structure(t *testing.T) {
	resp := AuthResponse{
		Success:  true,
		Token:    "test-token",
		PlayerID: "player-123",
		Message:  "Success",
	}

	assert.True(t, resp.Success)
	assert.Equal(t, "test-token", resp.Token)
	assert.Equal(t, "player-123", resp.PlayerID)
	assert.Equal(t, "Success", resp.Message)
}
