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
	lunaToBtc types.Pair
	interval time.Duration
}

var _ types.Source = &UpbitSource {}

func NewUpbitSource () *UpbitSource  {
	return &UpbitSource {
		lunaToBtc: types.NewPair(types.LUNA, types.BTC),
		interval:  time.Second * 5,
	}
}

func (source UpbitSource) Market() string {
	return "upbit"
}

func (source UpbitSource) Pair() types.Pair {
	return source.lunaToBtc
}

func (source *UpbitSource) Fetch() error {
	resp, err := http.Get("https://crix-api-cdn.upbit.com/v1/crix/trades/ticks?code=CRIX.UPBIT.BTC-LUNA&count=1")
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
	source.lunaToBtc = source.lunaToBtc.SetPrice(decAmount)

	return nil
}

func (source UpbitSource) Interval() time.Duration {
	return source.interval
}
