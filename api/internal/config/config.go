package config

import (
	"errors"
	"flag"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/wraper"
	"github.com/ilyakaznacheev/cleanenv"
)

const (
	POSTGRES_PASSWORD = "POSTGRES_PASSWORD"
	REDIS_PASSWORD    = "REDIS_PASSWORD"
)

type Config struct {
	Env      string          `yaml:"env"  env:"ENV"`
	Server   ServerConfig    `yaml:"server"`
	Logger   LoggerConfig    `yaml:"logger"`
	Postgres PostreSQLConfig `yaml:"postgres"`
	Redis    RedisCongig     `yaml:"redis"`
}

type ServerConfig struct {
	Host         string        `yaml:"host" env:"SERVER_HOST"`
	Port         string        `yaml:"port" env:"SERVER_PORT"`
	StdAliasLen  int           `yaml:"std_alias_len" env:"SERVER_STD_ALIAS_LEN" env-default:"6"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"  env:"SERVER_IDLE_TIMEOUT"  env-default:"10m"`
	ReadTimeout  time.Duration `yaml:"read_timeout"  env:"SERVER_READ_TIMEOUT"  env-default:"5m"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT" env-default:"5m"`
	RequesLimit  int           `yaml:"request_limit" env:"SERVER_REQUEST_LIMIT" env-default:"100"`
	WindowLength time.Duration `yaml:"window_length" env:"SERVER_WINDOW_LENGTH" env-default:"1m"`
}

type PostreSQLConfig struct {
	Host     string                   `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Name     string                   `yaml:"name" env:"POSTGRES_DB" env-default:"postgres"`
	User     string                   `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
	Port     string                   `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	SSLMode  string                   `yaml:"ssl_mode" env:"POSTGRES_SSLMODE" env-default:"enable"`
	Password string                   `env:"POSTGRES_PASSWORD"`
	Options  OptionalPostgreSQLConfig `yaml:"options"`
}

type OptionalPostgreSQLConfig struct {
	MaxConns         int           `yaml:"max_conns" env:"MAX_CONNS"`
	MinConns         int           `yaml:"min_conns" env:"MIN_CONNS"`
	MaxConnIdleTime  time.Duration `yaml:"max_conn_idle_time" env:"MAX_CONN_IDLE_TIME"`
	MaxConnLifetime  time.Duration `yaml:"max_conn_life_time" env:"MAX_CONN_LIFE_TIME"`
	CheckHelthPeriod time.Duration `yaml:"check_helth_period" env:"CHECK_HELTH_PERIOD"`
}

type RedisCongig struct {
	DB       int           `yaml:"db"   env:"REDIS_DB" env-default:"0"`
	Host     string        `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
	Port     string        `yaml:"port" env:"REDIS_PORT" env-dafault:"6379"`
	TTL      time.Duration `yaml:"ttl"  env:"REDIS_TTL"`
	Password string        `env:"REDIS_PASSWORD"`
}

type LoggerConfig struct {
	Level string `yaml:"level" env:"LOGGER_LEVEL" env-default:"info"`
}

var ErrEmptyConfigPath = errors.New("config path must not be emprty")

func Load() (*Config, error) {
	const fn = "Load"

	wp := wraper.New(fn)

	cfgPath := featcheCfgPath()

	if err := validatePath(cfgPath); err != nil {
		return nil, wp.WrapMsg(cfgPath, err)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		return nil, wp.WrapMsg(cfgPath, err)
	}

	cfg.Postgres.Password = os.Getenv("POSTGRES_PASS")
	cfg.Redis.Password = os.Getenv("REDIS_PASS")

	return &cfg, nil
}

const (
	INFO  = "info"
	DEBUG = "debug"
	WARN  = "warn"
	ERROR = "error"
)

var levels = map[string]slog.Level{
	INFO:  slog.LevelInfo,
	DEBUG: slog.LevelDebug,
	WARN:  slog.LevelWarn,
	ERROR: slog.LevelError,
}

func (cfg *LoggerConfig) LevelFromString() (slog.Level, error) {
	lvl, ok := levels[strings.TrimSpace(strings.ToLower(cfg.Level))]
	if !ok {
		return slog.Level(-1), wraper.Wrap("LevelFromString", errors.New("unknown logger level"))
	}
	return lvl, nil
}

func featcheCfgPath() string {
	var cfgPath string
	flag.StringVar(&cfgPath, "config_path", "", "path to config file")
	flag.Parse()

	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG_PATH")
	}

	return strings.TrimSpace(cfgPath)
}

func validatePath(cfgPath string) error {
	if cfgPath == "" {
		return ErrEmptyConfigPath
	}
	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		return os.ErrNotExist
	}
	return nil
}
