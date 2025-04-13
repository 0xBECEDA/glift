package server

import (
	"app/internal/blockchain"
	blockchainmock "app/internal/blockchain/mock"
	dbmock "app/internal/database/mock"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetBalance_Success(t *testing.T) {
	const (
		address             = "0xa512eb36e162bfb0e9f55b56bc2a070ca0d3ecd7"
		expectedFILBalance  = 100.0
		expectedIFILBalance = 50.0
	)

	logger := zap.NewNop()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := blockchainmock.NewMockClient(ctrl)
	mockDatabase := dbmock.NewMockDatabase(ctrl)

	srv := NewServer(mockClient, mockDatabase, logger)
	go srv.Start(":8080")
	defer srv.Stop(context.Background())

	expectedBalances := blockchain.NewWalletBalance(big.NewFloat(expectedFILBalance), big.NewFloat(expectedIFILBalance))
	mockClient.EXPECT().
		GetBalances(gomock.Any(), common.HexToAddress(strings.TrimPrefix(address, "0x"))).
		Return(expectedBalances, nil)

	testServer := httptest.NewServer(srv.e)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/balance/" + address)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	response := &BalanceResponse{}
	err = json.Unmarshal(body, response)
	require.NoError(t, err)

	require.Equal(t, fmt.Sprintf("%v", expectedFILBalance), response.FIL)
	require.Equal(t, fmt.Sprintf("%v", expectedIFILBalance), response.IFIL)
}
