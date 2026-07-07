package api

import "context"

type Holding struct {
	Exchange      string `json:"exch,omitempty"`
	Token         string `json:"token,omitempty"`
	TradingSymbol string `json:"tsym,omitempty"`
	Quantity      string `json:"holdqty,omitempty"`
	Product       string `json:"prd,omitempty"`
}

type Limits struct {
	Cash          string `json:"cash,omitempty"`
	Payin         string `json:"payin,omitempty"`
	Payout        string `json:"payout,omitempty"`
	UnclearedCash string `json:"unclearedcash,omitempty"`
	MarginUsed    string `json:"marginused,omitempty"`
}

func (c *Client) Holdings(ctx context.Context) ([]Holding, error) {
	var out []Holding
	if err := c.postAuthenticated(ctx, EndpointHoldings, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Limits(ctx context.Context) (*Limits, error) {
	var out Limits
	if err := c.postAuthenticated(ctx, EndpointLimits, c.userPayload(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
