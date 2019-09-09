package main

import (
	"fmt"
	btcUpbit "github.com/everett-protocol/terra-oracle/source/btckrw/upbit"
	"github.com/everett-protocol/terra-oracle/source/btcusd/binance"
	"github.com/everett-protocol/terra-oracle/source/lunabtc/upbit"
	"github.com/everett-protocol/terra-oracle/source/lunakrw/coinone"
	"github.com/everett-protocol/terra-oracle/source/sdrkrw/imf"
	"github.com/everett-protocol/terra-oracle/source/usdkrw/forex"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/terra-project/core/app"
	"github.com/terra-project/core/types/util"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"

	_ "github.com/terra-project/core/client/lcd/statik"

	"github.com/everett-protocol/terra-oracle/oracle"
	"github.com/everett-protocol/terra-oracle/price"
)

var (
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
)

func main() {
	// Configure cobra to sort commands
	cobra.EnableCommandSorting = false

	// Instantiate the codec for the command line application
	cdc := app.MakeCodec()

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(util.Bech32PrefixAccAddr, util.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(util.Bech32PrefixValAddr, util.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(util.Bech32PrefixConsAddr, util.Bech32PrefixConsPub)
	config.Seal()

	rootCmd := &cobra.Command{
		Use: "terraoracled",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		svcCmd(cdc),
		client.LineBreak,
		keys.Commands(),
	)

	// Add flags and prefix all env exposed with GA
	executor := cli.PrepareMainCmd(rootCmd, "TE", app.DefaultCLIHome)

	err := executor.Execute()
	if err != nil {
		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
		os.Exit(1)
	}
}

func svcCmd(cdc *amino.Codec) *cobra.Command {
	svcCmd := &cobra.Command{
		Use:   "service",
		Short: "Transactions subcommands",
		RunE: func(cmd *cobra.Command, args []string) error {
			ps := price.NewPriceService()
			ps.SetLogger(logger.With("module", "price"))

			ps.PushSourceMeta(price.SourceMeta{
				Source: coinone.NewCoinoneSource(),
				Weight: 10,
			})
			ps.PushSourceMeta(price.SourceMeta{
				Source: upbit.NewUpbitSource(),
				Weight: 10,
			})
			ps.PushSourceMeta(price.SourceMeta{
				Source: btcUpbit.NewUpbitSource(),
				Weight: 10,
			})
			ps.PushSourceMeta(price.SourceMeta{
				Source: forex.NewForexDonamuSource(),
				Weight: 10,
			})
			ps.PushSourceMeta(price.SourceMeta{
				Source: binance.NewBinanceSource(),
				Weight: 10,
			})
			ps.PushSourceMeta(price.SourceMeta{
				Source: imf.NewIMFSource(),
				Weight: 10,
			})

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

	svcCmd = client.PostCommands(svcCmd)[0]
	svcCmd.MarkFlagRequired(client.FlagFrom)
	svcCmd.MarkFlagRequired(oracle.FlagValidator)

	return svcCmd
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
