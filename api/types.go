package api

type UserDetails struct {
	UserID    string `json:"uid,omitempty"`
	AccountID string `json:"actid,omitempty"`
	Name      string `json:"uname,omitempty"`
	Email     string `json:"email,omitempty"`
	Mobile    string `json:"m_num,omitempty"`
}

type Order struct {
	Exchange        string `json:"exch,omitempty"`
	Token           string `json:"token,omitempty"`
	TradingSymbol   string `json:"tsym,omitempty"`
	Quantity        string `json:"qty,omitempty"`
	Price           string `json:"prc,omitempty"`
	Product         string `json:"prd,omitempty"`
	TransactionType string `json:"trantype,omitempty"`
	PriceType       string `json:"prctyp,omitempty"`
	Retention       string `json:"ret,omitempty"`
}

type Quote struct {
	Exchange      string `json:"exch,omitempty"`
	Token         string `json:"token,omitempty"`
	TradingSymbol string `json:"tsym,omitempty"`
	LastPrice     string `json:"lp,omitempty"`
	Open          string `json:"o,omitempty"`
	High          string `json:"h,omitempty"`
	Low           string `json:"l,omitempty"`
	Close         string `json:"c,omitempty"`
}
