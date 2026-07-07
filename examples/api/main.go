package main

import (
	"context"
	"fmt"

	"github.com/suyotech/flattradeapi-go/api"
)

func main() {
	ctx := context.Background()

	apiKey := "your-api-key"
	apiSecret := "your-api-secret"
	requestCode := ""
	userID := "your-user-id"
	accessToken := ""

	client := api.NewClient(api.WithAPIKeys(apiKey, apiSecret))

	fmt.Println("login URL:", client.GetLoginURL())

	if requestCode != "" {
		session, err := client.GenerateSession(ctx, api.SessionRequest{
			RequestCode: requestCode,
		})
		if err != nil {
			panic(err)
		}
		fmt.Printf("session: %+v\n", session)
	}

	client.UserID = userID
	client.SetAccessToken(accessToken)

	if client.AccessToken == "" || client.UserID == "" {
		return
	}

	details, err := client.UserDetails(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("user details: %+v\n", details)
}
