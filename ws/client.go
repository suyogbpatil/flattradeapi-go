package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/suyotech/flattradeapi-go/instruments"
)

const DefaultURL = "wss://piconnect.flattrade.in/PiConnectWSAPI/"

type TickHandler func(Tick)
type AckHandler func(Ack)
type OrderUpdateHandler func(OrderUpdate)
type PositionUpdateHandler func(PositionUpdate)
type MessageHandler func(Message)
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
	Debug             bool

	OnTick     TickHandler
	OnAck      AckHandler
	OnOrder    OrderUpdateHandler
	OnPosition PositionUpdateHandler
	OnMessage  MessageHandler
	OnError    ErrorHandler
	OnClose    CloseHandler

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
		Timeout:           45 * time.Second,
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
		}
	}

	return c
}

func NewWSClient(userID, accountID, accessToken string) *WSClient {
	if accountID == "" {
		accountID = userID
	}
	return NewClient(WithCredentials(userID, accountID, accessToken))
}

func NewWSClientWithURL(url, userID, accountID, accessToken string) *WSClient {
	if accountID == "" {
		accountID = userID
	}
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
		if accountID == "" {
			accountID = userID
		}
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

func WithDebug(debug bool) Option {
	return func(c *WSClient) {
		c.Debug = debug
	}
}

func WithTickHandler(handler TickHandler) Option {
	return func(c *WSClient) {
		c.OnTick = handler
	}
}

func WithAckHandler(handler AckHandler) Option {
	return func(c *WSClient) {
		c.OnAck = handler
	}
}

func WithOrderUpdateHandler(handler OrderUpdateHandler) Option {
	return func(c *WSClient) {
		c.OnOrder = handler
	}
}

func WithPositionUpdateHandler(handler PositionUpdateHandler) Option {
	return func(c *WSClient) {
		c.OnPosition = handler
	}
}

func WithMessageHandler(handler MessageHandler) Option {
	return func(c *WSClient) {
		c.OnMessage = handler
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

func (c *WSClient) SetOnAck(handler AckHandler) *WSClient {
	c.OnAck = handler
	return c
}

func (c *WSClient) SetOnOrderUpdate(handler OrderUpdateHandler) *WSClient {
	c.OnOrder = handler
	return c
}

func (c *WSClient) SetOnPositionUpdate(handler PositionUpdateHandler) *WSClient {
	c.OnPosition = handler
	return c
}

func (c *WSClient) SetOnMessage(handler MessageHandler) *WSClient {
	c.OnMessage = handler
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

func (c *WSClient) SetSource(source string) *WSClient {
	if source != "" {
		c.Source = source
	}
	return c
}

func (c *WSClient) SetReconnect(interval time.Duration, max int) *WSClient {
	if interval > 0 {
		c.ReconnectInterval = interval
	}
	c.MaxReconnects = max
	return c
}

func (c *WSClient) SetDebug(debug bool) *WSClient {
	c.Debug = debug
	return c
}

func (c *WSClient) IsDebug() bool {
	return c.Debug
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

func (c *WSClient) SubscribeTouchline(items ...instruments.Instrument) error {
	tokens, err := instrumentTokens(items)
	if err != nil {
		return err
	}
	return c.SubscribeTouchlineTokens(tokens...)
}

func (c *WSClient) UnsubscribeTouchline(items ...instruments.Instrument) error {
	tokens, err := instrumentTokens(items)
	if err != nil {
		return err
	}
	return c.UnsubscribeTouchlineTokens(tokens...)
}

func (c *WSClient) SubscribeDepth(items ...instruments.Instrument) error {
	tokens, err := instrumentTokens(items)
	if err != nil {
		return err
	}
	return c.SubscribeDepthTokens(tokens...)
}

func (c *WSClient) UnsubscribeDepth(items ...instruments.Instrument) error {
	tokens, err := instrumentTokens(items)
	if err != nil {
		return err
	}
	return c.UnsubscribeDepthTokens(tokens...)
}

func (c *WSClient) SubscribeTouchlineTokens(tokens ...string) error {
	return c.sendTokens(EventSubscribeTouchline, tokens)
}

func (c *WSClient) UnsubscribeTouchlineTokens(tokens ...string) error {
	return c.sendTokens(EventUnsubscribeTouchline, tokens)
}

func (c *WSClient) SubscribeDepthTokens(tokens ...string) error {
	return c.sendTokens(EventSubscribeDepth, tokens)
}

func (c *WSClient) UnsubscribeDepthTokens(tokens ...string) error {
	return c.sendTokens(EventUnsubscribeDepth, tokens)
}

func (c *WSClient) SubscribeOrderUpdate() error {
	return c.send(AccountMessage{Event: EventSubscribeOrderUpdate, AccountID: c.AccountID})
}

func (c *WSClient) UnsubscribeOrderUpdate() error {
	return c.send(SubscribeMessage{Event: EventUnsubscribeOrderUpdate})
}

func (c *WSClient) SubscribePositionUpdate() error {
	return c.send(AccountMessage{Event: EventSubscribePositionUpdate, AccountID: c.AccountID})
}

func (c *WSClient) UnsubscribePositionUpdate() error {
	return c.send(SubscribeMessage{Event: EventUnsubscribePositionUpdate})
}

func (c *WSClient) Send(v any) error {
	return c.send(v)
}

func (c *WSClient) sendTokens(event Event, tokens []string) error {
	if len(tokens) == 0 {
		return errors.New("at least one token is required")
	}
	return c.send(SubscribeMessage{Event: event, Key: strings.Join(tokens, "#")})
}

func instrumentTokens(items []instruments.Instrument) ([]string, error) {
	if len(items) == 0 {
		return nil, errors.New("at least one instrument is required")
	}

	tokens := make([]string, 0, len(items))
	for _, item := range items {
		if item.Exchange == "" || item.Token == "" {
			return nil, errors.New("instrument exchange and token are required")
		}
		tokens = append(tokens, item.Exchange+"|"+item.Token)
	}
	return tokens, nil
}

func (c *WSClient) connectOnce(ctx context.Context) error {
	conn, resp, err := c.dialer.DialContext(ctx, c.URL, nil)
	if err != nil {
		return handshakeError(err, resp)
	}

	c.mu.Lock()
	c.conn = conn
	c.closed = false
	c.mu.Unlock()

	return c.login(ctx)
}

func handshakeError(err error, resp *http.Response) error {
	if resp == nil {
		return err
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if readErr != nil {
		return fmt.Errorf("%w: status %d", err, resp.StatusCode)
	}
	if len(body) == 0 {
		return fmt.Errorf("%w: status %d", err, resp.StatusCode)
	}
	return fmt.Errorf("%w: status %d: %s", err, resp.StatusCode, string(body))
}

func (c *WSClient) login(ctx context.Context) error {
	if err := c.send(LoginMessage{
		Event:       EventConnect,
		UserID:      c.UserID,
		AccountID:   c.AccountID,
		AccessToken: c.AccessToken,
		Source:      c.Source,
	}); err != nil {
		return err
	}

	conn := c.currentConn()
	if conn == nil {
		return errors.New("websocket is not connected")
	}

	deadline := time.Now().Add(c.Timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	if err := conn.SetReadDeadline(deadline); err != nil {
		return err
	}
	_, data, err := conn.ReadMessage()
	_ = conn.SetReadDeadline(time.Time{})
	if err != nil {
		return err
	}

	var ack ConnectAck
	if err := json.Unmarshal(data, &ack); err != nil {
		return err
	}
	if ack.Event != EventConnectAck {
		return fmt.Errorf("unexpected websocket login response: %s", string(data))
	}
	if !strings.EqualFold(ack.Status, "Ok") {
		return fmt.Errorf("websocket login failed: %s", string(data))
	}

	return nil
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
			if c.isClosed() || ctx.Err() != nil {
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

		if err := c.dispatch(data); err != nil {
			c.emitError(err)
		}
	}
}

func (c *WSClient) dispatch(data []byte) error {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	switch msg.Event {
	case EventTouchlineAck, EventDepthAck, EventUnsubscribeTouchline, EventUnsubscribeDepthAck, EventUnsubscribeOrderUpdate, EventUnsubscribeOrderUpdateAck, EventSubscribePositionUpdateAck, EventUnsubscribePositionUpdate, EventUnsubscribePositionUpdateAck, EventHeartbeatAck:
		var ack Ack
		if err := json.Unmarshal(data, &ack); err != nil {
			return err
		}
		if c.OnAck != nil {
			c.OnAck(ack)
		}
	case EventTouchlineFeed, EventDepthFeed:
		var tick Tick
		if err := json.Unmarshal(data, &tick); err != nil {
			return err
		}
		if c.OnTick != nil {
			c.OnTick(tick)
		}
	case EventOrderFeed:
		var order OrderUpdate
		if err := json.Unmarshal(data, &order); err != nil {
			return err
		}
		if c.OnOrder != nil {
			c.OnOrder(order)
		}
	case EventPositionFeed:
		var position PositionUpdate
		if err := json.Unmarshal(data, &position); err != nil {
			return err
		}
		if c.OnPosition != nil {
			c.OnPosition(position)
		}
	default:
		if c.OnMessage != nil {
			c.OnMessage(msg)
		}
	}

	return nil
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
	if c.AccountID == "" {
		return errors.New("account id is required")
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
