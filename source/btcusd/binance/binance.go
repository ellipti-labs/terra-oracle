package binance

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/everett-protocol/terra-oracle/types"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type BinanceSource struct {
	btcToUsd types.Pair
	interval time.Duration
}

var _ types.Source = &BinanceSource{}

func SourceProvider() (string, string, types.SourceProvider) {
	return types.PairStr(types.BTC, types.USD), "binance", NewBinanceSource
}

func NewBinanceSource(interval time.Duration) types.Source {
	return &BinanceSource{
		btcToUsd: types.NewPair(types.BTC, types.USD),
		interval: interval,
	}
}

func (source BinanceSource) Market() string {
	return "binance"
}

func (source BinanceSource) Pair() types.Pair {
	return source.btcToUsd
}

func (source *BinanceSource) Fetch() error {
	resp, err := http.Get("https://www.binance.com/api/v1/aggTrades?limit=1&symbol=BTCUSDT")
	if err != nil {
		return err
	}
	defer func() {
		resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	re, _ := regexp.Compile("\"p\":\"[0-9.]+\"")
	str := re.FindString(string(body))
	re, _ = regexp.Compile("[0-9.]+")
	price := re.FindString(str)

	decAmount, err := sdk.NewDecFromStr(price)
	if err != nil {
		return err
	}
	source.btcToUsd = source.btcToUsd.SetPrice(decAmount)

	return nil
}

func (source BinanceSource) Interval() time.Duration {
	return source.interval
}
