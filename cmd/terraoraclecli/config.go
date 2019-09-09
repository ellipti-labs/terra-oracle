package main

import (
	"time"
)

type OracleConfig struct {
	BtcUsd  []sourceConfig `mapstructure:"btc-usd"`
	BtcKrw  []sourceConfig `mapstructure:"btc-krw"`
	LunaBtc []sourceConfig `mapstructure:"luna-btc"`
	LunaKrw []sourceConfig `mapstructure:"luna-krw"`
	SdrKrw  []sourceConfig `mapstructure:"sdr-krw"`
	UsdKrw  []sourceConfig `mapstructure:"usd-krw"`
}

type sourceConfig struct {
	Name     string        `mapstructure:"name"`
	Weight   uint64        `mapstructure:"weight"`
	Interval time.Duration `mapstructure:"interval"`
}

const defaultOracleConfigTemplate = `
[[btc-usd]]
name = "binance"
weight = 1
interval = "10s"

[[btc-krw]]
name = "upbit"
weight = 1
interval = "10s"

[[luna-btc]]
name = "upbit"
weight = 1
interval = "10s"

[[luna-krw]]
name = "coinone"
weight = 1
interval = "10s"

[[sdr-krw]]
name = "imf"
weight = 1
interval = "30m"

[[usd-krw]]
name = "forex-dunamu-api"
weight = 1
interval = "30m"
`
