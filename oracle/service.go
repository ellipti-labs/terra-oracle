package oracle

import (
	"time"

	"github.com/tendermint/go-amino"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/everett-protocol/terra-oracle/price"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
)

const VotePeriod = 12

type OracleService struct {
	cmn.BaseService
	ps  price.PriceService
	cdc *amino.Codec

	passphrase    string
	txBldr        authtxb.TxBuilder
	cliCtx        context.CLIContext
	prevoteInited bool

	salts         map[string]string
	lunaPrices    map[string]sdk.Dec
	preLunaPrices map[string]sdk.Dec
}

func NewOracleService(ps price.PriceService, cdc *amino.Codec) *OracleService {
	os := &OracleService{
		ps:            ps,
		cdc:           cdc,
		salts:         make(map[string]string),
		lunaPrices:    make(map[string]sdk.Dec),
		preLunaPrices: make(map[string]sdk.Dec),
	}
	os.BaseService = *cmn.NewBaseService(nil, "OracleService", os)
	return os
}

func (os *OracleService) OnStart() error {
	err := os.init()
	if err != nil {
		return err
	}

	err = os.ps.Start()
	if err != nil {
		return err
	}

	// Wait a second until price service fetchs price initially
	time.Sleep(3 * time.Second)

	go os.startRoutine()

	return nil
}
