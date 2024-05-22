package ethparser

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
)

type Parser interface {
	// last parsed block
	GetCurrentBlock() int

	// add address to observer
	Subscribe(address string) bool

	// list of inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}

type ethParser struct {
	mu                sync.RWMutex
	httpProvider      string
	wssProvider       string
	client            *http.Client
	idCounter         atomic.Uint32
	subscribedAddress map[string]bool
	websocketConnect  WebSocketConnect
	database          Database
}

type Config struct {
	HTTPProvider string
	WSSProvider  string
	Client       *http.Client
	MaxInboxSize int
}

func New(ctx context.Context, c *Config) (Parser, error) {
	websocketConnect, err := NewWebSocketConnect(c.WSSProvider, c.MaxInboxSize)
	if err != nil {
		return nil, err
	}
	database := NewDatabase()
	parser := &ethParser{
		mu:                sync.RWMutex{},
		httpProvider:      c.HTTPProvider,
		wssProvider:       c.WSSProvider,
		client:            c.Client,
		idCounter:         atomic.Uint32{},
		subscribedAddress: make(map[string]bool),
		websocketConnect:  websocketConnect,
		database:          database,
	}
	go func() {
		if err := parser.backgroundProcess(ctx); err != nil {
			log.Printf("background process failed: %v", err)
		}
	}()
	return parser, nil
}

func NewWithWebSocketConnect(ctx context.Context, c *Config, websocketConnect WebSocketConnect, subscribedAddress map[string]bool) (Parser, error) {
	database := NewDatabase()
	parser := &ethParser{
		mu:                sync.RWMutex{},
		httpProvider:      c.HTTPProvider,
		wssProvider:       c.WSSProvider,
		client:            c.Client,
		idCounter:         atomic.Uint32{},
		subscribedAddress: subscribedAddress,
		websocketConnect:  websocketConnect,
		database:          database,
	}
	go func() {
		if err := parser.backgroundProcess(ctx); err != nil {
			log.Printf("background process failed: %v", err)
		}
	}()
	return parser, nil
}

func (p *ethParser) GetCurrentBlock() int {
	payload := EthRequest{
		Method:  "eth_blockNumber",
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
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribedAddress[address] = true
	return true
}

func (p *ethParser) isSubscribed(address string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.subscribedAddress[address]
	return ok
}

func (p *ethParser) GetTransactions(address string) []Transaction {
	return p.database.GetTransactions(address)
}

func (p *ethParser) nextID() uint32 {
	return p.idCounter.Add(1)
}

// internal methods for getting transaction by hash which is used in background process
func (p *ethParser) getTransactionByHash(ctx context.Context, hash string) (*Transaction, error) {
	payload := EthRequest{
		Method:  "eth_getTransactionByHash",
		Params:  []string{hash},
		ID:      p.nextID(),
		JsonRPC: "2.0",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, ErrFailedToParseJSON
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.httpProvider, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, ErrFailedToCreateRequest
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, ErrRequestFailed
	}
	defer resp.Body.Close()

	var transactionByHash TransactionByHashResponse
	if err := json.NewDecoder(resp.Body).Decode(&transactionByHash); err != nil {
		return nil, ErrFailedToParseJSON
	}

	return transactionByHash.Result, nil
}

func (p *ethParser) backgroundProcess(ctx context.Context) error {
	method := `{"id":1,"jsonrpc":"2.0","method":"eth_subscribe","params":["newPendingTransactions"]}`
	outCh, err := p.websocketConnect.Subscribe(ctx, method)
	if err != nil {
		return err
	}

	for message := range outCh {
		var result SubscribeTransactionResult
		if err := json.Unmarshal([]byte(message), &result); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			continue
		}
		// can skip first message which return subscription id
		if result.Params != nil {
			transaction, err := p.getTransactionByHash(ctx, result.Params.Result)
			if err != nil {
				log.Printf("failed to get transaction by hash: %v", err)
				continue
			}
			if p.isSubscribed(transaction.From) {
				p.database.AddTransaction(transaction.From, *transaction)
			}
			if transaction.To != nil {
				if p.isSubscribed(*transaction.To) {
					p.database.AddTransaction(*transaction.To, *transaction)
				}
			}
		}
	}

	return nil
}
