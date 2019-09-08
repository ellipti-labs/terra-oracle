package price

import (
	"github.com/everett-protocol/terra-oracle/types"
	"time"
)

type SourceMetaSet []*SourceMeta

type SourceMeta struct {
	Source types.Source
	Weight uint64

	lastFetchTimestamp time.Time
}
