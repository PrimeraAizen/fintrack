package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/diyas/fintrack/pkg/logger"
	"github.com/spf13/viper"
)

var ErrInvalidConfig = errors.New("invalid config")

const (
	MigrationDir = "migrations"
	PathToConfig = "./config"
)

type Config struct {
	Http     Http          `mapstructure:"http"`
	PG       PG            `mapstructure:"database"`
	Redis    Redis         `mapstructure:"redis"`
	JWT      JWT           `mapstructure:"jwt"`
	Currency Currency      `mapstructure:"currency"`
	Logger   logger.Config `mapstructure:"logger"`
}

func LoadConfig() (*Config, error) {
	return LoadConfigFromDirectory(PathToConfig)
}

func LoadConfigFromDirectory(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode into struct: %w", err)
	}

	cfg.PG.URL = cfg.PG.connString()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (cfg *Config) Validate() error {
	if cfg.Http.Host == "" {
		return fmt.Errorf("missing http host")
	}
	if cfg.Http.Port == "" {
		return fmt.Errorf("missing http port")
	}
	if cfg.PG.Host == "" || cfg.PG.Port == "" || cfg.PG.Database == "" || cfg.PG.Username == "" {
		return fmt.Errorf("missing database connection settings")
	}
	if cfg.PG.URL == "" {
		return fmt.Errorf("missing database url")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("missing jwt secret")
	}
	if cfg.JWT.AccessTTL == 0 {
		cfg.JWT.AccessTTL = 15 * time.Minute
	}
	if cfg.JWT.RefreshTTL == 0 {
		cfg.JWT.RefreshTTL = 168 * time.Hour
	}
	if cfg.Currency.CacheTTL == 0 {
		cfg.Currency.CacheTTL = time.Hour
	}

	if cfg.Logger.Level == "" {
		cfg.Logger.Level = logger.LevelInfo
	}
	if cfg.Logger.Format == "" {
		cfg.Logger.Format = "json"
	}
	if cfg.Logger.Output == "" {
		cfg.Logger.Output = "stdout"
	}
	if cfg.Logger.Service == "" {
		cfg.Logger.Service = "fintrack"
	}
	if cfg.Logger.Version == "" {
		cfg.Logger.Version = "1.0.0"
	}
	if cfg.Logger.Environment == "" {
		cfg.Logger.Environment = "development"
	}

	return nil
}

func (d *PG) connString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.Username, d.Password, d.Host, d.Port, d.Database, d.SSLMode)
}

type Http struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type PG struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
	MaxConns int    `mapstructure:"max_conns"`
	MinConns int    `mapstructure:"min_conns"`
	URL      string
}

type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWT struct {
	Secret     string        `mapstructure:"secret"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

type Currency struct {
	APIURL   string        `mapstructure:"api_url"`
	CacheTTL time.Duration `mapstructure:"cache_ttl"`
}
