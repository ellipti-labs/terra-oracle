package oracle

import (
	"errors"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

	"github.com/tendermint/tendermint/rpc/client"

	"github.com/everett-protocol/terra-oracle/types"
)

const (
	FlagValidator = "validator"
)

func (os *OracleService) init() error {
	os.txBldr = authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(os.cdc))
	os.cliCtx = context.NewCLIContext().
		WithCodec(os.cdc).
		WithAccountDecoder(os.cdc)

	if os.cliCtx.BroadcastMode != "block" {
		return errors.New("I recommend to use block broadcast mode")
	}

	fromName := os.cliCtx.GetFromName()
	_passphrase, err := keys.GetPassphrase(fromName)
	if err != nil {
		return err
	}
	os.passphrase = _passphrase

	return nil
}

func (os *OracleService) startRoutine() {
	httpClient := client.NewHTTP(os.cliCtx.NodeURI, "/websocket")

	var latestVoteHeight int64 = 0

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					os.Logger.Error("Unknown error", r)
				}

				time.Sleep(5 * time.Second)
			}()

			status, err := httpClient.Status()
			if err != nil {
				os.Logger.Error("Fail to fetch status", err.Error())
				return
			}
			latestHeignt := status.SyncInfo.LatestBlockHeight

			var tick int64 = latestHeignt / VotePeriod
			if tick <= latestVoteHeight/VotePeriod {
				return
			}
			latestVoteHeight = latestHeignt
			os.Logger.Info(fmt.Sprintf("Tick: %d", tick))

			err = os.calculatePrice()
			if err != nil {
				os.Logger.Error(fmt.Sprintf("Error when calculate price: %s", err.Error()))
			}

			denoms := []string{types.KRW, types.USD, types.SDR}

			if os.prevoteInited {
				os.Logger.Info(fmt.Sprintf("Try to send vote msg (including prevote for next vote msg)"))

				voteMsgs, err := os.makeVoteMsgs(denoms)
				if err != nil {
					os.Logger.Error("Fail to make vote msgs", err.Error())
				}

				// Because vote tx includes prevote for next price,
				// use twice as much gas.
				res, err := os.broadcast(voteMsgs)
				if err != nil {
					os.Logger.Error("Fail to send vote msgs", err.Error())
					return
				}
				if tick > res.Height/VotePeriod {
					os.Logger.Error("Tx couldn't be sent within vote period")
					os.prevoteInited = false // Retry initialization (prevote msg).
				}
			} else {
				os.Logger.Info(fmt.Sprintf("Try to send prevote msg"))

				prevoteMsgs, err := os.makePrevoteMsgs(denoms)
				if err != nil {
					os.Logger.Error("Fail to make prevote msgs", err.Error())
				}

				_, err = os.broadcast(prevoteMsgs)
				if err != nil {
					os.Logger.Error("Fail to send prevote msgs", err.Error())

					os.prevoteInited = false // Retry initialization (prevote msg).
					return
				}

				os.prevoteInited = true
			}
		}()
	}
}

func (os *OracleService) calculatePriceSynthetic(base string, quote string, intermediation string) (sdk.Dec, error) {
	baseToInter, baseToInterWeight, baseToInterErr := os.ps.GetPrice(base, intermediation)
	interToQuote, _, interToQuoteErr := os.ps.GetPrice(intermediation, quote)

	sumBaseToQuote := sdk.ZeroDec()
	sumBaseToQuoteWeight := uint64(0)
	baseToQuote, baseToQuoteWeight, baseToQuoteErr := os.ps.GetPrice(base, quote)
	if baseToQuoteErr == nil {
		sumBaseToQuote = sumBaseToQuote.Add(baseToQuote.MulInt64(int64(baseToQuoteWeight)))
		sumBaseToQuoteWeight += baseToQuoteWeight
	}

	if baseToQuoteErr != nil && (baseToInterErr != nil || interToQuoteErr != nil) {
		return sdk.Dec{}, fmt.Errorf("can't calculate pair %s", types.PairStr(base, quote))
	}

	sumBaseToQuote = sumBaseToQuote.Add(baseToInter.Mul(interToQuote).MulInt64(int64(baseToInterWeight)))
	sumBaseToQuoteWeight += baseToInterWeight

	if sumBaseToQuoteWeight == 0 {
		return sdk.Dec{}, fmt.Errorf("can't calculate weighted mean pair %s", types.PairStr(base, quote))
	}

	return sumBaseToQuote.QuoInt64(int64(sumBaseToQuoteWeight)), nil
}

func (os *OracleService) calculatePrice() error {
	lunaToKrw, err := os.calculatePriceSynthetic(types.LUNA, types.KRW, types.BTC)
	if err != nil {
		return err
	}
	os.lunaPrices[types.PairStr(types.LUNA, types.KRW)] = lunaToKrw
	os.Logger.Info(fmt.Sprintf("Calculated %s: %s", types.PairStr(types.LUNA, types.KRW), lunaToKrw))

	lunaToUsd, err := os.calculatePriceSynthetic(types.LUNA, types.USD, types.BTC)
	if err != nil {
		return err
	}
	os.lunaPrices[types.PairStr(types.LUNA, types.USD)] = lunaToUsd
	os.Logger.Info(fmt.Sprintf("Calculated %s: %s", types.PairStr(types.LUNA, types.USD), lunaToUsd))

	// TODO: Use usd as intermediation instead of krw
	lunaToSdr, err := os.calculatePriceSynthetic(types.LUNA, types.SDR, types.KRW)
	if err != nil {
		return err
	}
	os.lunaPrices[types.PairStr(types.LUNA, types.SDR)] = lunaToSdr
	os.Logger.Info(fmt.Sprintf("Calculated %s: %s", types.PairStr(types.LUNA, types.SDR), lunaToSdr))

	return nil
}
