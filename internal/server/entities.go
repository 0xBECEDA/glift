package server

type BalanceResponse struct {
	FIL  string `json:"fil"`
	IFIL string `json:"ifil"`
}

type SubmitTransactionRequest struct {
	PrivateKeyHex string `json:"private_key_hex"`
	Receiver      string `json:"receiver"`
	Amount        string `json:"amount"`
}

type SubmitTransactionResponse struct {
	Hash string `json:"hash"`
}
