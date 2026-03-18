package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type MockRow struct {
	ScanFunc func(dest ...interface{}) error
}

func (m *MockRow) Scan(dest ...interface{}) error {
	return m.ScanFunc(dest...)
}

type MockCommandResult struct{}

func (m *MockCommandResult) RowsAffected() int64 { return 0 }

type MockPostgresPool struct {
	QueryRowFunc func(ctx context.Context, sql string, args ...interface{}) pgx.Row
	ExecFunc     func(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
}

func (m *MockPostgresPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return m.QueryRowFunc(ctx, sql, args...)
}

func (m *MockPostgresPool) Exec(ctx context.Context, sql string, args ...interface{}) (interface{}, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, sql, args...)
	}
	return &MockCommandResult{}, nil
}

func (m *MockPostgresPool) Ping(ctx context.Context) error {
	return nil
}

func (m *MockPostgresPool) Close() {}

type MockRedisClient struct {
	SetSessionFunc    func(userID string, token string, expiry time.Duration) error
	GetSessionFunc    func(userID string) (string, error)
	DeleteSessionFunc func(userID string) error
	SetRateLimitFunc  func(key string, limit int, window time.Duration) (bool, error)
	ClientFunc        func() *redis.Client
}

func (m *MockRedisClient) SetSession(userID string, token string, expiry time.Duration) error {
	return m.SetSessionFunc(userID, token, expiry)
}

func (m *MockRedisClient) GetSession(userID string) (string, error) {
	return m.GetSessionFunc(userID)
}

func (m *MockRedisClient) DeleteSession(userID string) error {
	return m.DeleteSessionFunc(userID)
}

func (m *MockRedisClient) SetRateLimit(key string, limit int, window time.Duration) (bool, error) {
	if m.SetRateLimitFunc != nil {
		return m.SetRateLimitFunc(key, limit, window)
	}
	return true, nil
}

func (m *MockRedisClient) Client() *redis.Client {
	if m.ClientFunc != nil {
		return m.ClientFunc()
	}
	return nil
}

func (m *MockRedisClient) Close() {}

var _ pgx.Row = &MockRow{}
var _ interface{} = &MockCommandResult{}
