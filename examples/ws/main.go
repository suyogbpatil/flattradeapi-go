package main

import (
	"context"
	"fmt"
	"time"

	"github.com/suyotech/flattradeapi-go/instruments"
	"github.com/suyotech/flattradeapi-go/ws"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userID := "your-user-id"
	accountID := userID
	accessToken := "your-access-token"

	client := ws.NewWSClient(userID, accountID, accessToken).SetReconnect(2*time.Second, 10).SetOnTick(func(tick ws.Tick) {
		fmt.Printf("tick: %+v\n", tick)
	}).SetOnReconnected(func(info ws.ReconnectInfo) {
		fmt.Printf("ws reconnected; restored %d touchline tokens\n", info.TouchlineRestored)
	}).SetOnError(func(err error) {
		fmt.Println("ws diagnostic:", err)
	}).SetOnDisconnected(func(err error) {
		fmt.Println("ws unavailable:", err)
	})

	if err := client.Connect(ctx); err != nil {
		panic(err)
	}
	defer client.Disconnect()

	nifty := instruments.Instrument{Exchange: "NSE", Token: "26000", Symbol: "NIFTY"}
	if err := client.SubscribeTouchline(nifty); err != nil {
		panic(err)
	}

	<-ctx.Done()
}
