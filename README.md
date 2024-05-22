# ethparser

Package `ethparser` uses to subscribe transactions from provided address

## Example

```go
package main

import (
	"log"
	"github.com/tirogen/ethparser"
)

func main() {
    client, err := ethparser.New(context.TODO(), &Config{
		HTTPProvider: "https://cloudflare-eth.com",
		WSSProvider:  "wss://ethereum-rpc.publicnode.com",
		MaxInboxSize: 100,
		Client:       http.DefaultClient,
	})
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Subscribe to transactions
    p.Subscribe("0x0000000000000000000000000000000000000002")

    // Get transactions
    transactions := p.GetTransactions("0x0000000000000000000000000000000000000002")
    for _, tx := range transactions {
        log.Println(tx)
    }
}
```

## Design

This library are trying to be as simple as possible. It is avoiding to use any external dependencies and trying to be as fast as possible.
