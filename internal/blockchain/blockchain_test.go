package blockchain

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestWallet(t *testing.T) {
	const addr = "0xa986b79597588E4519FE0ABEfCBa37A343c44046"
	logger := zap.NewNop()

	cl, err := NewClient(logger, testNet, DefaultTestnetRPCConfig)
	assert.NoError(t, err)

	balances, err := cl.GetBalances(context.Background(), common.HexToAddress(addr))
	assert.NoError(t, err)
	fmt.Printf("fil: %v, ifil: %v", balances.GetFIL().String(), balances.GetIFIL().String())
}
