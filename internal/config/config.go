package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

var (
	ErrCantFindConfig = errors.New("cant find config")
)

type Config struct {
	Env         string `yaml:"env" env-required:"true"`
	HTTPServer  `yaml:"http-server" env-required:"true"`
	Database    `yaml:"database" env-required:"true"`
	JWT         `yaml:"jwt" env-required:"true"`
	Redis       `yaml:"redis" env-required:"true"`
	RateLimiter `yaml:"rate-limiter" env-required:"true"`
}

type HTTPServer struct {
	Address         string        `yaml:"address"`
	ReadTimeout     time.Duration `yaml:"read-timeout"`
	WriteTimeout    time.Duration `yaml:"write-timeout"`
	IdleTimeout     time.Duration `yaml:"idle-timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown-timeout"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     uint16 `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBname   string `yaml:"db-name"`
	SslMode  string `yaml:"ssl-mode"`
	MaxConn  uint16 `yaml:"max-conn"`
	MinConn  uint16 `yaml:"min-conn"`
}

type Redis struct {
	Host     string        `yaml:"host"`
	Username string        `yaml:"username"`
	Port     int           `yaml:"port"`
	Password string        `yaml:"password"`
	DB       int           `yaml:"db"`
	URLTTL   time.Duration `yaml:"url-ttl"`
}

type JWT struct {
	Issuer     string `yaml:"issuer" evn-required:"true"`
	SecretKey  string `yaml:"secret-key" env-required:"true"`
	JWTAccess  `yaml:"jwt-access" env-required:"true"`
	JWTRefresh `yaml:"jwt-refresh" env-required:"true"`
}

type JWTAccess struct {
	ExpiredTime string `yaml:"jwt-access-expired" env-required:"true"`
}

type JWTRefresh struct {
	ExpiredTime string `yaml:"jwt-refresh-expired" env-required:"true"`
}

type RateLimiter struct {
	Limit int           `yaml:"limit" env-required:"true"`
	TTL   time.Duration `yaml:"ttl" env-required:"true"`
}

func Load(env string) (Config, error) {
	const op = "internal/config/Load"

	err := godotenv.Load()

	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", op, err)
	}

	var configPath string

	switch env {
	case "local":
		configPath = os.Getenv("APP_CONFIG_PATH_LOCAL")
	case "dev":
		configPath = os.Getenv("APP_CONFIG_PATH_DEV")
	}

	if configPath == "" {
		return Config{}, fmt.Errorf("%s: %w", op, ErrCantFindConfig)
	}

	cfg := Config{}

	err = cleanenv.ReadConfig(configPath, &cfg)

	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", op, err)
	}

	cfg.JWT.SecretKey = os.Getenv("APP_JWT_SECRET_KEY")

	return cfg, nil
}
