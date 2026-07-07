package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	UserID      string
	AccountID   string
	APIKey      string
	APISecret   string
	AccessToken string
}

type Option func(*Client)

func NewClient(opts ...Option) *Client {
	c := &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.BaseURL = strings.TrimRight(baseURL, "/")
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.HTTPClient = httpClient
		}
	}
}

func WithCredentials(userID, accountID, accessToken string) Option {
	return func(c *Client) {
		c.UserID = userID
		c.AccountID = accountID
		c.AccessToken = accessToken
	}
}

func WithAPIKey(apiKey string) Option {
	return func(c *Client) {
		c.APIKey = apiKey
	}
}

func WithAPISecret(apiSecret string) Option {
	return func(c *Client) {
		c.APISecret = apiSecret
	}
}

func WithAPIKeys(apiKey, apiSecret string) Option {
	return func(c *Client) {
		c.APIKey = apiKey
		c.APISecret = apiSecret
	}
}

func (c *Client) SetAccessToken(token string) {
	c.AccessToken = token
}

func (c *Client) NewRequest(ctx context.Context, method, path string, payload any) (*http.Request, error) {
	var body io.Reader
	if payload != nil {
		if reader, ok := payload.(io.Reader); ok {
			body = reader
		} else {
			data, err := json.Marshal(payload)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(data)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.url(path), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}

	return req, nil
}

func (c *Client) Do(req *http.Request, out any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &Error{StatusCode: resp.StatusCode, Body: string(data)}
	}
	if out == nil {
		return nil
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return errors.New("flattrade api: empty response")
	}

	var status apiStatus
	if err := json.Unmarshal(data, &status); err == nil {
		if strings.EqualFold(status.Status, "Not_Ok") {
			return &APIError{Status: status.Status, Message: status.Message}
		}
		if status.Status == "" && status.Message != "" {
			return &APIError{Status: status.Status, Message: status.Message}
		}
	}

	return json.Unmarshal(data, out)
}

func (c *Client) postAuthenticated(ctx context.Context, endpoint string, payload any, out any) error {
	if c.AccessToken == "" {
		return errors.New("access token is required")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	form := url.Values{}
	form.Set("jKey", c.AccessToken)

	body := "jData=" + string(data) + "&" + form.Encode()
	req, err := c.NewRequest(ctx, "POST", endpoint, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.Do(req, out)
}

func (c *Client) url(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return strings.TrimRight(c.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")
}

type Error struct {
	StatusCode int
	Body       string
}

func (e *Error) Error() string {
	return fmt.Sprintf("flattrade api: status %d: %s", e.StatusCode, e.Body)
}

type apiStatus struct {
	Status  string `json:"stat,omitempty"`
	Message string `json:"emsg,omitempty"`
}

type APIError struct {
	Status  string
	Message string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return "flattrade api: " + e.Status
	}
	if e.Status == "" {
		return "flattrade api: " + e.Message
	}
	return fmt.Sprintf("flattrade api: %s: %s", e.Status, e.Message)
}
