package keeper_test

import (
	"github.com/AstraProtocol/astra/v3/x/mint/keeper"
	"github.com/AstraProtocol/astra/v3/x/mint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"math/rand"
)

func (suite *KeeperTestSuite) TestNewQuerier() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.app.LegacyAmino())
	querier := keeper.NewQuerier(suite.app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(suite.ctx, []string{types.QueryParameters}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{types.QueryInflation}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{types.QueryAnnualProvisions}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{types.QueryTotalMintedProvision}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{types.QueryBlockProvision}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{"foo"}, query)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestQueryParams() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.app.LegacyAmino())
	querier := keeper.NewQuerier(suite.app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	var params types.Params
	for i := 0; i <= numTests; i++ {
		newParams := randomizedValidParams()
		suite.app.MintKeeper.SetParams(suite.ctx, newParams)

		res, err := querier(suite.ctx, []string{types.QueryParameters}, query)
		suite.Require().NoError(err)

		err = suite.app.LegacyAmino().UnmarshalJSON(res, &params)
		suite.Require().NoError(err)

		suite.Require().Equal(newParams, params)
	}
}

func (suite *KeeperTestSuite) TestQueryInflation() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.app.LegacyAmino())
	querier := keeper.NewQuerier(suite.app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	var inflation sdk.Dec
	for i := 0; i <= numTests; i++ {
		newMinter := randomMinter()
		suite.app.MintKeeper.SetMinter(suite.ctx, newMinter)

		res, err := querier(suite.ctx, []string{types.QueryInflation}, query)
		suite.Require().NoError(err)

		err = suite.app.LegacyAmino().UnmarshalJSON(res, &inflation)
		suite.Require().NoError(err)

		suite.Require().Equal(newMinter.Inflation, inflation)
	}
}

func (suite *KeeperTestSuite) TestQueryAnnualProvisions() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.app.LegacyAmino())
	querier := keeper.NewQuerier(suite.app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	var annualProvisions sdk.Dec
	for i := 0; i <= numTests; i++ {
		newMinter := randomMinter()
		suite.app.MintKeeper.SetMinter(suite.ctx, newMinter)

		res, err := querier(suite.ctx, []string{types.QueryAnnualProvisions}, query)
		suite.Require().NoError(err)

		err = suite.app.LegacyAmino().UnmarshalJSON(res, &annualProvisions)
		suite.Require().NoError(err)

		suite.Require().Equal(newMinter.AnnualProvisions, annualProvisions)
	}
}

func (suite *KeeperTestSuite) TestQueryTotalMintedProvision() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.app.LegacyAmino())
	querier := keeper.NewQuerier(suite.app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	var totalMinted sdk.Dec
	for i := 0; i <= numTests; i++ {
		newTotalMinted := sdk.NewDec(1 + rand.Int63())
		suite.app.MintKeeper.SetTotalMintProvision(suite.ctx, newTotalMinted)

		res, err := querier(suite.ctx, []string{types.QueryTotalMintedProvision}, query)
		suite.Require().NoError(err)

		err = suite.app.LegacyAmino().UnmarshalJSON(res, &totalMinted)
		suite.Require().NoError(err)

		suite.Require().Equal(newTotalMinted, totalMinted)
	}
}

func (suite *KeeperTestSuite) TestQueryBlockProvision() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.app.LegacyAmino())
	querier := keeper.NewQuerier(suite.app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	var blockProvision sdk.Coin
	for i := 0; i <= numTests; i++ {
		newMinter := randomMinter()
		newParams := randomizedValidParams()
		suite.app.MintKeeper.SetMinter(suite.ctx, newMinter)
		suite.app.MintKeeper.SetParams(suite.ctx, newParams)

		res, err := querier(suite.ctx, []string{types.QueryBlockProvision}, query)
		suite.Require().NoError(err)

		err = suite.app.LegacyAmino().UnmarshalJSON(res, &blockProvision)
		suite.Require().NoError(err)

		suite.Require().Equal(newMinter.BlockProvision(newParams), blockProvision)
	}
}
