package config

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const configTestPath = "D:/url-shortener/api/internal/config/cfg_test.yaml"

func Test_validatePath(t *testing.T) {
	type args struct {
		cfgPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base-case",
			args: args{
				cfgPath: configTestPath,
			},
			wantErr: false,
		},

		{
			name: "empty config path",
			args: args{
				cfgPath: "",
			},
			wantErr: true,
		},

		{
			name: "unknown path",
			args: args{
				cfgPath: "api/x/unknow/x/cfg.yaml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.args.cfgPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_LevelFromString(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		want    slog.Level
		wantErr bool
	}{
		{
			name:    "level info",
			cfg:     &Config{Logger: LoggerConfig{Level: "info"}},
			want:    slog.LevelInfo,
			wantErr: false,
		},
		{
			name:    "level error",
			cfg:     &Config{Logger: LoggerConfig{Level: "error"}},
			want:    slog.LevelError,
			wantErr: false,
		},
		{
			name:    "level debug",
			cfg:     &Config{Logger: LoggerConfig{Level: "debug"}},
			want:    slog.LevelDebug,
			wantErr: false,
		},
		{
			name:    "level warn",
			cfg:     &Config{Logger: LoggerConfig{Level: "warn"}},
			want:    slog.LevelWarn,
			wantErr: false,
		},
		{
			name:    "level in up regist",
			cfg:     &Config{Logger: LoggerConfig{Level: "INFO"}},
			want:    slog.LevelInfo,
			wantErr: false,
		},

		{
			name:    "invalid level",
			cfg:     &Config{Logger: LoggerConfig{Level: "INVALID LEVEL"}},
			want:    slog.Level(-1),
			wantErr: true,
		},

		{
			name:    "empty config",
			cfg:     &Config{},
			want:    slog.Level(-1),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cfg.Logger.LevelFromString()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		before  func()
		want    *Config
		wantErr bool
	}{
		{
			name: "valid data",
			before: func() {
				os.Setenv("CONFIG_PATH", configTestPath)
			},
			// from cfg_test.yaml
			want: &Config{
				Env: "local",

				Server: ServerConfig{
					Host:         "0.0.0.0",
					Port:         "5000",
					IdleTimeout:  time.Minute,
					ReadTimeout:  time.Minute,
					WriteTimeout: time.Minute,
					StdAliasLen:  5,
					RequesLimit:  120,
					WindowLength: time.Minute,
				},

				Logger: LoggerConfig{
					Level: "info",
				},

				Postgres: PostreSQLConfig{
					Host:    "0.0.0.0",
					Port:    "5432",
					Name:    "postgres",
					User:    "postgres",
					SSLMode: "disable",
					Options: OptionalPostgreSQLConfig{
						MaxConns:         10,
						MinConns:         1,
						MaxConnIdleTime:  time.Minute,
						MaxConnLifetime:  time.Minute,
						CheckHelthPeriod: time.Minute,
					},
				},

				Redis: RedisCongig{
					Host: "0.0.0.0",
					Port: "6379",
					TTL:  time.Minute,
					DB:   0,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.before != nil {
				tt.before()
			}

			got, err := Load()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			got.Postgres.Password = ""
			got.Redis.Password = ""

			assert.Equal(t, *tt.want, *got)
		})
	}
}
