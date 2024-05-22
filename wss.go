package ethparser

import (
	"context"
	"log"

	"golang.org/x/net/websocket"
)

type WebSocketConnect interface {
	Subscribe(ctx context.Context, method string) (<-chan string, error)
}

type webSocketConnect struct {
	url       string
	inboxSize int
}

func NewWebSocketConnect(url string, inboxSize int) (WebSocketConnect, error) {
	return &webSocketConnect{
		url:       url,
		inboxSize: inboxSize,
	}, nil
}

func (w *webSocketConnect) Subscribe(ctx context.Context, method string) (<-chan string, error) {
	// origin url can be any valid url
	ws, err := websocket.Dial(w.url, "", "http://localhost/")
	if err != nil {
		return nil, ErrDialFailed
	}

	_, err = ws.Write([]byte(method))
	if err != nil {
		return nil, ErrWriteMessageFailed
	}

	ch := make(chan string, w.inboxSize)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg := make([]byte, 512)
				n := 0
				n, err = ws.Read(msg)
				if err != nil {
					log.Printf("read: %v", err)
				} else {
					ch <- string(msg[:n])
				}
			}
		}
	}()

	return ch, nil
}
