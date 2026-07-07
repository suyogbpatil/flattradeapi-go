package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const DefaultURL = "wss://piconnect.flattrade.in/PiConnectWSTp"

type TickHandler func(Tick)
type ErrorHandler func(error)
type CloseHandler func(error)

type WSClient struct {
	URL         string
	UserID      string
	AccountID   string
	AccessToken string
	Source      string

	Timeout           time.Duration
	PingInterval      time.Duration
	ReconnectInterval time.Duration
	MaxReconnects     int

	OnTick  TickHandler
	OnError ErrorHandler
	OnClose CloseHandler

	mu      sync.Mutex
	conn    *websocket.Conn
	closed  bool
	dialer  *websocket.Dialer
	done    chan struct{}
	writeMu sync.Mutex
}

type Option func(*WSClient)

func NewClient(opts ...Option) *WSClient {
	c := &WSClient{
		URL:               DefaultURL,
		Source:            "API",
		Timeout:           15 * time.Second,
		PingInterval:      25 * time.Second,
		ReconnectInterval: 2 * time.Second,
		MaxReconnects:     10,
		done:              make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.dialer == nil {
		c.dialer = &websocket.Dialer{
			HandshakeTimeout: c.Timeout,
			Proxy:            http.ProxyFromEnvironment,
		}
	}

	return c
}

func NewWSClient(userID, accountID, accessToken string) *WSClient {
	return NewClient(WithCredentials(userID, accountID, accessToken))
}

func NewWSClientWithURL(url, userID, accountID, accessToken string) *WSClient {
	return NewClient(WithURL(url), WithCredentials(userID, accountID, accessToken))
}

func WithURL(url string) Option {
	return func(c *WSClient) {
		c.URL = url
	}
}

func WithCredentials(userID, accountID, accessToken string) Option {
	return func(c *WSClient) {
		c.UserID = userID
		c.AccountID = accountID
		c.AccessToken = accessToken
	}
}

func WithSource(source string) Option {
	return func(c *WSClient) {
		if source != "" {
			c.Source = source
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *WSClient) {
		if timeout > 0 {
			c.Timeout = timeout
		}
	}
}

func WithReconnect(interval time.Duration, max int) Option {
	return func(c *WSClient) {
		if interval > 0 {
			c.ReconnectInterval = interval
		}
		c.MaxReconnects = max
	}
}

func WithPingInterval(interval time.Duration) Option {
	return func(c *WSClient) {
		if interval > 0 {
			c.PingInterval = interval
		}
	}
}

func WithTickHandler(handler TickHandler) Option {
	return func(c *WSClient) {
		c.OnTick = handler
	}
}

func WithErrorHandler(handler ErrorHandler) Option {
	return func(c *WSClient) {
		c.OnError = handler
	}
}

func WithCloseHandler(handler CloseHandler) Option {
	return func(c *WSClient) {
		c.OnClose = handler
	}
}

func (c *WSClient) SetOnTick(handler TickHandler) *WSClient {
	c.OnTick = handler
	return c
}

func (c *WSClient) SetOnError(handler ErrorHandler) *WSClient {
	c.OnError = handler
	return c
}

func (c *WSClient) SetOnClose(handler CloseHandler) *WSClient {
	c.OnClose = handler
	return c
}

func (c *WSClient) SetCredentials(userID, accountID, accessToken string) *WSClient {
	c.UserID = userID
	c.AccountID = accountID
	c.AccessToken = accessToken
	return c
}

func (c *WSClient) SetReconnect(interval time.Duration, max int) *WSClient {
	if interval > 0 {
		c.ReconnectInterval = interval
	}
	c.MaxReconnects = max
	return c
}

func (c *WSClient) Connect(ctx context.Context) error {
	if err := c.validate(); err != nil {
		return err
	}

	if err := c.connectOnce(ctx); err != nil {
		return err
	}

	go c.readLoop(ctx)
	go c.pingLoop(ctx)
	return nil
}

func (c *WSClient) Disconnect() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	close(c.done)
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()

	if conn == nil {
		return nil
	}

	c.writeMu.Lock()
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.writeMu.Unlock()
	return conn.Close()
}

func (c *WSClient) Reconnect(ctx context.Context) error {
	_ = c.closeConn()
	return c.connectOnce(ctx)
}

func (c *WSClient) SubscribeTouchline(tokens ...string) error {
	return c.send(SubscribeMessage{Event: EventSubscribeTouchline, Tokens: tokens})
}

func (c *WSClient) UnsubscribeTouchline(tokens ...string) error {
	return c.send(SubscribeMessage{Event: EventUnsubscribeTouchline, Tokens: tokens})
}

func (c *WSClient) SubscribeDepth(tokens ...string) error {
	return c.send(SubscribeMessage{Event: EventSubscribeDepth, Tokens: tokens})
}

func (c *WSClient) UnsubscribeDepth(tokens ...string) error {
	return c.send(SubscribeMessage{Event: EventUnsubscribeDepth, Tokens: tokens})
}

func (c *WSClient) SubscribeOrderUpdate() error {
	return c.send(SubscribeMessage{Event: EventSubscribeOrderUpdate})
}

func (c *WSClient) UnsubscribeOrderUpdate() error {
	return c.send(SubscribeMessage{Event: EventUnsubscribeOrderUpdate})
}

func (c *WSClient) SubscribePositionUpdate() error {
	return c.send(SubscribeMessage{Event: EventSubscribePositionUpdate})
}

func (c *WSClient) UnsubscribePositionUpdate() error {
	return c.send(SubscribeMessage{Event: EventUnsubscribePositionUpdate})
}

func (c *WSClient) Send(v any) error {
	return c.send(v)
}

func (c *WSClient) connectOnce(ctx context.Context) error {
	conn, _, err := c.dialer.DialContext(ctx, c.URL, nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.closed = false
	c.mu.Unlock()

	return c.login()
}

func (c *WSClient) login() error {
	return c.send(LoginMessage{
		Event:       EventConnect,
		UserID:      c.UserID,
		AccountID:   c.AccountID,
		AccessToken: c.AccessToken,
		Source:      c.Source,
	})
}

func (c *WSClient) readLoop(ctx context.Context) {
	reconnects := 0
	for {
		conn := c.currentConn()
		if conn == nil {
			return
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			if c.isClosed() {
				c.emitClose(err)
				return
			}
			c.emitError(err)
			if !c.shouldReconnect(err, reconnects) {
				c.emitClose(err)
				return
			}
			reconnects++
			if err := c.waitAndReconnect(ctx); err != nil {
				c.emitClose(err)
				return
			}
			continue
		}
		reconnects = 0

		var tick Tick
		if err := json.Unmarshal(data, &tick); err != nil {
			c.emitError(err)
			continue
		}
		if c.OnTick != nil {
			c.OnTick(tick)
		}
	}
}

func (c *WSClient) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(c.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		case <-ticker.C:
			if err := c.send(HeartbeatMessage{Event: EventHeartbeat}); err != nil {
				c.emitError(err)
			}
		}
	}
}

