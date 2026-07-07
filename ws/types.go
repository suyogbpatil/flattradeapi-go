package ws

import "encoding/json"

type Event string

const (
	EventConnect                      Event = "a"
	EventConnectAck                   Event = "ak"
	EventTouchline                    Event = "t"
	EventTouchlineAck                 Event = "tk"
	EventTouchlineFeed                Event = "tf"
	EventDepth                        Event = "d"
	EventDepthAck                     Event = "dk"
	EventDepthFeed                    Event = "df"
	EventOrder                        Event = "o"
	EventOrderFeed                    Event = "om"
	EventHeartbeat                    Event = "h"
	EventHeartbeatAck                 Event = "hk"
	EventSubscribeTouchline           Event = "t"
	EventUnsubscribeTouchline         Event = "uk"
	EventSubscribeDepth               Event = "d"
	EventUnsubscribeDepth             Event = "ud"
	EventUnsubscribeDepthAck          Event = "udk"
	EventSubscribeOrderUpdate         Event = "o"
	EventUnsubscribeOrderUpdate       Event = "uo"
	EventUnsubscribeOrderUpdateAck    Event = "uok"
	EventSubscribePositionUpdate      Event = "p"
	EventSubscribePositionUpdateAck   Event = "pk"
	EventPositionFeed                 Event = "pm"
	EventUnsubscribePositionUpdate    Event = "up"
	EventUnsubscribePositionUpdateAck Event = "upk"
)

type LoginMessage struct {
	Event       Event  `json:"t"`
	UserID      string `json:"uid"`
	AccountID   string `json:"actid,omitempty"`
	AccessToken string `json:"accesstoken"`
	Source      string `json:"source"`
}

type SubscribeMessage struct {
	Event Event  `json:"t"`
	Key   string `json:"k,omitempty"`
}

type AccountMessage struct {
	Event     Event  `json:"t"`
	AccountID string `json:"actid,omitempty"`
}

type HeartbeatMessage struct {
	Event Event `json:"t"`
}

type ConnectAck struct {
	Event  Event  `json:"t,omitempty"`
	Status string `json:"s,omitempty"`
}

