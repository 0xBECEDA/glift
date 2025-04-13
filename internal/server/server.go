package server

import (
	"app/internal/blockchain"
	"app/internal/database/models"
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math/big"
	"net/http"
	"strings"

	"app/internal/database"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger

	e  *echo.Echo
	bc blockchain.Client
	db database.Database
}

func NewServer(bc blockchain.Client, db database.Database, logger *zap.Logger) *Server {
	e := echo.New()
	s := &Server{
		e:      e,
		bc:     bc,
		db:     db,
		logger: logger,
	}

	e.POST("/transaction/send", s.submitFILTransaction)
	e.GET("/transactions/", s.getTransactions)

	e.GET("/balance/:address", s.getBalance)
	return s
}

func (s *Server) Start(addr string) {
	go s.e.Start(addr)
	return
}

func (s *Server) Stop(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}

func (s *Server) getBalance(c echo.Context) error {
	address := c.Param("address")
	ctx := c.Request().Context()

	if !isValidAddress(address) {
		s.logger.Warn("Invalid address format", zap.String("address", address))
		return ErrInvalidAddress
	}

	balances, err := s.bc.GetBalances(ctx, common.HexToAddress(strings.TrimPrefix(address, "0x")))
	if err != nil {
		s.logger.Error("Failed to get iFIL balance", zap.String("address", address), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "failed to get iFIL balance"))
	}

	s.logger.Info("Balance retrieved", zap.String("address", address))
	return c.JSON(http.StatusOK, &BalanceResponse{
		FIL:  balances.GetFIL().String(),
		IFIL: balances.GetIFIL().String(),
	})
}

func (s *Server) submitFILTransaction(c echo.Context) error {
	ctx := c.Request().Context()

	var req SubmitTransactionRequest
	if err := c.Bind(&req); err != nil {
		s.logger.Warn("Invalid request body", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(req.PrivateKeyHex, "0x"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "invalid private key"))
	}

	if !isValidAddress(req.Receiver) {
		s.logger.Warn("Invalid receiver address", zap.String("receiver", req.Receiver))
		return ErrInvalidReceiverAddress
	}

	amount := big.NewInt(0)
	amount.SetString(req.Amount, 10)
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return ErrInvalidTxAmount
	}
	receiver := common.HexToAddress(strings.TrimPrefix(req.Receiver, "0x"))
	sender := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	txReceipt, err := s.bc.SubmitFILTransaction(ctx, privateKey, receiver, amount)
	if err != nil {
		s.logger.Error("Failed to submit transaction", zap.String("sender", sender), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, errors.Wrap(err, "failed to submit transaction"))
	}

	txHash := txReceipt.Hash().String()
	senderToLower := strings.ToLower(sender)
	receiverToLower := strings.ToLower(req.Receiver)

	tx := &models.Transaction{
		Hash:     txHash,
		Sender:   senderToLower,
		Receiver: receiverToLower,
		Amount:   decimal.NewFromBigInt(amount, 0),
		Status:   models.StatusPending,
	}

	if err := s.db.SaveTransaction(tx); err != nil {
		s.logger.Error("Failed to save transaction", zap.String("hash", txHash), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "failed to save transaction"))
	}

	s.logger.Info("Transaction submitted", zap.String("hash", txHash), zap.String("sender", sender), zap.String("receiver", req.Receiver))
	return c.JSON(http.StatusCreated, SubmitTransactionResponse{Hash: txHash})
}

func (s *Server) getTransactions(c echo.Context) error {
	sender := c.QueryParam("sender")
	receiver := c.QueryParam("receiver")

	if sender != "" && !isValidAddress(sender) {
		s.logger.Warn("Invalid sender address", zap.String("sender", sender))
		return ErrInvalidSenderAddress
	}

	if receiver != "" && !isValidAddress(receiver) {
		s.logger.Warn("Invalid receiver address", zap.String("receiver", receiver))
		return ErrInvalidReceiverAddress
	}

	txs, err := s.db.GetTransactions(c.Request().Context(), strings.ToLower(sender), strings.ToLower(receiver), 0)
	if err != nil {
		s.logger.Error("failed to retrieve transactions", zap.Error(err), zap.String("sender", sender), zap.String("receiver", receiver))
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "failed to retrieve transactions"))
	}

	s.logger.Info("Transactions retrieved", zap.Int("count", len(txs)), zap.String("sender", sender), zap.String("receiver", receiver))
	return c.JSON(http.StatusOK, txs)
}

func isValidAddress(addr string) bool {
	return common.IsHexAddress(strings.TrimSpace(addr))
}
