package oracle

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/terra-project/core/x/oracle"

	"github.com/everett-protocol/terra-oracle/types"
)

func (os *OracleService) makePrevoteMsgs(denoms []string) ([]sdk.Msg, error) {
	feeder := os.cliCtx.GetFromAddress()
	validator, err := sdk.ValAddressFromBech32(viper.GetString(FlagValidator))
	if err != nil {
		return nil, fmt.Errorf("invalid validator: %s", err.Error())
	}

	prevoteMsgs := make([]sdk.Msg, 0)
	for _, denom := range denoms {
		pairStr := types.PairStr(types.LUNA, denom)
		price := os.lunaPrices[pairStr]
		if price.IsZero() {
			return nil, fmt.Errorf("price %s is not set", pairStr)
		}

		salt, err := generateRandomString(4)
		if err != nil {
			return nil, fmt.Errorf("fail to generate salt: %s", err.Error())
		}
		os.salts[pairStr] = salt
		voteHash, err := oracle.VoteHash(salt, price, "u"+denom, validator)
		if err != nil {
			return nil, fmt.Errorf("fail to vote hash: %s", err.Error())
		}

		prevote := oracle.NewMsgPricePrevote(hex.EncodeToString(voteHash), "u"+denom, feeder, validator)
		prevoteMsgs = append(prevoteMsgs, prevote)

		os.preLunaPrices[pairStr] = os.lunaPrices[pairStr]
	}

	return prevoteMsgs, nil
}

func (os *OracleService) makeVoteMsgs(denoms []string) ([]sdk.Msg, error) {
	feeder := os.cliCtx.GetFromAddress()
	validator, err := sdk.ValAddressFromBech32(viper.GetString(FlagValidator))
	if err != nil {
		return nil, fmt.Errorf("invalid validator: %s", err.Error())
	}

	voteMsgs := make([]sdk.Msg, 0)
	for _, denom := range denoms {
		pairStr := types.PairStr(types.LUNA, denom)
		price := os.preLunaPrices[pairStr]
		if price.IsZero() {
			return nil, fmt.Errorf("price %s is not set", pairStr)
		}

		salt := os.salts[pairStr]
		if len(salt) == 0 {
			// It can occur before the first prevote was sent
			// So, this error may be temporary
			return nil, fmt.Errorf("fail to get salt")
		}
		vote := oracle.NewMsgPriceVote(price, salt, "u"+denom, feeder, validator)
		voteMsgs = append(voteMsgs, vote)
	}

	for _, denom := range denoms {
		pairStr := types.PairStr(types.LUNA, denom)
		price := os.lunaPrices[pairStr]
		if price.IsZero() {
			return nil, fmt.Errorf("price %s is not set", pairStr)
		}

		salt, err := generateRandomString(4)
		if err != nil {
			return nil, fmt.Errorf("fail to generate salt: %s", err.Error())
		}
		os.salts[pairStr] = salt
		voteHash, err := oracle.VoteHash(salt, price, "u"+denom, validator)
		if err != nil {
			return nil, fmt.Errorf("fail to vote hash: %s", err.Error())
		}

		prevote := oracle.NewMsgPricePrevote(hex.EncodeToString(voteHash), "u"+denom, feeder, validator)
		voteMsgs = append(voteMsgs, prevote)

		os.preLunaPrices[pairStr] = os.lunaPrices[pairStr]
	}

	return voteMsgs, nil
}

func (os *OracleService) broadcast(msgs []sdk.Msg) (*sdk.TxResponse, error) {
	txBldr, err := utils.PrepareTxBuilder(os.txBldr, os.cliCtx)
	if err != nil {
		return nil, err
	}

	fromName := os.cliCtx.GetFromName()

	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(fromName, os.passphrase, msgs)
	if err != nil {
		return nil, err
	}

	// broadcast to a Tendermint node
	res, err := os.cliCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	return &res, os.cliCtx.PrintOutput(res)
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
