package coinbase

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/everett-protocol/terra-oracle/types"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type CoinbaseSource struct {
	btcToUsd types.Pair
	interval time.Duration
}

var _ types.Source = &CoinbaseSource{}

func SourceProvider() (string, string, types.SourceProvider) {
	return types.PairStr(types.BTC, types.USD), "coinbase", NewCoinbaseSource
}

func NewCoinbaseSource(interval time.Duration) types.Source {
	return &CoinbaseSource{
		btcToUsd: types.NewPair(types.BTC, types.USD),
		interval: interval,
	}
}

func (source CoinbaseSource) Market() string {
	return "coinbase"
}

func (source CoinbaseSource) Pair() types.Pair {
	return source.btcToUsd
}

func (source *CoinbaseSource) Fetch() error {
	resp, err := http.Get("https://api.coinbase.com/v2/prices/BTC-USD/spot")
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

	re, _ := regexp.Compile("\"amount\":\"[0-9.]+\"")
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

func (source CoinbaseSource) Interval() time.Duration {
	return source.interval
}
