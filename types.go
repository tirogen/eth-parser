package ethparser

type EthRequest struct {
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	ID      uint32   `json:"id"`
	JsonRPC string   `json:"jsonrpc"`
}

type BlockNumberResponse struct {
	Result  string `json:"result"`
	ID      uint32 `json:"id"`
	JsonRPC string `json:"jsonrpc"`
}

type TransactionByHashResponse struct {
	ID      uint32       `json:"id"`
	JsonRPC string       `json:"jsonrpc"`
	Result  *Transaction `json:"result"`
}

type Transaction struct {
	BlockHash            *string `json:"blockHash,omitempty"`
	BlockNumber          *string `json:"blockNumber,omitempty"`
	From                 string  `json:"from,omitempty"`
	Gas                  string  `json:"gas"`
	GasPrice             *string `json:"gasPrice,omitempty"`
	MaxFeePerGas         *string `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas *string `json:"maxPriorityFeePerGas,omitempty"`
	Hash                 string  `json:"hash"`
	Input                string  `json:"input"`
	Nonce                string  `json:"nonce"`
	To                   *string `json:"to,omitempty"` // nil for contract creation
	Value                *string `json:"value,omitempty"`
}

type SubscribeTransactionResult struct {
	JsonRPC string  `json:"jsonrpc"`
	ID      *uint32 `json:"id,omitempty"`
	Result  *string `json:"result,omitempty"`
	Method  *string `json:"method,omitempty"`
	Params  *struct {
		Result       string `json:"result"`
		Subscription string `json:"subscription"`
	} `json:"params,omitempty"`
}
