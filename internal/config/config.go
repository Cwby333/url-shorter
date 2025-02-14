package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env string `yaml:"env" env-required:"true"`

	HTTPServer `yaml:"http-server" env-required:"true"`

	Database `yaml:"database" env-required:"true"`

	Owner `yaml:"owner"`
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

type Owner struct {
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
}

func Load(env string) Config {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	var configPath string
	defaultPassword := os.Getenv("APP_DEFAULT_PASS_DB")

	switch env {
	case "local":
		configPath = os.Getenv("APP_CONFIG_PATH_LOCAL")
	case "dev":
		configPath = os.Getenv("APP_CONFIG_PATH_DEV")
	}

	if configPath == "" {
		log.Println("cannot find config path, will use default params")

		cfg := Config{
			Env: "prod",

			HTTPServer: HTTPServer{
				Address:         "localhost:8080",
				ReadTimeout:     time.Duration(time.Second * 5),
				WriteTimeout:    time.Duration(time.Second * 5),
				IdleTimeout:     time.Duration(time.Second * 30),
				ShutdownTimeout: time.Duration(time.Second * 30),
			},

			Database: Database{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: defaultPassword,
				DBname:   "urls",
				MaxConn:  20,
				MinConn:  5,
				SslMode:  "disable",
			},
		}

		return cfg
	}

	cfg := Config{}

	err = cleanenv.ReadConfig(configPath, &cfg)

	if err != nil {
		log.Fatal(err)
	}

	cfg.SslMode = "disable"

	return cfg
}
