package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
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
type ConnectedHandler func()
type ReconnectAttemptHandler func(attempt, maxAttempts int)
type ReconnectedHandler func(ReconnectInfo)
type DisconnectedHandler func(error)

type ReconnectInfo struct {
	Attempt           int
	TouchlineRestored int
	DepthRestored     int
	OrderRestored     bool
	PositionRestored  bool
}

type connectResult struct {
	info    ReconnectInfo
	conn    *websocket.Conn
	pending [][]byte
}

type WSClient struct {
	URL         string
	UserID      string
	AccountID   string
	AccessToken string
	Source      string

	Timeout              time.Duration
	PingInterval         time.Duration
	LivenessTimeout      time.Duration
	ReconnectInterval    time.Duration
	MaxReconnectInterval time.Duration
	MaxReconnects        int
	ReconnectEnabled     bool
	Debug                bool

	OnTick              TickHandler
	OnAck               AckHandler
	OnOrder             OrderUpdateHandler
	OnPosition          PositionUpdateHandler
	OnMessage           MessageHandler
	OnError             ErrorHandler // Diagnostic: it does not imply that recovery has stopped.
	OnClose             CloseHandler
	OnConnected         ConnectedHandler
	OnReconnectAttempt  ReconnectAttemptHandler
	OnReconnected       ReconnectedHandler
	OnDisconnected      DisconnectedHandler
	OnConnectionError   ErrorHandler
	OnLoginError        ErrorHandler
	OnSubscriptionError ErrorHandler
	OnMessageError      ErrorHandler

	lifecycleMu          sync.Mutex
	mu                   sync.Mutex
	conn                 *websocket.Conn
	connected            bool
	intentional          bool
	generation           uint64
	runCancel            context.CancelFunc
	loopRunning          bool
	lastMessage          time.Time
	disconnectedNotified bool
	dialer               *websocket.Dialer
	writeMu              sync.Mutex
	expectedCloses       map[*websocket.Conn]chan struct{}

	subMu              sync.Mutex
	touchlineDesired   map[string]struct{}
	depthDesired       map[string]struct{}
	orderDesired       bool
	positionDesired    bool
	touchlineConfirmed map[string]struct{}
	depthConfirmed     map[string]struct{}
	orderConfirmed     bool
	positionConfirmed  bool
}

type Option func(*WSClient)

