package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/suyotech/flattradeapi-go/ws"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := ws.NewWSClient(
		os.Getenv("FLATTRADE_USER_ID"),
		os.Getenv("FLATTRADE_ACCOUNT_ID"),
		os.Getenv("FLATTRADE_ACCESS_TOKEN"),
	).SetOnTick(func(tick ws.Tick) {
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

	if err := client.SubscribeTouchline("NSE|26000"); err != nil {
		panic(err)
	}

	<-ctx.Done()
}
