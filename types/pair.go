package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Pair indicates how many {Quote} currency are needed to purchase one {Base} currency.
// And, pair should be immutable. So, if you want to set value in pair, you should return new structure.
type Pair interface {
	String() string
	Base() string
	Quote() string

	Price() (sdk.Dec, error)
	SetPrice(sdk.Dec) Pair

	// Maybe these can be used in future?
	// func (func Pair) Volume() (sdk.Dec, error)
	// func (func Pair) Bid() (sdk.Dec, error)
	// func (func Pair) Ask() (sdk.Dec, error)
	// func (func Pair) Spread() (sdk.Dec, error)
}

type BasePair struct {
	base string
	quote string
	price sdk.Dec
}

var _ Pair = BasePair{}

func NewPair(base string, quote string) BasePair {
	return BasePair {
		base: base,
		quote: quote,
		price: sdk.ZeroDec(),
	}
}

func (pair BasePair) String() string {
	return pair.base + "/" + pair.quote
}

func (pair BasePair) Base() string {
	return pair.base
}

func (pair BasePair) Quote() string {
	return pair.quote
}

func (pair BasePair) Price() (sdk.Dec, error) {
	if pair.price.LTE(sdk.ZeroDec()) {
		return sdk.ZeroDec(), fmt.Errorf("invalid price")
	}
	return pair.price, nil
}

func (pair BasePair) SetPrice(price sdk.Dec) Pair {
	return BasePair {
		base:  pair.base,
		quote: pair.quote,
		price: price,
	}
}
