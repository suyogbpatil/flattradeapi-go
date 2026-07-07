package instruments

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const DefaultDir = "instruments"
const baseURL = "https://flattrade.s3.ap-south-1.amazonaws.com/scripmaster"

var defaultFiles = map[string]string{
	"NSE":        baseURL + "/NSE_Equity.csv",
	"BSE":        baseURL + "/BSE_Equity.csv",
	"NFO_EQUITY": baseURL + "/Nfo_Equity_Derivatives.csv",
	"NFO_INDEX":  baseURL + "/Nfo_Index_Derivatives.csv",
	"BFO_EQUITY": baseURL + "/Bfo_Equity_Derivatives.csv",
	"BFO_INDEX":  baseURL + "/Bfo_Index_Derivatives.csv",
	"CURRENCY":   baseURL + "/Currency_Derivatives.csv",
	"COMMODITY":  baseURL + "/Commodity.csv",
}

type Instrument struct {
	Exchange      string
	Token         string
	LotSize       string
	Symbol        string
	TradingSymbol string
	Expiry        string
	Instrument    string
	OptionType    string
	Strike        string
	TickSize      string
}

type Filter struct {
	Exchange      string
	Symbol        string
	TradingSymbol string
	Token         string
	Expiry        string
	Instrument    string
	OptionType    string
	Strike        string
}

type OptionStrikes struct {
	ATM float64
	Up  []float64
	Dn  []float64
}

func CheckInstruments(dir string) ([]Instrument, error) {
	if dir == "" {
		dir = DefaultDir
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	download := false
	for name := range defaultFiles {
		ok, err := downloadedAfterMorning(filepath.Join(dir, fileName(name)), time.Now())
		if err != nil {
			return nil, err
		}
		if !ok {
			download = true
			break
		}
	}

	if download {
		if err := downloadAll(dir); err != nil {
			return nil, err
		}
	}

	return LoadAll(dir)
}

func FindInstrument(items []Instrument, filter Filter) []Instrument {
	out := make([]Instrument, 0)
	for _, item := range items {
		if !matches(item, filter) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func GetExpiry(items []Instrument, filter Filter) []string {
	seen := make(map[string]struct{})
	for _, item := range FindInstrument(items, filter) {
		if item.Expiry == "" {
			continue
		}
		seen[item.Expiry] = struct{}{}
	}

	out := make([]string, 0, len(seen))
	for expiry := range seen {
		out = append(out, expiry)
	}
	sort.Strings(out)
	return out
}

func GetOptionStrike(items []Instrument, filter Filter, price float64, count int) OptionStrikes {
	strikes := uniqueStrikes(FindInstrument(items, filter))
	if len(strikes) == 0 {
		return OptionStrikes{}
	}

	atmIndex := nearestStrikeIndex(strikes, price)
	out := OptionStrikes{ATM: strikes[atmIndex]}

	for i := atmIndex + 1; i < len(strikes) && len(out.Up) < count; i++ {
		out.Up = append(out.Up, strikes[i])
	}
	for i := atmIndex - 1; i >= 0 && len(out.Dn) < count; i-- {
		out.Dn = append(out.Dn, strikes[i])
	}

	return out
}

func LoadAll(dir string) ([]Instrument, error) {
	if dir == "" {
		dir = DefaultDir
	}

	all := make([]Instrument, 0)
	for name := range defaultFiles {
		file, err := os.Open(filepath.Join(dir, fileName(name)))
		if err != nil {
			return nil, err
		}

		items, readErr := ReadCSV(file)
		closeErr := file.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}

		all = append(all, items...)
	}
	return all, nil
}

func ReadCSV(r io.Reader) ([]Instrument, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}

	header := indexHeader(records[0])
	items := make([]Instrument, 0, len(records)-1)
	for _, record := range records[1:] {
		items = append(items, Instrument{
			Exchange:      value(record, header, "exch", "exchange"),
			Token:         value(record, header, "token"),
			LotSize:       value(record, header, "lotsize", "lot_size", "ls"),
			Symbol:        value(record, header, "symbol", "sym", "cname"),
			TradingSymbol: value(record, header, "trading_symbol", "tradingsymbol", "tsym"),
			Expiry:        value(record, header, "expiry", "exp", "exd"),
			Instrument:    value(record, header, "instrument", "instname", "instnam"),
			OptionType:    value(record, header, "option_type", "optt", "opttype"),
			Strike:        value(record, header, "strike", "strprc"),
			TickSize:      value(record, header, "tick_size", "ti"),
		})
	}

	return items, nil
}

func Match(items []Instrument, filter Filter) []Instrument {
	return FindInstrument(items, filter)
}

func downloadAll(dir string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	for name, url := range defaultFiles {
		if err := downloadFile(client, url, filepath.Join(dir, fileName(name))); err != nil {
			return err
		}
	}
	return nil
}

func fileName(name string) string {
	return name + ".csv"
}

func downloadFile(client *http.Client, url string, path string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("download failed: " + url + ": " + resp.Status)
	}

	tmp := path + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(file, resp.Body)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}

	return os.Rename(tmp, path)
}

func downloadedAfterMorning(path string, now time.Time) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	morning := time.Date(now.Year(), now.Month(), now.Day(), 8, 30, 0, 0, now.Location())
	return info.ModTime().After(morning), nil
}

func matches(item Instrument, filter Filter) bool {
	return same(filter.Exchange, item.Exchange) &&
		same(filter.Symbol, item.Symbol) &&
		same(filter.TradingSymbol, item.TradingSymbol) &&
		same(filter.Token, item.Token) &&
		same(filter.Expiry, item.Expiry) &&
		same(filter.Instrument, item.Instrument) &&
		same(filter.OptionType, item.OptionType) &&
		same(filter.Strike, item.Strike)
}

func same(filterValue string, value string) bool {
	return filterValue == "" || strings.EqualFold(filterValue, value)
}

func uniqueStrikes(items []Instrument) []float64 {
	seen := make(map[float64]struct{})
	for _, item := range items {
		if item.Strike == "" {
			continue
		}
		strike, err := strconv.ParseFloat(item.Strike, 64)
		if err != nil {
			continue
		}
		seen[strike] = struct{}{}
	}

	out := make([]float64, 0, len(seen))
	for strike := range seen {
		out = append(out, strike)
	}
	sort.Float64s(out)
	return out
}

func nearestStrikeIndex(strikes []float64, price float64) int {
	idx := sort.SearchFloat64s(strikes, price)
	if idx == 0 {
		return 0
	}
	if idx >= len(strikes) {
		return len(strikes) - 1
	}
	if price-strikes[idx-1] <= strikes[idx]-price {
		return idx - 1
	}
	return idx
}

func indexHeader(record []string) map[string]int {
	header := make(map[string]int, len(record))
	for i, name := range record {
		header[normalize(name)] = i
	}
	return header
}

func value(record []string, header map[string]int, names ...string) string {
	for _, name := range names {
		i, ok := header[normalize(name)]
		if ok && i < len(record) {
			return strings.TrimSpace(record[i])
		}
	}
	return ""
}

func normalize(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}
