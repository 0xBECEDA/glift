package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/glifio/go-pools/sdk"
	glifio "github.com/glifio/go-pools/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math/big"
)

type ChainId int64

const (
	testNet ChainId = 314159
)

func StringToChainId(s string) (ChainId, error) {
	switch s {
	case "testnet":
		return testNet, nil
	}

	return 0, errors.New("unknown chain")
}

type Client interface {
	GetBalances(ctx context.Context, address common.Address) (*WalletBalance, error)
	SubmitIFILTransaction(ctx context.Context, signer *ecdsa.PrivateKey, receiver common.Address, amount *big.Int) (*types.Transaction, error)
	SubmitFILTransaction(ctx context.Context, signer *ecdsa.PrivateKey, receiver common.Address, amount *big.Int) (*types.Transaction, error)
}

type client struct {
	logger *zap.Logger
	sdk    glifio.PoolsSDK
}

func NewClient(logger *zap.Logger, id ChainId, connectInfo glifio.Extern) (Client, error) {
	initedSdk, err := sdk.New(context.Background(), big.NewInt(int64(id)), connectInfo)
	if err != nil {
		return nil, err
	}
	return &client{
		logger: logger,
		sdk:    initedSdk,
	}, nil
}

var DefaultTestnetRPCConfig = glifio.Extern{
	AdoAddr:       "https://ado-testnet.glif.link/rpc/v1",
	LotusDialAddr: "https://api.calibration.node.glif.io/rpc/v1",
	EventsURL:     "https://events-calibration.glif.link",
}

type WalletBalance struct {
	fil  *big.Float
	ifil *big.Float
}

func NewWalletBalance(fil, ifil *big.Float) *WalletBalance {
	return &WalletBalance{fil, ifil}
}

func (wb *WalletBalance) GetFIL() *big.Float {
	return wb.fil
}

func (wb *WalletBalance) GetIFIL() *big.Float {
	return wb.ifil
}

func (c *client) GetBalances(ctx context.Context, address common.Address) (*WalletBalance, error) {
	ethClient, err := c.sdk.Extern().ConnectEthClient()
	if err != nil {
		return nil, err
	}
	defer ethClient.Close()

	balance, err := ethClient.BalanceAt(ctx, address, nil)
	if err != nil {
		c.logger.Error("failed to get FIL balance", zap.Error(err), zap.String("address", address.Hex()))
		return nil, err
	}

	// 1 attoFIL is equal to 10^-18 * FIL
	fils := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))

	queries := c.sdk.Query()
	ifils, err := queries.IFILBalanceOf(ctx, address)
	if err != nil {
		return nil, err
	}

	return &WalletBalance{
		fil:  fils,
		ifil: ifils,
	}, nil
}

func (c *client) SubmitIFILTransaction(ctx context.Context, signer *ecdsa.PrivateKey, receiver common.Address, amount *big.Int) (*types.Transaction, error) {
	ethClient, err := c.sdk.Extern().ConnectEthClient()
	if err != nil {
		return nil, err
	}
	defer ethClient.Close()

	chainID, err := ethClient.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	sender := crypto.PubkeyToAddress(signer.PublicKey)
	acts := c.sdk.Act()

	signerFn := func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != sender {
			return nil, fmt.Errorf("signer address mismatch: expected %s, got %s", sender.Hex(), address.Hex())
		}
		signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), signer)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %w", err)
		}
		return signedTx, nil
	}

	auth := &bind.TransactOpts{
		From:    sender,
		Signer:  signerFn,
		Context: ctx,
	}

	tx, err := acts.IFILTransfer(ctx, auth, receiver, amount)
	if err != nil {
		c.logger.Error("failed to submit transaction", zap.Error(err), zap.String("sender_address", sender.Hex()), zap.String("receiver_address", receiver.Hex()), zap.String("amount", amount.String()))
		return nil, err
	}
	return tx, nil
}

func (c *client) SubmitFILTransaction(ctx context.Context, signer *ecdsa.PrivateKey, receiver common.Address, amount *big.Int) (*types.Transaction, error) {
	ethClient, err := c.sdk.Extern().ConnectEthClient()
	if err != nil {
		return nil, err
	}
	defer ethClient.Close()

	chainID, err := ethClient.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	sender := crypto.PubkeyToAddress(signer.PublicKey)

	nonce, err := ethClient.PendingNonceAt(ctx, sender)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasTipCap, err := ethClient.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas tip cap: %w", err)
	}

	gasFeeCap := new(big.Int).Add(gasTipCap, big.NewInt(2e9))
	gasLimit, err := ethClient.EstimateGas(ctx, ethereum.CallMsg{
		From:  sender,
		To:    &receiver,
		Value: amount,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Gas:       uint64(float64(gasLimit) * 1.5), // to prevent 'out of gas' error
		To:        &receiver,
		Value:     amount,
	})

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), signer)
	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	err = ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send tx: %w", err)
	}

	return signedTx, nil
}
