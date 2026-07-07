package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

var istLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.FixedZone("IST", 5*60*60+30*60)
	}
	return loc
}()

type ExchangeRequest struct {
	UserID   string `json:"uid,omitempty"`
	Exchange string `json:"exch,omitempty"`
}

type TopListRequest struct {
	UserID   string `json:"uid,omitempty"`
	Exchange string `json:"exch,omitempty"`
	ListName string `json:"idxname,omitempty"`
}

type TimePriceSeriesRequest struct {
	UserID    string `json:"uid,omitempty"`
	Exchange  string `json:"exch,omitempty"`
	Token     string `json:"token,omitempty"`
	StartTime string `json:"st,omitempty"`
	EndTime   string `json:"et,omitempty"`
	Interval  string `json:"intrv,omitempty"`
}

type Candle struct {
	Time    time.Time `json:"time,omitempty"`
	Open    string    `json:"into,omitempty"`
	High    string    `json:"inth,omitempty"`
	Low     string    `json:"intl,omitempty"`
	Close   string    `json:"intc,omitempty"`
	Volume  string    `json:"intv,omitempty"`
	OpenInt string    `json:"intoi,omitempty"`
	Value   string    `json:"v,omitempty"`
}

func (c *Candle) UnmarshalJSON(data []byte) error {
	type candleAlias Candle
	var raw struct {
		candleAlias
		Time string `json:"time,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*c = Candle(raw.candleAlias)
	if strings.TrimSpace(raw.Time) == "" {
		return nil
	}

	parsed, err := parseCandleTime(raw.Time)
	if err != nil {
		return err
	}
	c.Time = parsed
	return nil
}

func parseCandleTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if unix, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Unix(unix, 0).In(istLocation), nil
	}

	for _, layout := range []string{
		"02-01-2006 15:04:05",
		"02-01-2006 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	} {
		if parsed, err := time.ParseInLocation(layout, value, istLocation); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("parse candle time %q: unsupported format", value)
}

type EODChartDataRequest struct {
	UserID   string `json:"uid,omitempty"`
	Exchange string `json:"exch,omitempty"`
	Token    string `json:"token,omitempty"`
	Start    string `json:"st,omitempty"`
	End      string `json:"et,omitempty"`
}

type OptionChainRequest struct {
	UserID        string `json:"uid,omitempty"`
	Exchange      string `json:"exch,omitempty"`
	TradingSymbol string `json:"tsym,omitempty"`
	StrikePrice   string `json:"strprc,omitempty"`
	Count         string `json:"cnt,omitempty"`
}

type OptionGreekRequest struct {
	UserID        string `json:"uid,omitempty"`
	Exchange      string `json:"exch,omitempty"`
	TradingSymbol string `json:"tsym,omitempty"`
	Price         string `json:"prc,omitempty"`
}

type SpanCalculatorRequest struct {
	UserID string         `json:"uid,omitempty"`
	Items  map[string]any `json:"items,omitempty"`
}

func (c *Client) IndexList(ctx context.Context, req ExchangeRequest) ([]map[string]any, error) {
	c.fillUser(&req.UserID)
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointIndexList, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) TopListNames(ctx context.Context, req ExchangeRequest) ([]map[string]any, error) {
	c.fillUser(&req.UserID)
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointTopListNames, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) TopList(ctx context.Context, req TopListRequest) ([]map[string]any, error) {
	c.fillUser(&req.UserID)
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointTopList, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) TimePriceSeries(ctx context.Context, req TimePriceSeriesRequest) ([]Candle, error) {
	c.fillUser(&req.UserID)
	var out []Candle
	if err := c.postAuthenticated(ctx, EndpointTimePriceSeries, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) EODChartData(ctx context.Context, req EODChartDataRequest) ([]map[string]any, error) {
	c.fillUser(&req.UserID)
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointEODChartData, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) OptionChain(ctx context.Context, req OptionChainRequest) ([]map[string]any, error) {
	c.fillUser(&req.UserID)
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointOptionChain, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) OptionGreek(ctx context.Context, req OptionGreekRequest) (map[string]any, error) {
	c.fillUser(&req.UserID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointOptionGreek, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ExchangeMessage(ctx context.Context, req ExchangeRequest) ([]map[string]any, error) {
	c.fillUser(&req.UserID)
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointExchangeMessage, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) BrokerMessage(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointBrokerMessage, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) SpanCalculator(ctx context.Context, req SpanCalculatorRequest) (map[string]any, error) {
	c.fillUser(&req.UserID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointSpanCalculator, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}
