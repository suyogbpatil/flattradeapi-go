package api

import (
	"context"
	"errors"
)

type OrderRequest struct {
	UserID          string `json:"uid,omitempty"`
	AccountID       string `json:"actid,omitempty"`
	Exchange        string `json:"exch,omitempty"`
	Token           string `json:"token,omitempty"`
	TradingSymbol   string `json:"tsym,omitempty"`
	Quantity        string `json:"qty,omitempty"`
	DisclosedQty    string `json:"dscqty,omitempty"`
	Price           string `json:"prc,omitempty"`
	TriggerPrice    string `json:"trgprc,omitempty"`
	Product         string `json:"prd,omitempty"`
	TransactionType string `json:"trantype,omitempty"`
	PriceType       string `json:"prctyp,omitempty"`
	Retention       string `json:"ret,omitempty"`
	Remarks         string `json:"remarks,omitempty"`
}

type ModifyOrderRequest struct {
	UserID        string `json:"uid,omitempty"`
	AccountID     string `json:"actid,omitempty"`
	OrderNo       string `json:"norenordno,omitempty"`
	Exchange      string `json:"exch,omitempty"`
	TradingSymbol string `json:"tsym,omitempty"`
	Quantity      string `json:"qty,omitempty"`
	Price         string `json:"prc,omitempty"`
	TriggerPrice  string `json:"trgprc,omitempty"`
	PriceType     string `json:"prctyp,omitempty"`
	Retention     string `json:"ret,omitempty"`
}

type CancelOrderRequest struct {
	UserID  string `json:"uid,omitempty"`
	OrderNo string `json:"norenordno,omitempty"`
}

type SingleOrderHistoryRequest struct {
	UserID  string `json:"uid,omitempty"`
	OrderNo string `json:"norenordno,omitempty"`
}

type ProductConversionRequest struct {
	UserID          string `json:"uid,omitempty"`
	AccountID       string `json:"actid,omitempty"`
	Exchange        string `json:"exch,omitempty"`
	TradingSymbol   string `json:"tsym,omitempty"`
	Quantity        string `json:"qty,omitempty"`
	Product         string `json:"prd,omitempty"`
	PreviousProduct string `json:"prevprd,omitempty"`
	TransactionType string `json:"trantype,omitempty"`
	PositionType    string `json:"postype,omitempty"`
}

type OrderMarginRequest struct {
	UserID          string `json:"uid,omitempty"`
	AccountID       string `json:"actid,omitempty"`
	Exchange        string `json:"exch,omitempty"`
	Token           string `json:"token,omitempty"`
	TradingSymbol   string `json:"tsym,omitempty"`
	Quantity        string `json:"qty,omitempty"`
	Price           string `json:"prc,omitempty"`
	Product         string `json:"prd,omitempty"`
	TransactionType string `json:"trantype,omitempty"`
	PriceType       string `json:"prctyp,omitempty"`
}

type BasketMarginRequest struct {
	UserID    string         `json:"uid,omitempty"`
	AccountID string         `json:"actid,omitempty"`
	Orders    []OrderRequest `json:"basketlists,omitempty"`
}

type GTTOrderRequest struct {
	UserID          string `json:"uid,omitempty"`
	AccountID       string `json:"actid,omitempty"`
	AlertID         string `json:"al_id,omitempty"`
	Exchange        string `json:"exch,omitempty"`
	Token           string `json:"token,omitempty"`
	TradingSymbol   string `json:"tsym,omitempty"`
	Quantity        string `json:"qty,omitempty"`
	Price           string `json:"prc,omitempty"`
	TriggerPrice    string `json:"trgprc,omitempty"`
	Product         string `json:"prd,omitempty"`
	TransactionType string `json:"trantype,omitempty"`
	PriceType       string `json:"prctyp,omitempty"`
	Retention       string `json:"ret,omitempty"`
	Remarks         string `json:"remarks,omitempty"`
}

type OCOOrderRequest struct {
	UserID          string `json:"uid,omitempty"`
	AccountID       string `json:"actid,omitempty"`
	AlertID         string `json:"al_id,omitempty"`
	Exchange        string `json:"exch,omitempty"`
	Token           string `json:"token,omitempty"`
	TradingSymbol   string `json:"tsym,omitempty"`
	Quantity        string `json:"qty,omitempty"`
	Price           string `json:"prc,omitempty"`
	TriggerPrice    string `json:"trgprc,omitempty"`
	BookLossPrice   string `json:"blprc,omitempty"`
	BookProfitPrice string `json:"bpprc,omitempty"`
	Product         string `json:"prd,omitempty"`
	TransactionType string `json:"trantype,omitempty"`
	PriceType       string `json:"prctyp,omitempty"`
	Retention       string `json:"ret,omitempty"`
	Remarks         string `json:"remarks,omitempty"`
}

type OrderResponse struct {
	OrderNo string `json:"norenordno,omitempty"`
}

type OrderBookEntry struct {
	Order
	OrderNo      string `json:"norenordno,omitempty"`
	Status       string `json:"status,omitempty"`
	RejectReason string `json:"rejreason,omitempty"`
	AveragePrice string `json:"avgprc,omitempty"`
	FilledShares string `json:"fillshares,omitempty"`
	OrderTime    string `json:"norentm,omitempty"`
}

