package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/ffreville/mmo-team-test/server/internal/database"
	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	authService *AuthService
}

func NewAuthMiddleware(authService *AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing authorization header",
				})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid authorization format",
				})
			}

			token := parts[1]
			claims, err := m.authService.ValidateToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid or expired token",
				})
			}

			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)

			return next(c)
		}
	}
}

type RateLimiter struct {
	redis  *database.RedisClient
	limit  int
	window int
}

func NewRateLimiter(redis *database.RedisClient, limit int, window int) *RateLimiter {
	return &RateLimiter{redis: redis, limit: limit, window: window}
}

func (r *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			key := "rate_limit:" + ip

			allowed, err := r.redis.SetRateLimit(key, r.limit, time.Duration(r.window)*time.Second)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "rate limit check failed",
				})
			}

			if !allowed {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "too many requests",
				})
			}

			return next(c)
		}
	}
}
