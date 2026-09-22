package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AppEnv     string `env:"APP_ENV" envDefault:"development"`
	ServerPort string `env:"SERVER_PORT" envDefault:"8080"`
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"DB_PORT" envDefault:"5432"`
	DBName     string `env:"DB_NAME" envDefault:"app"`
	DBUser     string `env:"DB_USER" envDefault:"app"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"app_pwd"`
	RedisHost  string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort  string `env:"REDIS_PORT" envDefault:"6379"`
	JWTSecret  string `env:"JWT_SECRET" envDefault:"development_secret"`
}

func Load() (Config, error) { var cfg Config; return cfg, env.Parse(&cfg) }
func (c Config) DatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai", c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}
func (c Config) RedisAddress() string          { return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort) }
func (c Config) ServerAddress() string         { return ":" + c.ServerPort }
func (c Config) StartupTimeout() time.Duration { return time.Second * 15 }
