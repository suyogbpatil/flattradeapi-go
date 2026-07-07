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

	client := ws.NewWSClient(userID, accountID, accessToken).SetOnTick(func(tick ws.Tick) {
		fmt.Printf("tick: %+v\n", tick)
	}).SetOnError(func(err error) {
		fmt.Println("ws error:", err)
	}).SetOnClose(func(err error) {
		fmt.Println("ws closed:", err)
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