type TradeBookEntry struct {
	OrderNo         string `json:"norenordno,omitempty"`
	Exchange        string `json:"exch,omitempty"`
	Token           string `json:"token,omitempty"`
	TradingSymbol   string `json:"tsym,omitempty"`
	Quantity        string `json:"qty,omitempty"`
	Price           string `json:"prc,omitempty"`
	Product         string `json:"prd,omitempty"`
	TransactionType string `json:"trantype,omitempty"`
	TradeNo         string `json:"flid,omitempty"`
	TradeTime       string `json:"fltm,omitempty"`
}

type PositionBookEntry struct {
	Exchange      string `json:"exch,omitempty"`
	Token         string `json:"token,omitempty"`
	TradingSymbol string `json:"tsym,omitempty"`
	Product       string `json:"prd,omitempty"`
	NetQuantity   string `json:"netqty,omitempty"`
	BuyQuantity   string `json:"daybuyqty,omitempty"`
	SellQuantity  string `json:"daysellqty,omitempty"`
	NetAverage    string `json:"netavgprc,omitempty"`
}

func (c *Client) PlaceOrder(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	c.fillOrderDefaults(&req.UserID, &req.AccountID)

	var out OrderResponse
	if err := c.postAuthenticated(ctx, EndpointPlaceOrder, req, &out); err != nil {
		return nil, err
	}
	if out.OrderNo == "" {
		return nil, errors.New("place order response missing order number")
	}
	return &out, nil
}

func (c *Client) ModifyOrder(ctx context.Context, req ModifyOrderRequest) (*OrderResponse, error) {
	if req.UserID == "" {
		req.UserID = c.UserID
	}

	var out OrderResponse
	if err := c.postAuthenticated(ctx, EndpointModifyOrder, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CancelOrder(ctx context.Context, req CancelOrderRequest) (*OrderResponse, error) {
	if req.UserID == "" {
		req.UserID = c.UserID
	}

	var out OrderResponse
	if err := c.postAuthenticated(ctx, EndpointCancelOrder, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ExitSNOOrder(ctx context.Context, req CancelOrderRequest) (*OrderResponse, error) {
	if req.UserID == "" {
		req.UserID = c.UserID
	}

	var out OrderResponse
	if err := c.postAuthenticated(ctx, EndpointExitSNOOrder, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) OrderMargin(ctx context.Context, req OrderMarginRequest) (map[string]any, error) {
	c.fillOrderDefaults(&req.UserID, &req.AccountID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointOrderMargin, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) BasketMargin(ctx context.Context, req BasketMarginRequest) (map[string]any, error) {
	c.fillOrderDefaults(&req.UserID, &req.AccountID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointBasketMargin, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) OrderBook(ctx context.Context) ([]OrderBookEntry, error) {
	var out []OrderBookEntry
	if err := c.postAuthenticated(ctx, EndpointOrderBook, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) TradeBook(ctx context.Context) ([]TradeBookEntry, error) {
	var out []TradeBookEntry
	if err := c.postAuthenticated(ctx, EndpointTradeBook, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) PositionBook(ctx context.Context) ([]PositionBookEntry, error) {
	var out []PositionBookEntry
	if err := c.postAuthenticated(ctx, EndpointPositionBook, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) SingleOrderHistory(ctx context.Context, req SingleOrderHistoryRequest) ([]OrderBookEntry, error) {
	if req.UserID == "" {
		req.UserID = c.UserID
	}

	var out []OrderBookEntry
	if err := c.postAuthenticated(ctx, EndpointSingleOrderHistory, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ProductConversion(ctx context.Context, req ProductConversionRequest) (*OrderResponse, error) {
	c.fillOrderDefaults(&req.UserID, &req.AccountID)

	var out OrderResponse
	if err := c.postAuthenticated(ctx, EndpointProductConversion, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) MultiLegOrderBook(ctx context.Context) ([]OrderBookEntry, error) {
	var out []OrderBookEntry
	if err := c.postAuthenticated(ctx, EndpointMultiLegOrderBook, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) PlaceGTTOrder(ctx context.Context, req GTTOrderRequest) (map[string]any, error) {
	c.fillOrderDefaults(&req.UserID, &req.AccountID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointPlaceGTTOrder, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ModifyGTTOrder(ctx context.Context, req GTTOrderRequest) (map[string]any, error) {
	c.fillOrderDefaults(&req.UserID, &req.AccountID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointModifyGTTOrder, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CancelGTTOrder(ctx context.Context, req GTTOrderRequest) (map[string]any, error) {
	c.fillOrderDefaults(&req.UserID, &req.AccountID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointCancelGTTOrder, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) PendingGTTOrders(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointPendingGTTOrders, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) EnabledGTTs(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointEnabledGTTs, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) PlaceOCOOrder(ctx context.Context, req OCOOrderRequest) (map[string]any, error) {
	c.fillOrderDefaults(&req.UserID, &req.AccountID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointPlaceOCOOrder, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ModifyOCOOrder(ctx context.Context, req OCOOrderRequest) (map[string]any, error) {
	c.fillOrderDefaults(&req.UserID, &req.AccountID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointModifyOCOOrder, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CancelOCOOrder(ctx context.Context, req OCOOrderRequest) (map[string]any, error) {
	c.fillOrderDefaults(&req.UserID, &req.AccountID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointCancelOCOOrder, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) userPayload() map[string]string {
	return map[string]string{"uid": c.UserID}
}

func (c *Client) fillUser(userID *string) {
	if *userID == "" {
		*userID = c.UserID
	}
}

func (c *Client) fillOrderDefaults(userID *string, accountID *string) {
	if *userID == "" {
		*userID = c.UserID
	}
	if *accountID == "" {
		*accountID = c.AccountID
	}
}
