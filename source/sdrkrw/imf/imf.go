package imf

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/everett-protocol/terra-oracle/types"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type IMFSource struct {
	sdrToKrw types.Pair
	interval time.Duration
}

var _ types.Source = &IMFSource{}

func NewIMFSource() *IMFSource {
	return &IMFSource{
		sdrToKrw: types.NewPair(types.SDR, types.KRW),
		interval: time.Minute * 30,
	}
}

func (source IMFSource) Market() string {
	return "imf"
}

func (source IMFSource) Pair() types.Pair {
	return source.sdrToKrw
}

func (source *IMFSource) Fetch() error {
	resp, err := http.Get("https://www.imf.org/external/np/fin/data/rms_five.aspx?tsvflag=Y")
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

	re, _ := regexp.Compile("Korean won[\\s]+[0-9.,]+")
	strs := re.FindAllString(string(body), 2)
	if len(strs) < 2 {
		return fmt.Errorf("fail to find sdr-won")
	}
	re, _ = regexp.Compile("[0-9.,]+")
	price := re.FindString(strs[1])
	price = strings.ReplaceAll(price, ",", "")

	decAmount, err := sdk.NewDecFromStr(price)
	if err != nil {
		return err
	}
	source.sdrToKrw = source.sdrToKrw.SetPrice(decAmount)

	return nil
}

func (source IMFSource) Interval() time.Duration {
	return source.interval
}
