package database

import (
	"app/internal/database/models"
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"path/filepath"
	"runtime"
)

const limit = 100

var ErrTxExists = errors.New("transaction already exists and it is not pending")

type driver struct {
	logger *zap.Logger
	db     *gorm.DB
}

type Database interface {
	SaveTransaction(tx *models.Transaction) error
	GetTransactions(ctx context.Context, sender, receiver string, offset int) ([]models.Transaction, error)
}

func NewDriver(logger *zap.Logger, dsn string) (Database, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying DB: %w", err)
	}

	driverDB, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate driver: %w", err)
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("cannot get current file path")
	}

	migrationsPath := filepath.Join(filepath.Dir(currentFile), "migrations")
	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", migrationsPath), "postgres", driverDB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migration: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	return &driver{logger: logger, db: db}, nil
}

func (d *driver) SaveTransaction(tx *models.Transaction) error {
	return d.db.Transaction(func(dbTx *gorm.DB) error {
		result := dbTx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "hash"}},
			DoUpdates: clause.AssignmentColumns([]string{"status"}),
			Where: clause.Where{Exprs: []clause.Expression{
				clause.Expr{SQL: `"transactions"."status" = ?`, Vars: []interface{}{"pending"}},
			}},
		}).Omit("id").Create(tx)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrTxExists
		}
		return nil
	})
}

func (d *driver) GetTransactions(ctx context.Context, sender, receiver string, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction

	db := d.db.WithContext(ctx)
	switch {
	case sender != "" && receiver != "":
		db = db.Where("sender = ? AND receiver = ?", sender, receiver)
	case sender != "":
		db = db.Where("sender = ?", sender)
	case receiver != "":
		db = db.Where("receiver = ?", receiver)
	}
	if err := db.Find(&transactions).Limit(limit).Offset(offset).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
