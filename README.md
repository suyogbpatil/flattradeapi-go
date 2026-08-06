# flattradeapi-go

Go SDK for Flattrade Pi REST APIs, websocket market feed, and instrument master files.

## Install

```sh
go get github.com/suyotech/flattradeapi-go
```

## Packages

- `api`: REST client, auth/session, orders, holdings, market data, alerts, funds, scrips, quotes.
- `ws`: websocket client with connect, disconnect, reconnect, callbacks, and subscriptions.
- `instruments`: instrument master download, cache check, load, search, expiry, and strike helpers.

## REST Auth

```go
client := api.NewClient(api.WithAPIKeys(apiKey, apiSecret))

loginURL := client.GetLoginURL()

session, err := client.GenerateSession(ctx, api.SessionRequest{
	RequestCode: requestCode,
})
```

`GenerateSession` stores the returned access token on the client:

```go
client.AccessToken
```

## User Details

```go
client := api.NewClient(api.WithCredentials(userID, accountID, accessToken))

details, err := client.UserDetails(ctx)
```

## Orders

```go
resp, err := client.PlaceOrder(ctx, api.OrderRequest{
	Exchange:        "NSE",
	TradingSymbol:   "SBIN-EQ",
	Quantity:        "1",
	Price:           "0",
	Product:         "I",
	TransactionType: "B",
	PriceType:       "MKT",
	Retention:       "DAY",
})
```

Other order methods include `ModifyOrder`, `CancelOrder`, `OrderBook`, `TradeBook`, `PositionBook`, `OrderMargin`, `BasketMargin`, GTT, and OCO helpers.

## Market Data

```go
candles, err := client.TimePriceSeries(ctx, api.TimePriceSeriesRequest{
	Exchange:  "NSE",
	Token:     "26000",
	StartTime: "1700000000",
	EndTime:   "1700003600",
	Interval:  "1",
})
```

## Websocket

```go
client := ws.NewWSClient(userID, accountID, accessToken).
	SetReconnect(2*time.Second, 10).
	SetLivenessTimeout(60*time.Second).
	SetOnTick(func(tick ws.Tick) {
		fmt.Println(tick)
	}).
	SetOnReconnected(func(info ws.ReconnectInfo) {
		fmt.Printf("reconnected after attempt %d; restored %d touchline and %d depth tokens\n",
			info.Attempt, info.TouchlineRestored, info.DepthRestored)
	}).
	SetOnDisconnected(func(err error) {
		fmt.Println("websocket unavailable:", err)
	}).
	SetOnError(func(err error) {
		fmt.Println("websocket diagnostic:", err)
	})

err := client.Connect(ctx)
nifty := instruments.Instrument{Exchange: "NSE", Token: "26000", Symbol: "NIFTY"}
err = client.SubscribeTouchline(nifty)
```

The client records and deduplicates desired touchline, depth, order-update, and
position-update subscriptions. After a reconnect is logged in, it restores the
desired subscriptions and waits for every acknowledgement the broker protocol
provides; only then does `OnReconnected` run. FlatTrade defines no acknowledgement
for order-update subscription, so its successful socket write is the confirmation.
Successful unsubscribe calls remove the corresponding desired state.
`ResetSubscriptions` clears local desired state without sending an unsubscribe
request.

`SetReconnect(interval, max)` uses `interval` as the initial exponential-backoff
delay (capped by `MaxReconnectInterval`) and `max` as retries after a connection
has been lost; the defaults are 2 seconds and 10 retries. Use
`DisableReconnect()` (or `SetReconnectEnabled(false)`) to explicitly disable SDK
recovery. Zero permits no retries; a negative `max` retains the existing bounded
setting. When SDK reconnect is enabled, applications should not start a second
reconnect loop.

`OnError` remains backward compatible and is diagnostic, not terminal. Typed
errors and the `OnConnectionError`, `OnLoginError`, `OnSubscriptionError`, and
`OnMessageError` callbacks separate failure categories. `OnDisconnected` and
the legacy `OnClose` run only after the socket is intentionally disconnected or
automatic recovery stops. `IsConnected()` and `LastMessageAt()` expose safe
connection status. Liveness is refreshed by any valid broker message, including
heartbeat acknowledgements; it does not depend on price changes or per-token
tick frequency.

## Instruments

```go
items, err := instruments.CheckInstruments("./insts")

filter := instruments.Filter{
	Exchange:   "NFO",
	Symbol:     "NIFTY",
	Instrument: "OPTIDX",
}

matches := instruments.FindInstrument(items, filter)
expiries := instruments.GetExpiry(items, filter)
strikes := instruments.GetOptionStrike(items, filter, 25000, 5)
```

`CheckInstruments` downloads all configured instrument master files when missing or older than today 8:30 AM.

## Examples

Runnable examples are under:

- `examples/api`
- `examples/ws`
- `examples/instruments`

## License

MIT. See `LICENSE`.