func NewClient(opts ...Option) *WSClient {
	c := &WSClient{
		URL: DefaultURL, Source: "API", Timeout: 45 * time.Second,
		PingInterval: 25 * time.Second, LivenessTimeout: 60 * time.Second,
		ReconnectInterval: 2 * time.Second, MaxReconnectInterval: 30 * time.Second,
		MaxReconnects: 10, ReconnectEnabled: true,
		touchlineDesired: make(map[string]struct{}), depthDesired: make(map[string]struct{}),
		touchlineConfirmed: make(map[string]struct{}), depthConfirmed: make(map[string]struct{}),
		expectedCloses: make(map[*websocket.Conn]chan struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.dialer == nil {
		c.dialer = &websocket.Dialer{HandshakeTimeout: c.Timeout}
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

func WithURL(url string) Option { return func(c *WSClient) { c.URL = url } }
func WithCredentials(userID, accountID, accessToken string) Option {
	return func(c *WSClient) { c.SetCredentials(userID, accountID, accessToken) }
}
func WithSource(source string) Option { return func(c *WSClient) { c.SetSource(source) } }
func WithTimeout(timeout time.Duration) Option {
	return func(c *WSClient) {
		if timeout > 0 {
			c.Timeout = timeout
		}
	}
}
func WithReconnect(interval time.Duration, max int) Option {
	return func(c *WSClient) { c.SetReconnect(interval, max) }
}
func WithPingInterval(interval time.Duration) Option {
	return func(c *WSClient) {
		if interval > 0 {
			c.PingInterval = interval
		}
	}
}
func WithLivenessTimeout(timeout time.Duration) Option {
	return func(c *WSClient) {
		if timeout > 0 {
			c.LivenessTimeout = timeout
		}
	}
}
func WithDebug(debug bool) Option                        { return func(c *WSClient) { c.Debug = debug } }
func WithTickHandler(h TickHandler) Option               { return func(c *WSClient) { c.OnTick = h } }
func WithAckHandler(h AckHandler) Option                 { return func(c *WSClient) { c.OnAck = h } }
func WithOrderUpdateHandler(h OrderUpdateHandler) Option { return func(c *WSClient) { c.OnOrder = h } }
func WithPositionUpdateHandler(h PositionUpdateHandler) Option {
	return func(c *WSClient) { c.OnPosition = h }
}
func WithMessageHandler(h MessageHandler) Option { return func(c *WSClient) { c.OnMessage = h } }
func WithErrorHandler(h ErrorHandler) Option     { return func(c *WSClient) { c.OnError = h } }
func WithCloseHandler(h CloseHandler) Option     { return func(c *WSClient) { c.OnClose = h } }

func (c *WSClient) SetOnTick(h TickHandler) *WSClient                     { c.OnTick = h; return c }
func (c *WSClient) SetOnAck(h AckHandler) *WSClient                       { c.OnAck = h; return c }
func (c *WSClient) SetOnOrderUpdate(h OrderUpdateHandler) *WSClient       { c.OnOrder = h; return c }
func (c *WSClient) SetOnPositionUpdate(h PositionUpdateHandler) *WSClient { c.OnPosition = h; return c }
func (c *WSClient) SetOnMessage(h MessageHandler) *WSClient               { c.OnMessage = h; return c }
func (c *WSClient) SetOnError(h ErrorHandler) *WSClient                   { c.OnError = h; return c }
func (c *WSClient) SetOnClose(h CloseHandler) *WSClient                   { c.OnClose = h; return c }
func (c *WSClient) SetOnConnected(h ConnectedHandler) *WSClient           { c.OnConnected = h; return c }
func (c *WSClient) SetOnReconnectAttempt(h ReconnectAttemptHandler) *WSClient {
	c.OnReconnectAttempt = h
	return c
}
func (c *WSClient) SetOnReconnected(h ReconnectedHandler) *WSClient   { c.OnReconnected = h; return c }
func (c *WSClient) SetOnDisconnected(h DisconnectedHandler) *WSClient { c.OnDisconnected = h; return c }
func (c *WSClient) SetOnConnectionError(h ErrorHandler) *WSClient     { c.OnConnectionError = h; return c }
func (c *WSClient) SetOnLoginError(h ErrorHandler) *WSClient          { c.OnLoginError = h; return c }
func (c *WSClient) SetOnSubscriptionError(h ErrorHandler) *WSClient {
	c.OnSubscriptionError = h
	return c
}
func (c *WSClient) SetOnMessageError(h ErrorHandler) *WSClient { c.OnMessageError = h; return c }

func (c *WSClient) SetCredentials(userID, accountID, accessToken string) *WSClient {
	if accountID == "" {
		accountID = userID
	}
	c.UserID, c.AccountID, c.AccessToken = userID, accountID, accessToken
	return c
}
func (c *WSClient) SetSource(source string) *WSClient {
	if source != "" {
		c.Source = source
	}
	return c
}

// SetReconnect sets the initial retry delay and retries after a lost connection.
// A negative max retains the current bounded value; zero permits no retries.
func (c *WSClient) SetReconnect(interval time.Duration, max int) *WSClient {
	if interval > 0 {
		c.ReconnectInterval = interval
	}
	if max >= 0 {
		c.MaxReconnects = max
	}
	return c
}
func (c *WSClient) SetReconnectEnabled(enabled bool) *WSClient {
	c.ReconnectEnabled = enabled
	return c
}
func (c *WSClient) DisableReconnect() *WSClient { return c.SetReconnectEnabled(false) }
func (c *WSClient) SetLivenessTimeout(timeout time.Duration) *WSClient {
	if timeout > 0 {
		c.LivenessTimeout = timeout
	}
	return c
}
func (c *WSClient) SetDebug(debug bool) *WSClient { c.Debug = debug; return c }
func (c *WSClient) IsDebug() bool                 { return c.Debug }
func (c *WSClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected && c.conn != nil
}
func (c *WSClient) LastMessageAt() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.lastMessage }

func (c *WSClient) Connect(ctx context.Context) error {
	c.lifecycleMu.Lock()
	if err := c.validate(); err != nil {
		c.lifecycleMu.Unlock()
		return err
	}
	if c.IsConnected() {
		c.lifecycleMu.Unlock()
		return nil
	}
	runCtx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	oldCancel := c.runCancel
	c.runCancel = cancel
	c.intentional = false
	c.disconnectedNotified = false
	c.generation++
	c.loopRunning = true
	gen := c.generation
	c.mu.Unlock()
	if oldCancel != nil {
		oldCancel()
	}
	result, err := c.connectOnce(runCtx, false)
	if err != nil {
		cancel()
		c.mu.Lock()
		if c.generation == gen {
			c.loopRunning = false
		}
		c.mu.Unlock()
		c.lifecycleMu.Unlock()
		c.dispatchPending(result.pending)
		c.emitDiagnostic(err)
		return err
	}
	c.lifecycleMu.Unlock()
	c.dispatchPending(result.pending)
	if !c.sameGeneration(gen) || c.currentConn() != result.conn {
		return nil
	}
	go c.readLoop(runCtx, gen)
	go c.pingLoop(runCtx, gen)
	if c.OnConnected != nil {
		c.OnConnected()
	}
	return nil
}

func (c *WSClient) Disconnect() error {
	c.lifecycleMu.Lock()
	c.mu.Lock()
	c.intentional = true
	c.generation++
	cancel := c.runCancel
	c.runCancel = nil
	c.loopRunning = false
	conn := c.conn
	c.conn = nil
	c.connected = false
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	var err error
	if conn != nil {
		c.writeMu.Lock()
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.writeMu.Unlock()
		err = conn.Close()
	}
	c.lifecycleMu.Unlock()
	c.notifyDisconnected(nil)
	return err
}

func (c *WSClient) Reconnect(ctx context.Context) error {
	c.lifecycleMu.Lock()
	if err := c.validate(); err != nil {
		c.lifecycleMu.Unlock()
		return err
	}
	c.mu.Lock()
	startLoops := !c.loopRunning
	gen := c.generation
	runCtx := ctx
	var newCancel context.CancelFunc
	if startLoops {
		if c.runCancel != nil {
			c.runCancel()
		}
		runCtx, newCancel = context.WithCancel(ctx)
		c.runCancel = newCancel
		c.intentional = false
		c.disconnectedNotified = false
		c.generation++
		gen = c.generation
		c.loopRunning = true
	}
	c.mu.Unlock()
	oldConn := c.currentConn()
	var expectedClose chan struct{}
	if !startLoops && oldConn != nil {
		expectedClose = c.markExpectedClose(oldConn)
	}
	c.closeCurrent()
	result, err := c.connectOnce(runCtx, true)
	if err != nil {
		if startLoops {
			newCancel()
			c.mu.Lock()
			c.loopRunning = false
			c.mu.Unlock()
		}
		c.lifecycleMu.Unlock()
		c.dispatchPending(result.pending)
		c.emitDiagnostic(err)
		c.finishExpectedClose(expectedClose)
		return err
	}
	c.lifecycleMu.Unlock()
	c.dispatchPending(result.pending)
	c.finishExpectedClose(expectedClose)
	if c.currentConn() != result.conn {
		return nil
	}
	if startLoops {
		go c.readLoop(runCtx, gen)
		go c.pingLoop(runCtx, gen)
	}
	if c.OnReconnected != nil {
		c.OnReconnected(result.info)
	}
	return nil
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
	return c.changeTokens(EventSubscribeTouchline, tokens, true)
}
func (c *WSClient) UnsubscribeTouchlineTokens(tokens ...string) error {
	return c.changeTokens(EventUnsubscribeTouchline, tokens, false)
}
func (c *WSClient) SubscribeDepthTokens(tokens ...string) error {
	return c.changeTokens(EventSubscribeDepth, tokens, true)
}
func (c *WSClient) UnsubscribeDepthTokens(tokens ...string) error {
	return c.changeTokens(EventUnsubscribeDepth, tokens, false)
}

func (c *WSClient) SubscribeOrderUpdate() error {
	return c.changeAccountSubscription(EventSubscribeOrderUpdate, true)
}
func (c *WSClient) UnsubscribeOrderUpdate() error {
	return c.changeAccountSubscription(EventUnsubscribeOrderUpdate, false)
}
func (c *WSClient) SubscribePositionUpdate() error {
	return c.changeAccountSubscription(EventSubscribePositionUpdate, true)
}
func (c *WSClient) UnsubscribePositionUpdate() error {
	return c.changeAccountSubscription(EventUnsubscribePositionUpdate, false)
}
func (c *WSClient) Send(v any) error { return c.send(v) }

// ResetSubscriptions clears locally desired and confirmed state. It sends no broker request.
func (c *WSClient) ResetSubscriptions() {
	c.lifecycleMu.Lock()
	defer c.lifecycleMu.Unlock()
	c.subMu.Lock()
	defer c.subMu.Unlock()
	c.touchlineDesired = make(map[string]struct{})
	c.depthDesired = make(map[string]struct{})
	c.touchlineConfirmed = make(map[string]struct{})
	c.depthConfirmed = make(map[string]struct{})
	c.orderDesired, c.positionDesired, c.orderConfirmed, c.positionConfirmed = false, false, false, false
}

func (c *WSClient) changeTokens(event Event, tokens []string, subscribe bool) error {
	tokens = uniqueTokens(tokens)
	if len(tokens) == 0 {
		return errors.New("at least one token is required")
	}
	c.lifecycleMu.Lock()
	defer c.lifecycleMu.Unlock()
	c.subMu.Lock()
	defer c.subMu.Unlock()
	desired := c.touchlineDesired
	if event == EventSubscribeDepth || event == EventUnsubscribeDepth {
		desired = c.depthDesired
	}
	if subscribe {
		filtered := tokens[:0]
		for _, token := range tokens {
			if _, ok := desired[token]; !ok {
				filtered = append(filtered, token)
			}
		}
		tokens = filtered
		if len(tokens) == 0 {
			return nil
		}
	}
	if err := c.send(SubscribeMessage{Event: event, Key: strings.Join(tokens, "#")}); err != nil {
		return &SubscriptionError{Event: event, Err: err}
	}
	for _, token := range tokens {
		if subscribe {
			desired[token] = struct{}{}
		} else {
			delete(desired, token)
			delete(c.confirmedMap(event), token)
		}
	}
	return nil
}

func (c *WSClient) changeAccountSubscription(event Event, subscribe bool) error {
	c.lifecycleMu.Lock()
	defer c.lifecycleMu.Unlock()
	c.subMu.Lock()
	defer c.subMu.Unlock()
	current := c.positionDesired
	if event == EventSubscribeOrderUpdate || event == EventUnsubscribeOrderUpdate {
		current = c.orderDesired
	}
	if current == subscribe {
		return nil
	}
	var msg any = SubscribeMessage{Event: event}
	if subscribe {
		msg = AccountMessage{Event: event, AccountID: c.AccountID}
	}
	if err := c.send(msg); err != nil {
		return &SubscriptionError{Event: event, Err: err}
	}
	if event == EventSubscribeOrderUpdate || event == EventUnsubscribeOrderUpdate {
		c.orderDesired = subscribe
		c.orderConfirmed = subscribe // The broker defines no order-subscription acknowledgement.
	} else {
		c.positionDesired = subscribe
		if !subscribe {
			c.positionConfirmed = false
		}
	}
	return nil
}

func uniqueTokens(tokens []string) []string {
	seen := make(map[string]struct{}, len(tokens))
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if _, ok := seen[token]; !ok {
			seen[token] = struct{}{}
			out = append(out, token)
		}
	}
	return out
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
	return uniqueTokens(tokens), nil
}

func (c *WSClient) connectOnce(ctx context.Context, restoring bool) (connectResult, error) {
	conn, resp, err := c.dialer.DialContext(ctx, c.URL, nil)
	if err != nil {
		return connectResult{}, &ConnectionError{Op: "dial", Err: handshakeError(err, resp)}
	}
	c.mu.Lock()
	c.conn = conn
	c.connected = false
	c.lastMessage = time.Time{}
	c.mu.Unlock()
	if err := c.login(ctx, conn); err != nil {
		c.closeCurrent()
		return connectResult{}, err
	}
	info, pending, err := c.restoreSubscriptions(ctx, conn, restoring)
	if err != nil {
		c.closeCurrent()
		return connectResult{info: info, conn: conn, pending: pending}, err
	}
	c.mu.Lock()
	if c.conn == conn {
		c.connected = true
	}
	c.mu.Unlock()
	return connectResult{info: info, conn: conn, pending: pending}, nil
}

func handshakeError(err error, resp *http.Response) error {
	if resp == nil {
		return err
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if readErr != nil || len(body) == 0 {
		return fmt.Errorf("%w: status %d", err, resp.StatusCode)
	}
	return fmt.Errorf("%w: status %d: %s", err, resp.StatusCode, string(body))
}

func (c *WSClient) login(ctx context.Context, conn *websocket.Conn) error {
	if err := c.writeJSON(conn, LoginMessage{Event: EventConnect, UserID: c.UserID, AccountID: c.AccountID, AccessToken: c.AccessToken, Source: c.Source}); err != nil {
		return &ConnectionError{Op: "login write", Err: err}
	}
	data, err := c.readWithTimeout(ctx, conn, c.Timeout)
	if err != nil {
		return &BrokerLoginError{Err: err}
	}
	var ack ConnectAck
	if err := json.Unmarshal(data, &ack); err != nil {
		return &BrokerLoginError{Err: err}
	}
	if ack.Event != EventConnectAck {
		return &BrokerLoginError{Err: fmt.Errorf("unexpected acknowledgement event %q", ack.Event)}
	}
	if !positiveStatus(ack.Status) {
		return &BrokerLoginError{Status: ack.Status}
	}
	c.markMessage()
	return nil
}

func (c *WSClient) restoreSubscriptions(ctx context.Context, conn *websocket.Conn, restoring bool) (ReconnectInfo, [][]byte, error) {
	touch, depth, order, position := c.desiredSnapshot()
	restoring = restoring || len(touch) > 0 || len(depth) > 0 || order || position
	c.clearConfirmed()
	info := ReconnectInfo{TouchlineRestored: len(touch), DepthRestored: len(depth), OrderRestored: order, PositionRestored: position}
	var pending [][]byte
	requests := []struct {
		event, ack Event
		tokens     []string
	}{
		{EventSubscribeTouchline, EventTouchlineAck, touch}, {EventSubscribeDepth, EventDepthAck, depth},
	}
	for _, req := range requests {
		if len(req.tokens) == 0 {
			continue
		}
		if err := c.writeJSON(conn, SubscribeMessage{Event: req.event, Key: strings.Join(req.tokens, "#")}); err != nil {
			return info, pending, &SubscriptionError{Event: req.event, Restore: restoring, Err: err}
		}
		acks, err := c.awaitAcks(ctx, conn, req.ack, len(req.tokens), restoring)
		pending = append(pending, acks...)
		if err != nil {
			return info, pending, err
		}
	}
	if order {
		if err := c.writeJSON(conn, AccountMessage{Event: EventSubscribeOrderUpdate, AccountID: c.AccountID}); err != nil {
			return info, pending, &SubscriptionError{Event: EventSubscribeOrderUpdate, Restore: restoring, Err: err}
		}
		c.subMu.Lock()
		c.orderConfirmed = true // FlatTrade documents no acknowledgement for this subscription.
		c.subMu.Unlock()
	}
	if position {
		if err := c.writeJSON(conn, AccountMessage{Event: EventSubscribePositionUpdate, AccountID: c.AccountID}); err != nil {
			return info, pending, &SubscriptionError{Event: EventSubscribePositionUpdate, Restore: restoring, Err: err}
		}
		acks, err := c.awaitAcks(ctx, conn, EventSubscribePositionUpdateAck, 1, restoring)
		pending = append(pending, acks...)
		if err != nil {
			return info, pending, err
		}
	}
	return info, pending, nil
}

func (c *WSClient) awaitAcks(ctx context.Context, conn *websocket.Conn, expected Event, count int, restoring bool) ([][]byte, error) {
	seen := make(map[string]struct{}, count)
	pending := make([][]byte, 0, count)
	for len(seen) < count {
		data, err := c.readWithTimeout(ctx, conn, c.Timeout)
		if err != nil {
			return pending, &SubscriptionError{Event: expected, Restore: restoring, Err: err}
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			return pending, &SubscriptionError{Event: expected, Restore: restoring, Err: err}
		}
		c.markMessage()
		pending = append(pending, append([]byte(nil), data...))
		if msg.Event != expected {
			continue
		}
		var ack Ack
		if err := json.Unmarshal(data, &ack); err != nil {
			return pending, &SubscriptionError{Event: expected, Restore: restoring, Err: err}
		}
		if !positiveStatus(ack.Status) {
			return pending, &SubscriptionError{Event: expected, Restore: restoring, Err: fmt.Errorf("broker status %q", ack.Status)}
		}
		c.confirmAck(ack)
		key := ack.Exchange + "|" + ack.Token
		if key == "|" {
			key = fmt.Sprintf("ack-%d", len(seen))
		}
		seen[key] = struct{}{}
	}
	return pending, nil
}

func (c *WSClient) readLoop(ctx context.Context, gen uint64) {
	defer func() {
		c.mu.Lock()
		if c.generation == gen {
			c.loopRunning = false
		}
		c.mu.Unlock()
	}()
	for {
		conn := c.currentConn()
		if conn == nil || !c.sameGeneration(gen) {
			return
		}
		deadline := time.Now().Add(c.LivenessTimeout)
		if last := c.LastMessageAt(); !last.IsZero() {
			deadline = last.Add(c.LivenessTimeout)
		}
		_ = conn.SetReadDeadline(deadline)
		_, data, err := conn.ReadMessage()
		if err != nil {
			if expected := c.takeExpectedClose(conn); expected != nil {
				select {
				case <-ctx.Done():
					return
				case <-expected:
				}
				if current := c.currentConn(); current != nil && current != conn {
					continue
				}
			}
			if !c.sameGeneration(gen) || c.isIntentional() {
				return
			}
			failure := &ConnectionError{Op: "read/liveness", Err: err}
			c.emitDiagnostic(failure)
			if !c.recover(ctx, gen, conn, failure) {
				return
			}
			continue
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.emitDiagnostic(&MessageError{Err: err})
			continue
		}
		c.markMessage()
		if err := c.dispatch(data); err != nil {
			c.emitDiagnostic(err)
		}
	}
}

func (c *WSClient) recover(ctx context.Context, gen uint64, failed *websocket.Conn, cause error) bool {
	c.lifecycleMu.Lock()
	if !c.sameGeneration(gen) || c.isIntentional() {
		c.lifecycleMu.Unlock()
		return false
	}
	if current := c.currentConn(); current != failed && current != nil {
		c.lifecycleMu.Unlock()
		return true
	}
	c.closeCurrent()
	if !c.ReconnectEnabled || c.MaxReconnects == 0 {
		c.lifecycleMu.Unlock()
		c.notifyDisconnected(cause)
		return false
	}
	c.lifecycleMu.Unlock()
	var lastErr error = cause
	for attempt := 1; attempt <= c.MaxReconnects; attempt++ {
		if c.OnReconnectAttempt != nil {
			c.OnReconnectAttempt(attempt, c.MaxReconnects)
		}
		if err := c.waitBackoff(ctx, gen, attempt); err != nil {
			lastErr = err
			break
		}
		c.lifecycleMu.Lock()
		if !c.sameGeneration(gen) || c.isIntentional() {
			c.lifecycleMu.Unlock()
			return false
		}
		if c.IsConnected() {
			c.lifecycleMu.Unlock()
			return true
		}
		result, err := c.connectOnce(ctx, true)
		c.lifecycleMu.Unlock()
		if err == nil {
			c.dispatchPending(result.pending)
			if c.currentConn() != result.conn {
				return c.IsConnected()
			}
			result.info.Attempt = attempt
			if c.OnReconnected != nil {
				c.OnReconnected(result.info)
			}
			return true
		}
		c.dispatchPending(result.pending)
		lastErr = err
		c.emitDiagnostic(err)
	}
	c.notifyDisconnected(lastErr)
	return false
}

func (c *WSClient) pingLoop(ctx context.Context, gen uint64) {
	ticker := time.NewTicker(c.PingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			c.closeForGeneration(gen)
			return
		case <-ticker.C:
			if !c.sameGeneration(gen) {
				return
			}
			conn := c.currentConn()
			if conn == nil || !c.IsConnected() {
				continue
			}
			if err := c.writeJSON(conn, HeartbeatMessage{Event: EventHeartbeat}); err != nil {
				failure := &ConnectionError{Op: "heartbeat write", Err: err}
				c.emitDiagnostic(failure)
				c.closeIfCurrent(conn)
			}
		}
	}
}

func (c *WSClient) waitBackoff(ctx context.Context, gen uint64, attempt int) error {
	d := c.ReconnectInterval
	for i := 1; i < attempt && d < c.MaxReconnectInterval; i++ {
		d *= 2
	}
	if d > c.MaxReconnectInterval {
		d = c.MaxReconnectInterval
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		if !c.sameGeneration(gen) {
			return errors.New("websocket lifecycle changed")
		}
		return nil
	}
}

func (c *WSClient) dispatch(data []byte) error {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return &MessageError{Err: err}
	}
	switch msg.Event {
	case EventTouchlineAck, EventDepthAck, EventUnsubscribeTouchline, EventUnsubscribeDepthAck, EventUnsubscribeOrderUpdate, EventUnsubscribeOrderUpdateAck, EventSubscribePositionUpdateAck, EventUnsubscribePositionUpdate, EventUnsubscribePositionUpdateAck:
		var ack Ack
		if err := json.Unmarshal(data, &ack); err != nil {
			return &MessageError{Event: msg.Event, Err: err}
		}
		if c.OnAck != nil {
			c.OnAck(ack)
		}
		if !positiveStatus(ack.Status) {
			return &SubscriptionError{Event: msg.Event, Err: fmt.Errorf("broker status %q", ack.Status)}
		}
		c.confirmAck(ack)
	case EventHeartbeatAck:
		var ack Ack
		if err := json.Unmarshal(data, &ack); err != nil {
			return &MessageError{Event: msg.Event, Err: err}
		}
		if c.OnAck != nil {
			c.OnAck(ack)
		}
		if !positiveStatus(ack.Status) {
			return &ConnectionError{Op: "heartbeat acknowledgement", Err: fmt.Errorf("broker status %q", ack.Status)}
		}
	case EventTouchlineFeed, EventDepthFeed:
		var tick Tick
		if err := json.Unmarshal(data, &tick); err != nil {
			return &MessageError{Event: msg.Event, Err: err}
		}
		if c.OnTick != nil {
			c.OnTick(tick)
		}
	case EventOrderFeed:
		var order OrderUpdate
		if err := json.Unmarshal(data, &order); err != nil {
			return &MessageError{Event: msg.Event, Err: err}
		}
		if c.OnOrder != nil {
			c.OnOrder(order)
		}
	case EventPositionFeed:
		var position PositionUpdate
		if err := json.Unmarshal(data, &position); err != nil {
			return &MessageError{Event: msg.Event, Err: err}
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

func positiveStatus(status string) bool {
	status = strings.TrimSpace(status)
	return status == "" || strings.EqualFold(status, "ok") || strings.EqualFold(status, "success")
}
func (c *WSClient) confirmAck(ack Ack) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	switch ack.Event {
	case EventTouchlineAck:
		c.confirmTokens(c.touchlineConfirmed, c.touchlineDesired, ack)
	case EventDepthAck:
		c.confirmTokens(c.depthConfirmed, c.depthDesired, ack)
	case EventSubscribePositionUpdateAck:
		c.positionConfirmed = true
	}
}
func (c *WSClient) confirmTokens(confirmed, desired map[string]struct{}, ack Ack) {
	keys := uniqueTokens(strings.Split(ack.Key, "#"))
	if len(keys) == 0 && ack.Exchange != "" && ack.Token != "" {
		keys = []string{ack.Exchange + "|" + ack.Token}
	}
	if len(keys) == 0 {
		for token := range desired {
			confirmed[token] = struct{}{}
		}
		return
	}
	for _, token := range keys {
		if _, ok := desired[token]; ok {
			confirmed[token] = struct{}{}
		}
	}
}
func (c *WSClient) confirmedMap(event Event) map[string]struct{} {
	if event == EventSubscribeDepth || event == EventUnsubscribeDepth {
		return c.depthConfirmed
	}
	return c.touchlineConfirmed
}
func (c *WSClient) desiredSnapshot() ([]string, []string, bool, bool) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	touch, depth := mapKeys(c.touchlineDesired), mapKeys(c.depthDesired)
	return touch, depth, c.orderDesired, c.positionDesired
}
func mapKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
func (c *WSClient) clearConfirmed() {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	c.touchlineConfirmed = make(map[string]struct{})
	c.depthConfirmed = make(map[string]struct{})
	c.orderConfirmed = false
	c.positionConfirmed = false
}

func (c *WSClient) send(v any) error {
	conn := c.currentConn()
	if conn == nil || !c.IsConnected() {
		return errors.New("websocket is not connected")
	}
	if err := c.writeJSON(conn, v); err != nil {
		return &ConnectionError{Op: "write", Err: err}
	}
	return nil
}

func (c *WSClient) dispatchPending(messages [][]byte) {
	for _, data := range messages {
		if err := c.dispatch(data); err != nil {
			c.emitDiagnostic(err)
		}
	}
}

func (c *WSClient) markExpectedClose(conn *websocket.Conn) chan struct{} {
	done := make(chan struct{})
	c.mu.Lock()
	c.expectedCloses[conn] = done
	c.mu.Unlock()
	return done
}

func (c *WSClient) finishExpectedClose(done chan struct{}) {
	if done != nil {
		close(done)
	}
}

func (c *WSClient) takeExpectedClose(conn *websocket.Conn) chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	done := c.expectedCloses[conn]
	delete(c.expectedCloses, conn)
	return done
}

func (c *WSClient) writeJSON(conn *websocket.Conn, v any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return conn.WriteJSON(v)
}
func (c *WSClient) readWithTimeout(ctx context.Context, conn *websocket.Conn, timeout time.Duration) ([]byte, error) {
	deadline := time.Now().Add(timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := conn.SetReadDeadline(deadline); err != nil {
		return nil, err
	}
	_, data, err := conn.ReadMessage()
	_ = conn.SetReadDeadline(time.Time{})
	return data, err
}
func (c *WSClient) markMessage()                 { c.mu.Lock(); c.lastMessage = time.Now(); c.mu.Unlock() }
func (c *WSClient) currentConn() *websocket.Conn { c.mu.Lock(); defer c.mu.Unlock(); return c.conn }
func (c *WSClient) sameGeneration(gen uint64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.generation == gen
}
func (c *WSClient) isIntentional() bool { c.mu.Lock(); defer c.mu.Unlock(); return c.intentional }
func (c *WSClient) closeCurrent() {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.connected = false
	c.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}
func (c *WSClient) closeIfCurrent(conn *websocket.Conn) {
	c.mu.Lock()
	if c.conn == conn {
		c.conn = nil
		c.connected = false
	}
	c.mu.Unlock()
	_ = conn.Close()
}
func (c *WSClient) closeForGeneration(gen uint64) {
	if c.sameGeneration(gen) {
		c.closeCurrent()
	}
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
	if c.Timeout <= 0 || c.PingInterval <= 0 || c.LivenessTimeout <= 0 || c.ReconnectInterval <= 0 || c.MaxReconnectInterval <= 0 {
		return errors.New("websocket timeouts and intervals must be positive")
	}
	return nil
}

func (c *WSClient) emitDiagnostic(err error) {
	if err == nil {
		return
	}
	var connection *ConnectionError
	var login *BrokerLoginError
	var subscription *SubscriptionError
	var message *MessageError
	switch {
	case errors.As(err, &connection):
		if c.OnConnectionError != nil {
			c.OnConnectionError(err)
		}
	case errors.As(err, &login):
		if c.OnLoginError != nil {
			c.OnLoginError(err)
		}
	case errors.As(err, &subscription):
		if c.OnSubscriptionError != nil {
			c.OnSubscriptionError(err)
		}
	case errors.As(err, &message):
		if c.OnMessageError != nil {
			c.OnMessageError(err)
		}
	}
	if c.OnError != nil {
		c.OnError(err)
	}
}
func (c *WSClient) notifyDisconnected(err error) {
	c.mu.Lock()
	if c.disconnectedNotified {
		c.mu.Unlock()
		return
	}
	c.disconnectedNotified = true
	c.connected = false
	c.mu.Unlock()
	if c.OnDisconnected != nil {
		c.OnDisconnected(err)
	}
	if c.OnClose != nil {
		c.OnClose(err)
	}
}