func (c *WSClient) waitAndReconnect(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return errors.New("websocket closed")
	case <-time.After(c.ReconnectInterval):
	}
	return c.Reconnect(ctx)
}

func (c *WSClient) shouldReconnect(err error, count int) bool {
	if c.MaxReconnects >= 0 && count >= c.MaxReconnects {
		return false
	}
	var netErr net.Error
	return websocket.IsUnexpectedCloseError(err) || errors.As(err, &netErr)
}

func (c *WSClient) send(v any) error {
	conn := c.currentConn()
	if conn == nil {
		return errors.New("websocket is not connected")
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return conn.WriteJSON(v)
}

func (c *WSClient) closeConn() error {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()
	if conn == nil {
		return nil
	}
	return conn.Close()
}

func (c *WSClient) currentConn() *websocket.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn
}

func (c *WSClient) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

func (c *WSClient) validate() error {
	if c.URL == "" {
		return errors.New("websocket url is required")
	}
	if c.UserID == "" {
		return errors.New("user id is required")
	}
	if c.AccessToken == "" {
		return errors.New("access token is required")
	}
	return nil
}

func (c *WSClient) emitError(err error) {
	if c.OnError != nil {
		c.OnError(err)
	}
}

func (c *WSClient) emitClose(err error) {
	if c.OnClose != nil {
		c.OnClose(err)
	}
}
