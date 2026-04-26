package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	DBHost                 string `mapstructure:"DB_HOST"`
	DBPort                 int    `mapstructure:"DB_PORT"`
	DBUser                 string `mapstructure:"DB_USER"`
	DBPassword             string `mapstructure:"DB_PASSWORD"`
	DBName                 string `mapstructure:"DB_NAME"`
	DBSSLMode              string `mapstructure:"DB_SSLMODE"`
	JWTSecret              string `mapstructure:"JWT_SECRET"`
	JWTAccessExpireMinutes int    `mapstructure:"JWT_ACCESS_EXPIRE_MINUTES"`
	JWTRefreshExpireDays   int    `mapstructure:"JWT_REFRESH_EXPIRE_DAYS"`
	SMTPHost               string `mapstructure:"SMTP_HOST"`
	SMTPPort               int    `mapstructure:"SMTP_PORT"`
	AppPort                int    `mapstructure:"APP_PORT"`
	MailFrom               string `mapstructure:"MAIL_FROM"`
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env.example")
	_ = godotenv.Overload(".env")

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	defaults := map[string]any{
		"DB_HOST":                   "db",
		"DB_PORT":                   5432,
		"DB_USER":                   "postgres",
		"DB_PASSWORD":               "postgres",
		"DB_NAME":                   "authdb",
		"DB_SSLMODE":                "disable",
		"JWT_SECRET":                "change-me-in-production",
		"JWT_ACCESS_EXPIRE_MINUTES": 15,
		"JWT_REFRESH_EXPIRE_DAYS":   7,
		"SMTP_HOST":                 "mailhog",
		"SMTP_PORT":                 1025,
		"APP_PORT":                  8000,
		"MAIL_FROM":                 "no-reply@go-auth-template.local",
	}

	for key, value := range defaults {
		v.SetDefault(key, value)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC", c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

func (c *Config) MigrationDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func (c *Config) MailAddress() string {
	return fmt.Sprintf("%s:%d", c.SMTPHost, c.SMTPPort)
}
