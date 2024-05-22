package ethparser

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketConnect interface {
	Subscribe(ctx context.Context, method string) (<-chan string, error)
}

type webSocketConnect struct {
	url       string
	dialer    *websocket.Dialer
	inboxSize int
}

func NewWebSocketConnect(url string, inboxSize int) (WebSocketConnect, error) {
	return &webSocketConnect{
		url: url,
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		},
		inboxSize: inboxSize,
	}, nil
}

func (w *webSocketConnect) Subscribe(ctx context.Context, method string) (<-chan string, error) {
	c, _, err := w.dialer.DialContext(ctx, w.url, nil)
	if err != nil {
		return nil, ErrDialFailed
	}

	err = c.WriteMessage(websocket.TextMessage, []byte(method))
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
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Printf("read: %v", err)
				} else {
					ch <- string(message)
				}
			}
		}
	}()

	return ch, nil
}
