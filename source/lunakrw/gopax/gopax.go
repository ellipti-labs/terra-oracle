package gopax

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/everett-protocol/terra-oracle/types"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type GopaxSource struct {
	lunaToKrw types.Pair
	interval time.Duration
}

var _ types.Source = &GopaxSource{}

func SourceProvider() (string, string, types.SourceProvider) {
	return types.PairStr(types.LUNA, types.KRW), "gopax", NewGopaxSource
}

func NewGopaxSource(interval time.Duration) types.Source {
	return &GopaxSource{
		lunaToKrw: types.NewPair(types.LUNA, types.KRW),
		interval: interval,
	}
}

func (source GopaxSource) Market() string {
	return "gopax"
}

func (source GopaxSource) Pair() types.Pair {
	return source.lunaToKrw
}

func (source *GopaxSource) Fetch() error {
	resp, err := http.Get("https://api.gopax.co.kr/trading-pairs/LUNA-KRW/ticker")
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

	re, _ := regexp.Compile("\"price\":[0-9.]+")
	str := re.FindString(string(body))
	re, _ = regexp.Compile("[0-9.]+")
	price := re.FindString(str)

	decAmount, err := sdk.NewDecFromStr(price)
	if err != nil {
		return err
	}
	source.lunaToKrw = source.lunaToKrw.SetPrice(decAmount)

	return nil
}

func (source GopaxSource) Interval() time.Duration {
	return source.interval
}
