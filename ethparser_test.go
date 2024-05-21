package ethparser

import (
	"context"
	"testing"
	"time"
)

func TestE2EGetCurrentBlock(t *testing.T) {
	p, err := New(context.TODO(), &Config{
		HTTPProvider:        "https://cloudflare-eth.com",
		WSSProvider:         "wss://ethereum-rpc.publicnode.com:443",
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		Timeout:             10 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}
	block := p.GetCurrentBlock()
	if block == 0 {
		t.Fatalf("expected block number, got 0")
	}
}
