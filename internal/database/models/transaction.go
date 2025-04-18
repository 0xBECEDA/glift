package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type TransactionStatus string

const (
	StatusPending   TransactionStatus = "pending"
	StatusConfirmed TransactionStatus = "confirmed"
	StatusFailed    TransactionStatus = "failed"
)

type Transaction struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	Hash      string
	Sender    string
	Receiver  string
	Amount    decimal.Decimal `gorm:"type:decimal(30,18)"`
	Timestamp time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
	Status    TransactionStatus
}
