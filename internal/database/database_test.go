package database

import (
	"app/internal/database/models"
	"context"
	"database/sql"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestDriver_WithDockerCompose(t *testing.T) {
	ctx := context.Background()
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "cannot get current file location")

	composeFile := "docker-compose.yml"
	composePath := filepath.Join(filepath.Dir(currentFile), "../../", composeFile)
	absComposePath, err := filepath.Abs(composePath)
	require.NoError(t, err, "failed to get absolute path to docker-compose.yml")

	composeProject, err := compose.NewDockerCompose(absComposePath)
	require.NoError(t, err)

	err = composeProject.Up(ctx, compose.RunServices("postgres"))
	require.NoError(t, err)

	t.Cleanup(func() { composeProject.Down(ctx) })

	time.Sleep(5 * time.Second)

	dsn := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"

	logger := zap.NewNop()
	driver, err := NewDriver(logger, dsn)
	require.NoError(t, err)
	require.NotNil(t, driver)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'transactions'
		);
	`
	err = db.QueryRow(query).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists, "transactions table should exist after migration")

	// get any tx from db
	txs, err := driver.GetTransactions(ctx, "", "", 0)
	require.NoError(t, err)
	require.Len(t, txs, 0)

	const (
		txHash   = "0x4033cf2e690e8f6078ab9e665be681c1a9d7c79b6a75d0668da5336fabf43b46"
		sender   = "0xa83114A443dA1CecEFC50368531cACE9F37fCCcb"
		receiver = "0xFFEEDDCcBbAA0000000000000000000000000000"
	)

	// save tx
	tx := &models.Transaction{
		Hash:     txHash,
		Sender:   sender,
		Receiver: receiver,
		Amount:   decimal.NewFromFloat(0.999999999),
		Status:   models.StatusPending,
	}

	err = driver.SaveTransaction(tx)
	require.NoError(t, err)

	txs, err = driver.GetTransactions(ctx, sender, "", 0)
	require.NoError(t, err)
	require.Len(t, txs, 1)
	actualTx := txs[0]

	require.NotEmpty(t, actualTx.ID)
	require.Equal(t, tx.Hash, actualTx.Hash)
	require.Equal(t, tx.Sender, actualTx.Sender)
	require.True(t, !tx.Timestamp.IsZero())
	require.Equal(t, tx.Amount.String(), actualTx.Amount.String())
	require.Equal(t, tx.Status, actualTx.Status)

	// save same tx, but with updated status
	tx.Status = models.StatusFailed
	err = driver.SaveTransaction(tx)
	require.NoError(t, err)

	txs, err = driver.GetTransactions(ctx, sender, "", 0)
	require.NoError(t, err)
	require.Len(t, txs, 1)

	updatedTx := txs[0]
	require.Equal(t, tx.Hash, updatedTx.Hash)
	require.Equal(t, tx.Sender, updatedTx.Sender)
	require.Equal(t, tx.Amount.String(), updatedTx.Amount.String())
	require.True(t, !tx.Timestamp.IsZero())
	require.Equal(t, models.StatusFailed, updatedTx.Status)

	// save same tx, but with again updated status -> fail, because tx might be updated only if status is 'pending'
	tx.Status = models.StatusConfirmed
	err = driver.SaveTransaction(tx)
	require.ErrorIs(t, err, ErrTxExists)
}
