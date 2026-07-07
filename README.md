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
	SetOnTick(func(tick ws.Tick) {
		fmt.Println(tick)
	}).
	SetOnError(func(err error) {
		fmt.Println(err)
	})

err := client.Connect(ctx)
err = client.SubscribeTouchline("NSE|26000")
```

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