type Message struct {
	Event Event          `json:"t,omitempty"`
	Raw   map[string]any `json:"-"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	type message Message
	var out message
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*m = Message(out)
	m.Raw = raw
	return nil
}

type Ack struct {
	Event     Event          `json:"t,omitempty"`
	Exchange  string         `json:"e,omitempty"`
	Token     string         `json:"tk,omitempty"`
	Key       string         `json:"k,omitempty"`
	UserID    string         `json:"uid,omitempty"`
	AccountID string         `json:"actid,omitempty"`
	Status    string         `json:"s,omitempty"`
	Raw       map[string]any `json:"-"`
}

func (a *Ack) UnmarshalJSON(data []byte) error {
	type ack Ack
	var out ack
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*a = Ack(out)
	a.Raw = raw
	return nil
}

type Tick struct {
	Event                Event          `json:"t,omitempty"`
	Exchange             string         `json:"e,omitempty"`
	Token                string         `json:"tk,omitempty"`
	TradingSymbol        string         `json:"ts,omitempty"`
	PricePrecision       string         `json:"pp,omitempty"`
	TickSize             string         `json:"ti,omitempty"`
	LotSize              string         `json:"ls,omitempty"`
	LastPrice            string         `json:"lp,omitempty"`
	PercentChange        string         `json:"pc,omitempty"`
	Volume               string         `json:"v,omitempty"`
	Open                 string         `json:"o,omitempty"`
	High                 string         `json:"h,omitempty"`
	Low                  string         `json:"l,omitempty"`
	Close                string         `json:"c,omitempty"`
	ClosePrice           string         `json:"cp,omitempty"`
	AveragePrice         string         `json:"ap,omitempty"`
	LastTradeTime        string         `json:"ltt,omitempty"`
	LastTradeQty         string         `json:"ltq,omitempty"`
	OpenInterest         string         `json:"oi,omitempty"`
	PreviousOpenInterest string         `json:"poi,omitempty"`
	TotalOpenInterest    string         `json:"toi,omitempty"`
	TotalBuyQty          string         `json:"tbq,omitempty"`
	TotalSellQty         string         `json:"tsq,omitempty"`
	FeedTime             string         `json:"ft,omitempty"`
	LowerCircuit         string         `json:"lc,omitempty"`
	UpperCircuit         string         `json:"uc,omitempty"`
	Week52High           string         `json:"52h,omitempty"`
	Week52Low            string         `json:"52l,omitempty"`
	UpperExchangeRange   string         `json:"ue,omitempty"`
	LowerExchangeRange   string         `json:"le,omitempty"`
	BestBuyQty1          string         `json:"bq1,omitempty"`
	BestBuyQty2          string         `json:"bq2,omitempty"`
	BestBuyQty3          string         `json:"bq3,omitempty"`
	BestBuyQty4          string         `json:"bq4,omitempty"`
	BestBuyQty5          string         `json:"bq5,omitempty"`
	BestBuyPrice1        string         `json:"bp1,omitempty"`
	BestBuyPrice2        string         `json:"bp2,omitempty"`
	BestBuyPrice3        string         `json:"bp3,omitempty"`
	BestBuyPrice4        string         `json:"bp4,omitempty"`
	BestBuyPrice5        string         `json:"bp5,omitempty"`
	BestBuyOrders1       string         `json:"bo1,omitempty"`
	BestBuyOrders2       string         `json:"bo2,omitempty"`
	BestBuyOrders3       string         `json:"bo3,omitempty"`
	BestBuyOrders4       string         `json:"bo4,omitempty"`
	BestBuyOrders5       string         `json:"bo5,omitempty"`
	BestSellQty1         string         `json:"sq1,omitempty"`
	BestSellQty2         string         `json:"sq2,omitempty"`
	BestSellQty3         string         `json:"sq3,omitempty"`
	BestSellQty4         string         `json:"sq4,omitempty"`
	BestSellQty5         string         `json:"sq5,omitempty"`
	BestSellPrice1       string         `json:"sp1,omitempty"`
	BestSellPrice2       string         `json:"sp2,omitempty"`
	BestSellPrice3       string         `json:"sp3,omitempty"`
	BestSellPrice4       string         `json:"sp4,omitempty"`
	BestSellPrice5       string         `json:"sp5,omitempty"`
	BestSellOrders1      string         `json:"so1,omitempty"`
	BestSellOrders2      string         `json:"so2,omitempty"`
	BestSellOrders3      string         `json:"so3,omitempty"`
	BestSellOrders4      string         `json:"so4,omitempty"`
	BestSellOrders5      string         `json:"so5,omitempty"`
	Raw                  map[string]any `json:"-"`
}

type OrderUpdate struct {
	Event           Event          `json:"t,omitempty"`
	OrderNo         string         `json:"norenordno,omitempty"`
	UserID          string         `json:"uid,omitempty"`
	AccountID       string         `json:"actid,omitempty"`
	Exchange        string         `json:"exch,omitempty"`
	TradingSymbol   string         `json:"tsym,omitempty"`
	Quantity        string         `json:"qty,omitempty"`
	Price           string         `json:"prc,omitempty"`
	Product         string         `json:"pcode,omitempty"`
	Status          string         `json:"status,omitempty"`
	ReportType      string         `json:"reporttype,omitempty"`
	TransactionType string         `json:"trantype,omitempty"`
	PriceType       string         `json:"prctyp,omitempty"`
	Retention       string         `json:"ret,omitempty"`
	FilledShares    string         `json:"fillshares,omitempty"`
	AveragePrice    string         `json:"avgprc,omitempty"`
	FillTime        string         `json:"fltm,omitempty"`
	FillID          string         `json:"flid,omitempty"`
	FillQty         string         `json:"flqty,omitempty"`
	FillPrice       string         `json:"flprc,omitempty"`
	RejectReason    string         `json:"rejreason,omitempty"`
	ExchangeOrderID string         `json:"exchordid,omitempty"`
	CancelQty       string         `json:"cancelqty,omitempty"`
	Remarks         string         `json:"remarks,omitempty"`
	DisclosedQty    string         `json:"dscqty,omitempty"`
	TriggerPrice    string         `json:"trgprc,omitempty"`
	ExchangeTime    string         `json:"exch_tm,omitempty"`
	Time            string         `json:"tm,omitempty"`
	NanoTime        string         `json:"ntm,omitempty"`
	Raw             map[string]any `json:"-"`
}

func (o *OrderUpdate) UnmarshalJSON(data []byte) error {
	type orderUpdate OrderUpdate
	var out orderUpdate
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*o = OrderUpdate(out)
	o.Raw = raw
	return nil
}

type PositionUpdate struct {
	Event             Event          `json:"t,omitempty"`
	Exchange          string         `json:"exch,omitempty"`
	Token             string         `json:"token,omitempty"`
	UserID            string         `json:"uid,omitempty"`
	AccountID         string         `json:"actid,omitempty"`
	Product           string         `json:"prd,omitempty"`
	DayBuyQty         string         `json:"daybuyqty,omitempty"`
	DaySellQty        string         `json:"daysellqty,omitempty"`
	DayBuyAmt         string         `json:"daybuyamt,omitempty"`
	DaySellAmt        string         `json:"daysellamt,omitempty"`
	CFBuyQty          string         `json:"cfbuyqty,omitempty"`
	CFSellQty         string         `json:"cfsellqty,omitempty"`
	CFBuyAmt          string         `json:"cfbuyamt,omitempty"`
	CFSellAmt         string         `json:"cfsellamt,omitempty"`
	OpenBuyQty        string         `json:"openbuyqty,omitempty"`
	OpenSellQty       string         `json:"opensellqty,omitempty"`
	OpenBuyAmt        string         `json:"openbuyamt,omitempty"`
	OpenSellAmt       string         `json:"opensellamt,omitempty"`
	Instrument        string         `json:"instname,omitempty"`
	UploadPrice       string         `json:"upload_prc,omitempty"`
	BuyAvgPrice       string         `json:"buyavgprc,omitempty"`
	SellAvgPrice      string         `json:"sellavgprc,omitempty"`
	RealizedPNL       string         `json:"rpnl,omitempty"`
	NetQty            string         `json:"netqty,omitempty"`
	TotalBuyAmt       string         `json:"totbuyamt,omitempty"`
	TotalSellAmt      string         `json:"totsellamt,omitempty"`
	TotalBuyAvgPrice  string         `json:"totbuyavgprc,omitempty"`
	TotalSellAvgPrice string         `json:"totsellavgprc,omitempty"`
	Raw               map[string]any `json:"-"`
}

func (p *PositionUpdate) UnmarshalJSON(data []byte) error {
	type positionUpdate PositionUpdate
	var out positionUpdate
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*p = PositionUpdate(out)
	p.Raw = raw
	return nil
}

func (t *Tick) UnmarshalJSON(data []byte) error {
	type tick Tick
	var out tick
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*t = Tick(out)
	t.Raw = raw
	return nil
}
