package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"

	"go.uber.org/zap"
)

type PostgresConfig interface {
	DSN() string
}

type RedisConfig interface {
	URI() string
}

type HTTP struct {
	Address         string        `yaml:"address"`
	Port            int           `yaml:"port"`
	SecurePort      int           `yaml:"secure_port"`
	Timeout         time.Duration `yaml:"timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type PostgresDB struct {
	MaxPoolSize       int32         `yaml:"max_pool_size"`
	MinPoolSize       int32         `yaml:"min_pool_size"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	MaxLifetime       time.Duration `yaml:"max_lifetime"`
}

type Servers struct {
	Env      string     `yaml:"env"`
	Web      HTTP       `yaml:"http_server"`
	Postgres PostgresDB `yaml:"postgres_db"`
}

func NewServersConfig(logger *zap.Logger, confPath string) (*Servers, error) {

	var cfg Servers

	if err := cleanenv.ReadConfig(confPath, &cfg); err != nil {
		logger.Fatal("Error reading config", zap.Error(err))
		return nil, err
	}

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress != "" {
		logger.Info("Overriding server address from ENV", zap.String("address", serverAddress))
		cfg.Web.Address = serverAddress
	} else {
		logger.Info("No SERVER_ADDRESS in ENV, using YAML config")
	}

	return &cfg, nil
}
