package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig   `mapstructure:"server"`
	Database     DatabaseConfig `mapstructure:"database"`
	Redis        RedisConfig    `mapstructure:"redis"`
	Auth         AuthConfig     `mapstructure:"auth"`
	ServerLimits ServerLimits   `mapstructure:"server_limits"`
	Movement     MovementConfig `mapstructure:"movement"`
	BuildVersion string         `mapstructure:"-"`
	BuildCommit  string         `mapstructure:"-"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
}

type AuthConfig struct {
	JWTSecret  string        `mapstructure:"jwt_secret"`
	JWTExpiry  time.Duration `mapstructure:"jwt_expiry"`
	BcryptCost int           `mapstructure:"bcrypt_cost"`
}

type ServerLimits struct {
	MaxPlayersPerZone      int `mapstructure:"max_players_per_zone"`
	MaxCharactersPerUser   int `mapstructure:"max_characters_per_user"`
	MaxUsernameLength      int `mapstructure:"max_username_length"`
	MaxCharacterNameLength int `mapstructure:"max_character_name_length"`
}

type MovementConfig struct {
	MaxSpeed                float64       `mapstructure:"max_speed"`
	PositionSyncInterval    time.Duration `mapstructure:"position_sync_interval"`
	AntiCheatSpeedThreshold float64       `mapstructure:"anti_cheat_speed_threshold"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("redis.port", 6379)

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	return &cfg, nil
}
