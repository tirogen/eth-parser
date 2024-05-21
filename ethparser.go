package ethparser

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"
)

type Parser interface {
	// last parsed block
	GetCurrentBlock() int

	// add address to observer
	Subscribe(address string) bool

	// list of inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}

type Transaction struct {
	BlockNumber *string `json:"blockNumber,omitempty"`
	BlockHash   *string `json:"blockHash,omitempty"`
	From        *string `json:"from,omitempty"`
	Nonce       string  `json:"nonce"`
	GasPrice    *string `json:"gasPrice,omitempty"`
	GasTipCap   *string `json:"gasTipCap,omitempty"`
	GasFeeCap   *string `json:"gasFeeCap,omitempty"`
	Gas         uint64  `json:"gas"`
	To          *string `json:"to,omitempty"`
}

type ethParser struct {
	httpProvider string
	wssProvider  string
	client       *http.Client
	idCounter    atomic.Uint32
	transactions map[string][]Transaction
}

type Config struct {
	HTTPProvider        string
	WSSProvider         string
	MaxIdleConns        int
	IdleConnTimeout     time.Duration
	TLSHandshakeTimeout time.Duration
	Timeout             time.Duration
}

func New(ctx context.Context, c *Config) (Parser, error) {
	transport := &http.Transport{
		MaxIdleConns:        c.MaxIdleConns,
		IdleConnTimeout:     c.IdleConnTimeout,
		TLSHandshakeTimeout: c.TLSHandshakeTimeout,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   c.Timeout,
	}
	parser := &ethParser{
		httpProvider: c.HTTPProvider,
		wssProvider:  c.WSSProvider,
		client:       client,
		idCounter:    atomic.Uint32{},
		transactions: map[string][]Transaction{},
	}
	if err := parser.backgroundProcess(ctx); err != nil {
		return nil, err
	}
	return parser, nil
}

type BlockNumberRequest struct {
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      uint32 `json:"id"`
	JsonRPC string `json:"jsonrpc"`
}

type BlockNumberResponse struct {
	Result  string `json:"result"`
	ID      uint32 `json:"id"`
	JsonRPC string `json:"jsonrpc"`
}

func (p *ethParser) GetCurrentBlock() int {
	payload := BlockNumberRequest{
		Method:  "eth_blockNumber",
		Params:  []any{},
		ID:      p.nextID(),
		JsonRPC: "2.0",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		panic(ErrFailedToParseJSON)
	}

	req, err := http.NewRequest("POST", p.httpProvider, bytes.NewBuffer(jsonPayload))
	if err != nil {
		panic(ErrFailedToCreateRequest)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		panic(ErrRequestFailed)
	}
	defer resp.Body.Close()

	var blockNumber BlockNumberResponse
	if err := json.NewDecoder(resp.Body).Decode(&blockNumber); err != nil {
		panic(ErrFailedToParseJSON)
	}

	if blockNumber.ID != payload.ID || len(blockNumber.Result) <= 2 {
		panic(ErrResultIsUnexpected)
	}

	number, err := strconv.ParseInt(blockNumber.Result[2:], 16, 64)
	if err != nil {
		panic(ErrParseToIntFailed)
	}

	return int(number)
}

func (p *ethParser) Subscribe(address string) bool {
	p.transactions[address] = []Transaction{}
	return true
}

func (p *ethParser) GetTransactions(address string) []Transaction {
	return p.transactions[address]
}

func (p *ethParser) nextID() uint32 {
	return p.idCounter.Add(1)
}

func (p *ethParser) backgroundProcess(ctx context.Context) error {
	u, err := url.Parse(p.wssProvider)
	if err != nil {
		return ErrInvalidAddress
	}

	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return ErrDialFailed
	}

	// Upgrade to TLS if necessary
	if u.Scheme == "wss" {
		conn = tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true,
		})
	}

	// Form the HTTP request for WebSocket handshake
	requestHeader := http.Header{
		"Connection":            {"Upgrade"},
		"Upgrade":               {"websocket"},
		"Sec-WebSocket-Version": {"13"},
		"Sec-WebSocket-Key":     {"dGhlIHNhbXBsZSBub25jZQ=="},
	}

	req := http.Request{
		Method: "GET",
		URL:    u,
		Host:   u.Host,
		Header: requestHeader,
	}

	if err = req.Write(conn); err != nil {
		return ErrWriteMessageFailed
	}

	// Read the response from the server
	resp, err := http.ReadResponse(bufio.NewReader(conn), &req)
	if err != nil {
		return ErrResultIsUnexpected
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return ErrUnexpectedStatusCode
	}

	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return
	// 	default:

	// 	}
	// }

	return nil
}
