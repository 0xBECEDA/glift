package server

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

var (
	ErrInvalidAddress         = echo.NewHTTPError(http.StatusBadRequest, "invalid address")
	ErrInvalidSenderAddress   = echo.NewHTTPError(http.StatusBadRequest, "invalid sender address")
	ErrInvalidReceiverAddress = echo.NewHTTPError(http.StatusBadRequest, "invalid receiver address")
	ErrInvalidTxAmount        = echo.NewHTTPError(http.StatusBadRequest, "invalid tx amount: must be positive value")
)
