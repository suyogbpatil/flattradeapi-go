package api

import "context"

type SearchScripRequest struct {
	UserID     string `json:"uid,omitempty"`
	Exchange   string `json:"exch,omitempty"`
	SearchText string `json:"stext,omitempty"`
}

type QuoteRequest struct {
	UserID   string `json:"uid,omitempty"`
	Exchange string `json:"exch,omitempty"`
	Token    string `json:"token,omitempty"`
}

func (c *Client) SearchScrip(ctx context.Context, req SearchScripRequest) ([]map[string]any, error) {
	c.fillUser(&req.UserID)
	var out []map[string]any
	if err := c.postAuthenticated(ctx, EndpointSearchScrip, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetQuotes(ctx context.Context, req QuoteRequest) (*Quote, error) {
	c.fillUser(&req.UserID)
	var out Quote
	if err := c.postAuthenticated(ctx, EndpointGetQuotes, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
