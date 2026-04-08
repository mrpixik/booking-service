package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string          `env:"ENVIRONMENT"`
	Postgres PostgresStorage `envPrefix:"POSTGRES_"`
	HTTP     HTTPServer      `envPrefix:"HTTP_"`
	JWT      JWT             `envPrefix:"JWT_"`
}

type PostgresStorage struct {
	User            string        `env:"USER,required"`
	Password        string        `env:"PASSWORD,required"`
	Host            string        `env:"HOST,required"`
	Port            string        `env:"PORT,required"`
	DbName          string        `env:"DB_NAME,required"`
	MaxConns        int32         `env:"MAX_CONNS,required"`
	MinConns        int32         `env:"MIN_CONNS,required"`
	MaxConnLifeTime time.Duration `env:"MAX_CONN_LIFE_TIME,required"`
}

type HTTPServer struct {
	Port            string        `env:"PORT" envDefault:"8080"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
	RateLimiter     RateLimiter   `envPrefix:"RATE_LIMITER_"`
}

type RateLimiter struct {
	MaxRPC    int `env:"MAX_RPC" envDefault:"5"`
	RPCRefill int `env:"RPC_REFILL" envDefault:"5"`
}

type JWT struct {
	Secret     string        `env:"SECRET,required"`
	Expiration time.Duration `env:"EXPIRATION" env-default:"24h"`
}

func MustLoad() Config {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("unable to load config from .env: \n%s", err.Error())
	}

	var config Config

	err := env.Parse(&config)
	if err != nil {
		log.Fatalf("unable to load config: \n%s", err.Error())
	}

	return config
}
