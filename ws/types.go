package ws

import "encoding/json"

type Event string

const (
	EventConnect                   Event = "c"
	EventTouchline                 Event = "t"
	EventDepth                     Event = "d"
	EventOrder                     Event = "o"
	EventHeartbeat                 Event = "h"
	EventSubscribeTouchline        Event = "t"
	EventUnsubscribeTouchline      Event = "u"
	EventSubscribeDepth            Event = "d"
	EventUnsubscribeDepth          Event = "ud"
	EventSubscribeOrderUpdate      Event = "o"
	EventUnsubscribeOrderUpdate    Event = "uo"
	EventSubscribePositionUpdate   Event = "p"
	EventUnsubscribePositionUpdate Event = "up"
)

type LoginMessage struct {
	Event       Event  `json:"t"`
	UserID      string `json:"uid"`
	AccountID   string `json:"actid,omitempty"`
	AccessToken string `json:"susertoken"`
	Source      string `json:"source"`
}

type SubscribeMessage struct {
	Event  Event    `json:"t"`
	Tokens []string `json:"k,omitempty"`
}

type HeartbeatMessage struct {
	Event Event `json:"t"`
}

type Tick struct {
	Event         Event  `json:"t,omitempty"`
	Exchange      string `json:"e,omitempty"`
	Token         string `json:"tk,omitempty"`
	TradingSymbol string `json:"ts,omitempty"`
	LastPrice     string `json:"lp,omitempty"`
	Volume        string `json:"v,omitempty"`
	Raw           map[string]any `json:"-"`
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
