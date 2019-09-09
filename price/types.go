package price

import (
	"fmt"
	"github.com/tendermint/tendermint/libs/log"
	"sync/atomic"
	"time"

	"github.com/everett-protocol/terra-oracle/types"
)

type SourceMetaSet []*SourceMeta

type SourceMeta struct {
	Source types.Source
	Weight uint64

	isFetching         int32
	lastFetchTimestamp time.Time
}

func (meta *SourceMeta) Fetch(logger log.Logger) {
	now := time.Now()
	if meta.lastFetchTimestamp.IsZero() || meta.lastFetchTimestamp.Add(meta.Source.Interval()).Before(now) {
		if atomic.LoadInt32(&meta.isFetching) != 0 {
			return
		}
		atomic.StoreInt32(&meta.isFetching, 1)
		defer func() {
			atomic.StoreInt32(&meta.isFetching, 0)
		}()

		meta.lastFetchTimestamp = now

		source := meta.Source
		err := source.Fetch()
		if err != nil {
			logger.Error(fmt.Sprintf("Fail to fetch %s from %s: %s", source.Pair(), source.Market(), err.Error()))
		}
		price, err := source.Pair().Price()
		if err != nil {
			logger.Error(fmt.Sprintf("Fail to parse price %s from %s: %s", source.Pair(), source.Market(), err.Error()))
		}
		logger.Info(fmt.Sprintf("Fetch %s from %s: %s", source.Pair(), source.Market(), price), "market", source.Market(), "pair", source.Pair())
	}
}
