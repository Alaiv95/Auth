package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env            string        `yaml:"env" env-default:"local"`
	Dsn            string        `yaml:"dsn" env-required:"true"`
	MigrationsPath string        `yaml:"migrations_path"`
	GRPC           GRPCConfig    `yaml:"grpc" env-required:"true"`
	TokenTtl       time.Duration `yaml:"token_ttl" env-default:"1h"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-default:"8080"`
	Timeout time.Duration `yaml:"timeout" env-default:"5m"`
}

func MustLoad() *Config {
	path := configPath()

	return MustLoadByPath(path)
}

func MustLoadByPath(path string) *Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		log.Fatal(err)
	}

	return &cfg
}

func configPath() string {
	var path string
	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	return path
}
