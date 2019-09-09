package types

import (
	"time"
)

// Source fetches pair and sets price from market every {interval}.
type Source interface {
	Market() string
	Pair() Pair
	Fetch() error
	Interval() time.Duration
}

type SourceProvider func(interval time.Duration) Source
