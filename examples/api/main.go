package main

import (
	"context"
	"fmt"
	"os"

	"github.com/suyotech/flattradeapi-go/api"
)

func main() {
	ctx := context.Background()

	client := api.NewClient(api.WithAPIKeys(
		os.Getenv("FLATTRADE_API_KEY"),
		os.Getenv("FLATTRADE_API_SECRET"),
	))

	fmt.Println("login URL:", client.GetLoginURL())

	if requestCode := os.Getenv("FLATTRADE_REQUEST_CODE"); requestCode != "" {
		session, err := client.GenerateSession(ctx, api.SessionRequest{
			RequestCode: requestCode,
		})
		if err != nil {
			panic(err)
		}
		fmt.Printf("session: %+v\n", session)
	}

	client.UserID = os.Getenv("FLATTRADE_USER_ID")
	client.SetAccessToken(os.Getenv("FLATTRADE_ACCESS_TOKEN"))

	if client.AccessToken == "" || client.UserID == "" {
		return
	}

	details, err := client.UserDetails(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("user details: %+v\n", details)
}
