package api

import "context"

type AlertRequest struct {
	UserID        string `json:"uid,omitempty"`
	AlertID       string `json:"al_id,omitempty"`
	Exchange      string `json:"exch,omitempty"`
	Token         string `json:"token,omitempty"`
	TradingSymbol string `json:"tsym,omitempty"`
	Price         string `json:"prc,omitempty"`
	AlertType     string `json:"al_type,omitempty"`
	Condition     string `json:"condition,omitempty"`
	Remarks       string `json:"remarks,omitempty"`
}

type AlertResponse struct {
	AlertID string `json:"al_id,omitempty"`
}

func (c *Client) SetAlert(ctx context.Context, req AlertRequest) (*AlertResponse, error) {
	c.fillUser(&req.UserID)
	var out AlertResponse
	if err := c.postAuthenticated(ctx, EndpointSetAlert, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ModifyAlert(ctx context.Context, req AlertRequest) (*AlertResponse, error) {
	c.fillUser(&req.UserID)
	var out AlertResponse
	if err := c.postAuthenticated(ctx, EndpointModifyAlert, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CancelAlert(ctx context.Context, req AlertRequest) (*AlertResponse, error) {
	c.fillUser(&req.UserID)
	var out AlertResponse
	if err := c.postAuthenticated(ctx, EndpointCancelAlert, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) PendingAlerts(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointPendingAlert, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) EnabledAlertTypes(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointEnabledAlertTypes, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}
