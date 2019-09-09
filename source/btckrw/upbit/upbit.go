package upbit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/everett-protocol/terra-oracle/types"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type UpbitSource struct {
	btcToKrw types.Pair
	interval time.Duration
}

var _ types.Source = &UpbitSource{}

func SourceProvider() (string, string, types.SourceProvider) {
	return types.PairStr(types.BTC, types.KRW), "upbit", NewUpbitSource
}

func NewUpbitSource(interval time.Duration) types.Source {
	return &UpbitSource{
		btcToKrw: types.NewPair(types.BTC, types.KRW),
		interval: interval,
	}
}

func (source UpbitSource) Market() string {
	return "upbit"
}

func (source UpbitSource) Pair() types.Pair {
	return source.btcToKrw
}

func (source *UpbitSource) Fetch() error {
	resp, err := http.Get("https://crix-api-cdn.upbit.com/v1/crix/trades/ticks?code=CRIX.UPBIT.KRW-BTC&count=1")
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

	re, _ := regexp.Compile("\"tradePrice\":[0-9.]+")
	str := re.FindString(string(body))
	re, _ = regexp.Compile("[0-9.]+")
	price := re.FindString(str)

	decAmount, err := sdk.NewDecFromStr(price)
	if err != nil {
		return err
	}
	source.btcToKrw = source.btcToKrw.SetPrice(decAmount)

	return nil
}

func (source UpbitSource) Interval() time.Duration {
	return source.interval
}
