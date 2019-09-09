package coinone

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/everett-protocol/terra-oracle/types"
	"io/ioutil"
	"net/http"
	"time"
)

type CoinoneSource struct {
	lunaToKrw types.Pair
	interval  time.Duration
}

var _ types.Source = &CoinoneSource{}

func SourceProvider() (string, string, types.SourceProvider) {
	return types.PairStr(types.LUNA, types.KRW), "coinone", NewCoinoneSource
}

func NewCoinoneSource(interval time.Duration) types.Source {
	return &CoinoneSource{
		lunaToKrw: types.NewPair(types.LUNA, types.KRW),
		interval:  interval,
	}
}

func (coinone CoinoneSource) Market() string {
	return "coinone"
}

func (coinone CoinoneSource) Pair() types.Pair {
	return coinone.lunaToKrw
}

func (coinone *CoinoneSource) Fetch() error {
	resp, err := http.Get("https://tb.coinone.co.kr/api/v1/tradehistory/recent/?market=krw&target=luna")
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

	th := TradeHistory{}
	err = json.Unmarshal(body, &th)
	if err != nil {
		return err
	}

	trades := th.Trades
	recent := trades[len(trades)-1]

	amount, ok := sdk.NewIntFromString(recent.Price)
	if !ok {
		return fmt.Errorf("fail to parse price to int")
	}

	coinone.lunaToKrw = coinone.lunaToKrw.SetPrice(sdk.NewDecFromInt(amount))
	return nil
}

func (coinone CoinoneSource) Interval() time.Duration {
	return coinone.interval
}
