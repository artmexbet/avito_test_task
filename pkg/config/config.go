package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Source int

const (
	SourceYAML Source = iota
	SourceEnv
)

type RouterConfig struct {
	Host string `yaml:"host" env:"HOST"`
	Port int    `yaml:"port" env:"PORT"`
}

type PostgresConfig struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT"`
	User     string `yaml:"user" env:"POSTGRES_USER"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"dbname" env:"POSTGRES_DBNAME"`
	SSLMode  string `yaml:"sslmode" env:"POSTGRES_SSLMODE"`
}

func (cfg *PostgresConfig) DSN() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
}

type Config struct {
	Router   RouterConfig   `yaml:"router" env_prefix:"ROUTER_"`
	Postgres PostgresConfig `yaml:"postgres" env_prefix:"POSTGRES_"`
}

func MustParseConfig(source Source, path ...string) Config {
	var cfg Config
	switch source {
	case SourceYAML:
		if len(path) != 1 {
			panic("YAML config source requires a single file path")
		}
		err := readConfigFromYAML(path[0], &cfg)
		if err != nil {
			panic(err)
		}
	case SourceEnv:
		err := cleanenv.ReadEnv(&cfg)
		if err != nil {
			panic(err)
		}
	default:
		panic("unsupported config source")
	}

	return cfg
}

func readConfigFromYAML(path string, cfg *Config) error {
	return cleanenv.ReadConfig(path, cfg)
}
