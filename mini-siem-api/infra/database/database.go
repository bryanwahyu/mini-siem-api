package database

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database connection parameters.
type Config struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Open initialises a GORM connection for the configured database.
func Open(cfg Config) (*gorm.DB, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))

	gormCfg := &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true}

	var (
		db  *gorm.DB
		err error
	)

	switch driver {
	case "postgres", "postgresql":
		db, err = gorm.Open(postgres.Open(cfg.DSN), gormCfg)
	case "sqlite", "sqlite3", "":
		dsn := cfg.DSN
		if dsn == "" {
			dsn = "mini_siem.db"
		}
		db, err = gorm.Open(sqlite.Open(dsn), gormCfg)
	default:
		return nil, fmt.Errorf("unsupported db driver %s", cfg.Driver)
	}

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	db.Logger = db.Logger.LogMode(logger.Warn)

	return db, nil
}
