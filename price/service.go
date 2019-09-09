package price

import (
	"fmt"
	"sync"

	cmn "github.com/tendermint/tendermint/libs/common"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/everett-protocol/terra-oracle/types"
)

// Price service fetches the sources periodically.
type PriceService struct {
	cmn.BaseService
	mutex  *sync.RWMutex
	sourceMetas map[string]SourceMetaSet

	currencies []string
}

func NewPriceService() *PriceService {
	ps := &PriceService{
		mutex:  new(sync.RWMutex),
		sourceMetas: make(map[string]SourceMetaSet),

		currencies: []string{types.BTC, types.KRW, types.LUNA, types.USD, types.SDR},
	}
	ps.BaseService = *cmn.NewBaseService(nil, "PriceService", ps)
	return ps
}

func (ps PriceService) OnStart() error {
	// TODO: gracefully quit go routine
	go ps.startRoutine()
	return nil
}

// Return pair price, but if it doesn't exist, return reverse pair price.
func (ps PriceService) GetPrice(base string, quote string) (sdk.Dec, uint64, error) {
	price, weight, err := ps.GetPriceExact(base, quote)
	if err == nil {
		return price, weight, nil
	}

	price, weight, err = ps.GetPriceExact(quote, base)
	if err == nil {
		return sdk.OneDec().Quo(price), weight, nil
	}

	return sdk.Dec{}, 0, err
}

// Return exact base/quote pair price.
func (ps PriceService) GetPriceExact(base string, quote string) (sdk.Dec, uint64, error) {
	metaset, ok := ps.sourceMetas[types.PairStr(base, quote)]
	if ok == false {
		return sdk.Dec{}, 0, fmt.Errorf("can't get source metadata set")
	}
	if len(metaset) == 0 {
		return sdk.Dec{}, 0, fmt.Errorf("invalid metadata set")
	}

	sum := sdk.ZeroDec()
	weightSum := uint64(0)
	for _, meta := range metaset {
		price, err := meta.Source.Pair().Price()
		if err != nil {
			ps.Logger.Error(fmt.Sprintf("Error when getting price %s: %s", meta.Source.Pair(), err.Error()))
			continue
		}
		sum = sum.Add(price)
		weightSum += meta.Weight
	}

	if weightSum == 0 {
		return sdk.Dec{}, 0, fmt.Errorf("can't calculate weighted mean")
	}

	return sum.QuoInt64(int64(weightSum)), weightSum, nil
}

func (ps PriceService) PushSourceMeta(sourceMeta SourceMeta) {
	if sourceMeta.Source == nil {
		panic(fmt.Errorf("source not set"))
	}
	if sourceMeta.Weight == 0 {
		panic(fmt.Errorf("weight should not be zero"))
	}

	metaset, ok := ps.sourceMetas[sourceMeta.Source.Pair().String()]
	if ok == false {
		metaset = make(SourceMetaSet, 0)
	}
	ps.sourceMetas[sourceMeta.Source.Pair().String()] = append(metaset, &sourceMeta)
}
