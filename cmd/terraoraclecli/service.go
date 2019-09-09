package main

import (
	"fmt"
	"github.com/everett-protocol/terra-oracle/types"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/tendermint/go-amino"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/everett-protocol/terra-oracle/oracle"
	"github.com/everett-protocol/terra-oracle/price"

	btcUpbit "github.com/everett-protocol/terra-oracle/source/btckrw/upbit"
	"github.com/everett-protocol/terra-oracle/source/btcusd/binance"
	"github.com/everett-protocol/terra-oracle/source/lunabtc/upbit"
	"github.com/everett-protocol/terra-oracle/source/lunakrw/coinone"
	"github.com/everett-protocol/terra-oracle/source/sdrkrw/imf"
	"github.com/everett-protocol/terra-oracle/source/usdkrw/forex"
)

func svcCmd(cdc *amino.Codec) *cobra.Command {
	svcCmd := &cobra.Command{
		Use: "service",
		RunE: func(cmd *cobra.Command, args []string) error {
			ps := price.NewPriceService()
			ps.SetLogger(logger.With("module", "price"))

			sourceManager := price.NewSourceManager()
			sourceManager.Register(coinone.SourceProvider())
			sourceManager.Register(upbit.SourceProvider())
			sourceManager.Register(btcUpbit.SourceProvider())
			sourceManager.Register(forex.SourceProvider())
			sourceManager.Register(binance.SourceProvider())
			sourceManager.Register(imf.SourceProvider())

			home := viper.GetString(cli.HomeFlag)

			cfgFile := path.Join(home, "config", "oracle.toml")
			if _, err := os.Stat(cfgFile); err != nil {
				return err
			}

			v := viper.New()

			v.SetConfigFile(cfgFile)

			if err := v.ReadInConfig(); err != nil {
				return err
			}

			oracleConfig := OracleConfig{}
			err := v.Unmarshal(&oracleConfig)
			if err != nil {
				return err
			}

			for _, config := range oracleConfig.BtcUsd {
				meta := sourceManager.GetSourceMeta(types.PairStr(types.BTC, types.USD), config.Name, config.Interval, config.Weight)
				ps.PushSourceMeta(meta)
			}

			for _, config := range oracleConfig.BtcKrw {
				meta := sourceManager.GetSourceMeta(types.PairStr(types.BTC, types.KRW), config.Name, config.Interval, config.Weight)
				ps.PushSourceMeta(meta)
			}

			for _, config := range oracleConfig.LunaBtc {
				meta := sourceManager.GetSourceMeta(types.PairStr(types.LUNA, types.BTC), config.Name, config.Interval, config.Weight)
				ps.PushSourceMeta(meta)
			}

			for _, config := range oracleConfig.LunaKrw {
				meta := sourceManager.GetSourceMeta(types.PairStr(types.LUNA, types.KRW), config.Name, config.Interval, config.Weight)
				ps.PushSourceMeta(meta)
			}

			for _, config := range oracleConfig.SdrKrw {
				meta := sourceManager.GetSourceMeta(types.PairStr(types.SDR, types.KRW), config.Name, config.Interval, config.Weight)
				ps.PushSourceMeta(meta)
			}

			for _, config := range oracleConfig.UsdKrw {
				meta := sourceManager.GetSourceMeta(types.PairStr(types.USD, types.KRW), config.Name, config.Interval, config.Weight)
				ps.PushSourceMeta(meta)
			}

			oracleService := oracle.NewOracleService(*ps, cdc)
			oracleService.SetLogger(logger.With("module", "oracle"))

			// Stop upon receiving SIGTERM or CTRL-C.
			cmn.TrapSignal(logger, func() {
				if ps.IsRunning() {
					oracleService.Stop()
				}
			})

			if err := oracleService.Start(); err != nil {
				return fmt.Errorf("failed to start node: %v", err)
			}

			// Run forever.
			select {}
		},
	}

	svcCmd.Flags().String(oracle.FlagValidator, "", "")
	svcCmd.Flags().String(oracle.FlagPassword, "", "")

	svcCmd = client.PostCommands(svcCmd)[0]
	svcCmd.MarkFlagRequired(client.FlagFrom)
	svcCmd.MarkFlagRequired(oracle.FlagValidator)

	return svcCmd
}
