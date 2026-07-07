package api

import "context"

type PayoutRequest struct {
	UserID    string `json:"uid,omitempty"`
	Amount    string `json:"amt,omitempty"`
	Remarks   string `json:"remarks,omitempty"`
	RequestID string `json:"req_id,omitempty"`
}

func (c *Client) MaxPayoutAmount(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointMaxPayoutAmount, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) FundsPayout(ctx context.Context, req PayoutRequest) (map[string]any, error) {
	c.fillUser(&req.UserID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointFundsPayout, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) PayinReport(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointPayinReport, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) PayoutReport(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointPayoutReport, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CancelPayout(ctx context.Context, req PayoutRequest) (map[string]any, error) {
	c.fillUser(&req.UserID)
	var out map[string]any
	if err := c.postAuthenticated(ctx, EndpointCancelPayout, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}
