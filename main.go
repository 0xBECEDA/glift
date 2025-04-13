package main

import (
	"app/internal/blockchain"
	"app/internal/database"
	"app/internal/server"
	"context"
	"os"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	chain := os.Getenv("CHAIN_ID")
	chainId, err := blockchain.StringToChainId(chain)
	if err != nil {
		logger.Fatal(err.Error())
	}

	dbDSN := os.Getenv("DATABASE_DSN")
	if dbDSN == "" {
		logger.Fatal("Missing DATABASE_DSN")
	}

	serverListenAddr := os.Getenv("SERVER_LISTEN_ADDR")
	if serverListenAddr == "" {
		logger.Fatal("Missing SERVER_LISTEN_ADDR")
	}

	client, err := blockchain.NewClient(logger, chainId, blockchain.DefaultTestnetRPCConfig)
	if err != nil {
		logger.Fatal("Failed to initialize blockchain client", zap.Error(err))
	}

	dbDriver, err := database.NewDriver(logger, dbDSN)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	srv := server.NewServer(client, dbDriver, logger)

	logger.Info("Starting server", zap.String("address", serverListenAddr))
	srv.Start(serverListenAddr)

	ctx := context.Background()
	<-ctx.Done()
	srv.Stop(ctx)
}
