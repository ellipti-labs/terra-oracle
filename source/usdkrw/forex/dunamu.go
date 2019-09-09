package forex

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/everett-protocol/terra-oracle/types"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type ForexDonamuSource struct {
	usdToKrw types.Pair
	interval time.Duration
}

var _ types.Source = &ForexDonamuSource{}

func NewForexDonamuSource() *ForexDonamuSource {
	return &ForexDonamuSource{
		usdToKrw: types.NewPair(types.USD, types.KRW),
		interval: time.Minute * 10,
	}
}

func (source ForexDonamuSource) Market() string {
	return "forex-dunamu-api"
}

func (source ForexDonamuSource) Pair() types.Pair {
	return source.usdToKrw
}

func (source *ForexDonamuSource) Fetch() error {
	resp, err := http.Get("https://quotation-api-cdn.dunamu.com/v1/forex/recent?codes=FRX.KRWUSD")
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

	re, _ := regexp.Compile("\"basePrice\":[0-9.]+")
	str := re.FindString(string(body))
	re, _ = regexp.Compile("[0-9.]+")
	price := re.FindString(str)

	decAmount, err := sdk.NewDecFromStr(price)
	if err != nil {
		return err
	}
	source.usdToKrw = source.usdToKrw.SetPrice(decAmount)

	return nil
}

func (source ForexDonamuSource) Interval() time.Duration {
	return source.interval
}
