package main

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/libs/cli"
	cmn "github.com/tendermint/tendermint/libs/common"
)

func initCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := viper.GetString(cli.HomeFlag)

			cfgPath := path.Join(home, "config")
			cfgFile := path.Join(cfgPath, "oracle.toml")
			if _, err := os.Stat(cfgFile); err == nil {
				return fmt.Errorf("oracle config file already exist")
			}

			if err := os.MkdirAll(cfgPath, os.ModePerm); err != nil {
				return err
			}

			cmn.MustWriteFile(cfgFile, []byte(defaultOracleConfigTemplate), 0644)

			return nil
		},
	}

	return initCmd
}
