package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
)

type LoginURLRequest struct {
	APIKey string
}

type SessionRequest struct {
	RequestCode string `json:"request_code"`
	APIKey      string `json:"api_key,omitempty"`
	APISecret   string `json:"api_secret,omitempty"`
}

type SessionResponse struct {
	UserID      string `json:"uid,omitempty"`
	AccountID   string `json:"actid,omitempty"`
	AccessToken string `json:"token,omitempty"`
	Status      string `json:"stat,omitempty"`
	Message     string `json:"emsg,omitempty"`
}

func GetLoginURL(apiKey string) string {
	u, _ := url.Parse(DefaultLoginURL)
	q := u.Query()
	q.Set(loginURLAPIKeyName, apiKey)
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *Client) GetLoginURL() string {
	return GetLoginURL(c.APIKey)
}

func (c *Client) GenerateSession(ctx context.Context, req SessionRequest) (*SessionResponse, error) {
	if req.APIKey == "" {
		req.APIKey = c.APIKey
	}
	if req.APISecret == "" {
		req.APISecret = c.APISecret
	}
	if req.APIKey == "" {
		return nil, errors.New("api key is required")
	}
	if req.APISecret == "" {
		return nil, errors.New("api secret is required")
	}
	if req.RequestCode == "" {
		return nil, errors.New("request code is required")
	}

	secretHash := sha256.Sum256([]byte(req.APIKey + req.RequestCode + req.APISecret))
	req.APISecret = hex.EncodeToString(secretHash[:])

	httpReq, err := c.NewRequest(ctx, "POST", DefaultSessionURL, req)
	if err != nil {
		return nil, err
	}

	var out SessionResponse
	if err := c.Do(httpReq, &out); err != nil {
		return nil, err
	}

	if out.AccessToken != "" {
		c.SetAccessToken(out.AccessToken)
	}
	if out.UserID != "" {
		c.UserID = out.UserID
	}
	if out.AccountID != "" {
		c.AccountID = out.AccountID
	}

	return &out, nil
}

func (c *Client) UserDetails(ctx context.Context) (*UserDetails, error) {
	if c.UserID == "" {
		return nil, errors.New("user id is required")
	}
	if c.AccessToken == "" {
		return nil, errors.New("access token is required")
	}

	var out UserDetails
	if err := c.postAuthenticated(ctx, EndpointUserDetails, map[string]string{
		"uid": c.UserID,
	}, &out); err != nil {
		return nil, err
	}

	return &out, nil
}
