package price

import (
	"fmt"
	"github.com/everett-protocol/terra-oracle/types"
	"time"
)

type SourceManager struct {
	sourceProviders map[string]map[string]types.SourceProvider
}

func NewSourceManager() *SourceManager {
	return &SourceManager{
		sourceProviders: make(map[string]map[string]types.SourceProvider),
	}
}

func (manager *SourceManager) Register(pair string, name string, provider types.SourceProvider) {
	if manager.sourceProviders[pair] == nil {
		manager.sourceProviders[pair] = make(map[string]types.SourceProvider)
	}
	manager.sourceProviders[pair][name] = provider
}

func (manager *SourceManager) GetSourceMeta(pair string, name string, interval time.Duration, weight uint64) SourceMeta {
	sourceProvider := manager.sourceProviders[pair][name]
	source := sourceProvider(interval)
	if pair != source.Pair().String() {
		panic(fmt.Errorf("invalid source expected pair: %s, actual: %s", pair, source.Pair().String()))
	}
	return SourceMeta{
		Source: source,
		Weight: weight,
	}
}
