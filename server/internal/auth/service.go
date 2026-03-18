package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ffreville/mmo-team-test/server/internal/database"
	"github.com/ffreville/mmo-team-test/server/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type RedisClient interface {
	GetSession(userID string) (string, error)
	SetSession(userID, token string, expiry time.Duration) error
	DeleteSession(userID string) error
	SetRateLimit(key string, limit int, window time.Duration) (bool, error)
}

type AuthService struct {
	db         *database.PostgresDB
	redis      RedisClient
	jwtSecret  []byte
	bcryptCost int
}

func NewAuthService(db *database.PostgresDB, redis *database.RedisClient, jwtSecret string, bcryptCost int) *AuthService {
	return &AuthService{
		db:         db,
		redis:      redis,
		jwtSecret:  []byte(jwtSecret),
		bcryptCost: bcryptCost,
	}
}

func (s *AuthService) Register(username, email, password string) (*models.User, error) {
	ctx := context.Background()

	// Check if user already exists
	var existingUser models.User
	err := s.db.Pool().QueryRow(ctx,
		"SELECT user_id, username, email FROM users WHERE username = $1 OR email = $2",
		username, email).Scan(&existingUser.UserID, &existingUser.Username, &existingUser.Email)

	if err == nil {
		return nil, errors.New("username or email already exists")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	userID := ""
	err = s.db.Pool().QueryRow(ctx,
		"INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING user_id",
		username, email, string(passwordHash)).Scan(&userID)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &models.User{
		UserID:       userID,
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
	}, nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	ctx := context.Background()

	var user models.User
	err := s.db.Pool().QueryRow(ctx,
		"SELECT user_id, username, email, password_hash, is_banned FROM users WHERE username = $1",
		username).Scan(&user.UserID, &user.Username, &user.Email, &user.PasswordHash, &user.IsBanned)

	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if user.IsBanned {
		return "", errors.New("user is banned")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateToken(&user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login
	_, err = s.db.Pool().Exec(ctx,
		"UPDATE users SET last_login = $1 WHERE user_id = $2",
		time.Now(), user.UserID)
	if err != nil {
		log.Printf("Failed to update last login: %v", err)
	}

	// Store session in Redis
	expiry := 24 * time.Hour
	if err := s.redis.SetSession(user.UserID, token, expiry); err != nil {
		log.Printf("Failed to store session: %v", err)
	}

	return token, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Verify session still exists in Redis
		storedToken, err := s.redis.GetSession(claims.UserID)
		if err != nil || storedToken != tokenString {
			return nil, errors.New("session expired or invalid")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *AuthService) Logout(tokenString string) error {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	return s.redis.DeleteSession(claims.UserID)
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID:   user.UserID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "mmo-auth-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
