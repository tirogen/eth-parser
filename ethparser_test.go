package ethparser

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

func TestE2EGetCurrentBlock(t *testing.T) {
	p, err := New(context.TODO(), &Config{
		HTTPProvider: "https://cloudflare-eth.com",
		WSSProvider:  "wss://ethereum-rpc.publicnode.com",
		MaxInboxSize: 100,
		Client:       http.DefaultClient,
	})
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}
	block := p.GetCurrentBlock()
	if block == 0 {
		t.Fatalf("expected block number, got 0")
	}
}

func TestSubscribe(t *testing.T) {
	p, err := New(context.TODO(), &Config{
		HTTPProvider: "https://cloudflare-eth.com",
		WSSProvider:  "wss://ethereum-rpc.publicnode.com",
		MaxInboxSize: 100,
		Client:       http.DefaultClient,
	})
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}
	result := p.Subscribe("0x0000000000000000000000000000000000000002")
	if !result {
		t.Fatalf("failed to subscribe")
	}
}

type MockRoundTripper struct {
	// Define a function to return a mocked response
	roundTripFunc func(req *http.Request) *http.Response
}

// RoundTrip executes a single HTTP transaction and returns a mocked response
func (mrt *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return mrt.roundTripFunc(req), nil
}

func TestGetTransactions(t *testing.T) {
	ctx := context.Background()

	// Mock websocket result
	ctrl := gomock.NewController(t)
	websocketConnect := NewMockWebSocketConnect(ctrl)
	mockResult := make(chan string, 2)
	mockResult <- `{"jsonrpc":"2.0","id":1,"result":"0x7d196008e9ffbe655b64a52231ae5cae"}`
	mockResult <- `{"jsonrpc":"2.0","method":"eth_subscription","params":{"subscription":"0x7d196008e9ffbe655b64a52231ae5cae","result":"0x63336879edd91368ff2f924b605249f0e3b4926590c6afcfca7a02753b8c94a8"}}`
	websocketConnect.
		EXPECT().
		Subscribe(ctx, `{"id":1,"jsonrpc":"2.0","method":"eth_subscribe","params":["newPendingTransactions"]}`).
		Return(mockResult, nil)

	// Mock transaction result
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"jsonrpc":"2.0","result":{"type":"0x2","blockHash":"0x4f07d5497a16732a919647e5b7eb2c2cf3926ee51c3e20a33ae6991539b0f4b1","blockNumber":"0x1300bb6","from":"0x91199826dbc27ae3033357d91b6fd3b7eb4d2149","gas":"0x575f2","hash":"0x63336879edd91368ff2f924b605249f0e3b4926590c6afcfca7a02753b8c94a8","input":"0xe7a050aa000000000000000000000000acb55c530acdb2849e6d4f36992cd8c9d50ed8f7000000000000000000000000ec53bf9167f50cdeb3ae105f56099aaab9061f83000000000000000000000000000000000000000000000030c90a87fcc5acf964","nonce":"0x1b2","to":"0x858646372cc42e1a627fce94aa7a7033e7cf075a","transactionIndex":"0x95","value":"0x0","v":"0x0","r":"0xdca2c2411486787ba3e17c43c2d254a37af4287b707b57a35cf0d87a9493c9d8","s":"0x2b25156d6181d86416aceaaaaa98611a4270a890113009f4be0e92ec1c973085","gasPrice":"0x35f03481c","maxFeePerGas":"0x4ba2f83c3","maxPriorityFeePerGas":"0x2cdd988","chainId":"0x1","accessList":[]},"id":1}`)),
		Header:     make(http.Header),
	}
	// Create a new MockRoundTripper
	mockRoundTripper := &MockRoundTripper{
		roundTripFunc: func(req *http.Request) *http.Response {
			return mockResponse
		},
	}

	// Create an HTTP client with the mock RoundTripper
	client := &http.Client{Transport: mockRoundTripper}

	p, err := NewWithWebSocketConnect(ctx, &Config{
		HTTPProvider: "https://cloudflare-eth.com",
		WSSProvider:  "wss://ethereum-rpc.publicnode.com",
		MaxInboxSize: 100,
		Client:       client,
	}, websocketConnect, map[string]bool{
		"0x91199826dbc27ae3033357d91b6fd3b7eb4d2149": true,
	})
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	// waiting background process in go routine
	<-time.After(time.Second)

	transactions := p.GetTransactions("0x91199826dbc27ae3033357d91b6fd3b7eb4d2149")
	if len(transactions) != 1 {
		t.Fatalf("expected 1 transaction")
	}
}
