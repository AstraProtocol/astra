package v1_1_test

import (
	"fmt"
	"github.com/AstraProtocol/astra/v1/app/upgrades/v1_1"
	astratypes "github.com/AstraProtocol/astra/v1/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/AstraProtocol/astra/v1/app"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.Astra
	consAddress sdk.ConsAddress
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest(chainID string) {
	checkTx := false

	// consensus key
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// NOTE: this is the new binary, not the old one.
	suite.app = app.Setup(checkTx, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         chainID,
		Time:            time.Date(2022, 5, 9, 8, 0, 0, 0, time.UTC),
		ProposerAddress: suite.consAddress.Bytes(),

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

	cp := suite.app.BaseApp.GetConsensusParams(suite.ctx)
	suite.ctx = suite.ctx.WithConsensusParams(cp)
}

func (suite *UpgradeTestSuite) TestUpdateConsensusParams() {
	suite.SetupTest(astratypes.TestnetChainID + "-2") // reset
	unbondingDuration := suite.app.GetStakingKeeper().UnbondingTime(suite.ctx)

	testCases := []struct {
		name              string
		malleate          func()
		expEvidenceParams *tmproto.EvidenceParams
		expBlockParams    *abci.BlockParams
		pass              bool
	}{
		{
			"empty evidence params",
			func() {
				subspace, found := suite.app.ParamsKeeper.GetSubspace(baseapp.Paramspace)
				suite.Require().True(found)

				ep := &tmproto.EvidenceParams{}
				subspace.Set(suite.ctx, baseapp.ParamStoreKeyEvidenceParams, ep)
			},
			&tmproto.EvidenceParams{},
			&abci.BlockParams{},
			false,
		},
		{
			"success",
			func() {
				subspace, found := suite.app.ParamsKeeper.GetSubspace(baseapp.Paramspace)
				suite.Require().True(found)

				ep := &tmproto.EvidenceParams{
					MaxAgeDuration:  2 * 24 * time.Hour,
					MaxAgeNumBlocks: 100000,
					MaxBytes:        suite.ctx.ConsensusParams().Evidence.MaxBytes,
				}
				subspace.Set(suite.ctx, baseapp.ParamStoreKeyEvidenceParams, ep)

				bp := &abci.BlockParams{
					MaxBytes: 10000,
					MaxGas:   10000000,
				}
				subspace.Set(suite.ctx, baseapp.ParamStoreKeyBlockParams, bp)
			},
			&tmproto.EvidenceParams{
				MaxAgeDuration:  unbondingDuration,
				MaxAgeNumBlocks: int64(unbondingDuration / (3 * time.Second)),
				MaxBytes:        suite.ctx.ConsensusParams().Evidence.MaxBytes,
			},
			&abci.BlockParams{
				MaxBytes: suite.ctx.ConsensusParams().Block.GetMaxBytes(),
				MaxGas:   40000000,
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest(astratypes.TestnetChainID + "-1") // reset

			tc.malleate()

			suite.Require().NotPanics(func() {
				v1_1.UpdateConsensusParams(suite.ctx, suite.app.StakingKeeper, suite.app.ParamsKeeper)
				suite.app.Commit()
			})

			cp := suite.app.BaseApp.GetConsensusParams(suite.ctx)
			suite.Require().NotNil(cp)
			suite.Require().NotNil(cp.Evidence)
			suite.Require().Equal(tc.expEvidenceParams.String(), cp.Evidence.String())
			if tc.pass {
				suite.Require().Equal(tc.expBlockParams.MaxGas, cp.Block.MaxGas)
			}
		})
	}
}
