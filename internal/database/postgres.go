package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/vidwadeseram/go-auth-template/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for attempt := 1; attempt <= 15; attempt++ {
		db, err = gorm.Open(postgres.Open(cfg.DatabaseDSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err == nil {
			sqlDB, sqlErr := db.DB()
			if sqlErr != nil {
				return nil, fmt.Errorf("get sql db: %w", sqlErr)
			}
			if pingErr := sqlDB.Ping(); pingErr == nil {
				sqlDB.SetMaxIdleConns(5)
				sqlDB.SetMaxOpenConns(20)
				sqlDB.SetConnMaxLifetime(time.Hour)
				return db, nil
			}
		}

		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("connect postgres: %w", err)
}

func RunMigrations(cfg *config.Config) error {
	m, err := migrate.New("file://migrations", cfg.MigrationDSN())
	if err != nil {
		return fmt.Errorf("create migration runner: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
