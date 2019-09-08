package price

import (
	"fmt"
	"github.com/everett-protocol/terra-oracle/types"
	"time"
)

func (ps PriceService) startRoutine() {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					ps.Logger.Error("Unknown error", r)
				}

				time.Sleep(1 * time.Second)
			}()

			now := time.Now()
			for _, sourceMetaSet := range ps.sourceMetas {
				for _, sourceMeta := range sourceMetaSet {
					source := sourceMeta.Source
					if sourceMeta.lastFetchTimestamp.IsZero() || sourceMeta.lastFetchTimestamp.Add(source.Interval()).Before(now) {
						go ps.fetchSource(source)
						sourceMeta.lastFetchTimestamp = now
						
						// Sleep a bit. It seems that logger is not thread safe.
						// Without sleep, log is often broken.
						time.Sleep(time.Millisecond * 100)
					}
				}
			}
		}()
	}
}

func (ps PriceService) fetchSource(source types.Source) {
	err := source.Fetch()
	if err != nil {
		ps.Logger.Error(fmt.Sprintf("Fail to fetch %s from %s: %s", source.Pair(), source.Market(), err.Error()))
	}
	price, err := source.Pair().Price()
	if err != nil {
		ps.Logger.Error(fmt.Sprintf("Fail to parse price %s from %s: %s", source.Pair(), source.Market(), err.Error()))
	}
	ps.Logger.Info(fmt.Sprintf("Fetch %s from %s: %s", source.Pair(), source.Market(), price))
}
