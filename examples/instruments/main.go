package main

import (
	"fmt"

	"github.com/suyotech/flattradeapi-go/instruments"
)

func main() {
	items, err := instruments.CheckInstruments("./insts")
	if err != nil {
		panic(err)
	}

	filter := instruments.Filter{
		Exchange:   "NFO",
		Symbol:     "NIFTY",
		Instrument: "OPTIDX",
	}

	matches := instruments.FindInstrument(items, filter)
	expiries := instruments.GetExpiry(items, filter)
	strikes := instruments.GetOptionStrike(items, filter, 25000, 5)

	fmt.Println("matches:", len(matches))
	fmt.Println("expiries:", expiries)
	fmt.Printf("strikes: %+v\n", strikes)
}
