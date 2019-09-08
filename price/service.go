package price

import (
	"fmt"
	"sync"

	cmn "github.com/tendermint/tendermint/libs/common"
)

// Price service fetches the sources periodically.
type PriceService struct {
	cmn.BaseService
	mutex  *sync.RWMutex
	sourceMetas map[string]SourceMetaSet
}

func NewPriceService() *PriceService {
	ps := &PriceService{
		mutex:  new(sync.RWMutex),
		sourceMetas: make(map[string]SourceMetaSet),
	}
	ps.BaseService = *cmn.NewBaseService(nil, "PriceService", ps)
	return ps
}

func (ps PriceService) OnStart() error {
	// TODO: gracefully quit go routine
	go ps.startRoutine()
	return nil
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
