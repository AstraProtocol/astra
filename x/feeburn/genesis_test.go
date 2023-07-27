package feeburn_test

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/app"
	"github.com/AstraProtocol/astra/v2/x/feeburn"
	"github.com/AstraProtocol/astra/v2/x/feeburn/types"
	utiltx "github.com/evmos/evmos/v12/testutil/tx"

	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	feemarkettypes "github.com/evmos/evmos/v12/x/feemarket/types"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app     *app.Astra
	genesis types.GenesisState
}

func (suite *GenesisTestSuite) SetupTest() {
	// consensus key
	consAddress := sdk.ConsAddress(utiltx.GenerateAddress().Bytes())

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
		Height:          1,
		ChainID:         "astra_11111-1",
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	suite.genesis = *types.DefaultGenesis()
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestFeeBurnInitGenesis() {
	testCases := []struct {
		name     string
		genesis  types.GenesisState
		expPanic bool
	}{
		{
			"default genesis",
			suite.genesis,
			false,
		},
		{
			"custom genesis - feeburn disabled",
			types.GenesisState{
				Params: types.Params{
					EnableFeeBurn: false,
					FeeBurn:       types.DefaultFeeBurn,
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			if tc.expPanic {
				suite.Require().Panics(func() {
					feeburn.InitGenesis(suite.ctx, suite.app.FeeBurnKeeper, tc.genesis)
				})
			} else {
				suite.Require().NotPanics(func() {
					feeburn.InitGenesis(suite.ctx, suite.app.FeeBurnKeeper, tc.genesis)
				})

				params := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
				suite.Require().Equal(tc.genesis.Params, params)
			}
		})
	}
}

func (suite *GenesisTestSuite) TestFeeBurnExportGenesis() {
	feeburn.InitGenesis(suite.ctx, suite.app.FeeBurnKeeper, suite.genesis)

	genesisExported := feeburn.ExportGenesis(suite.ctx, suite.app.FeeBurnKeeper)
	suite.Require().Equal(genesisExported.Params, suite.genesis.Params)
}
