package keeper_test

import (
	"github.com/AstraProtocol/astra/v2/x/feeburn/keeper"
	"github.com/AstraProtocol/astra/v2/x/feeburn/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) TestNewQuerier() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.app.LegacyAmino())
	querier := keeper.NewQuerier(suite.app.FeeBurnKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(suite.ctx, []string{types.QueryParameters}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{types.QueryTotalFeeBurn}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{"foo"}, query)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestQueryParams() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.app.LegacyAmino())
	querier := keeper.NewQuerier(suite.app.FeeBurnKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	var params types.Params

	res, err := querier(suite.ctx, []string{types.QueryParameters}, query)
	suite.Require().NoError(err)

	err = suite.app.LegacyAmino().UnmarshalJSON(res, &params)
	suite.Require().NoError(err)

	suite.Require().Equal(suite.app.FeeBurnKeeper.GetParams(suite.ctx), params)
}

func (suite *KeeperTestSuite) TestQueryTotalFeeBurn() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.app.LegacyAmino())
	querier := keeper.NewQuerier(suite.app.FeeBurnKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	var totalFeeBurn sdk.Dec
	suite.app.FeeBurnKeeper.SetTotalFeeBurn(suite.ctx, sdk.NewDec(1000))

	res, err := querier(suite.ctx, []string{types.QueryTotalFeeBurn}, query)
	suite.Require().NoError(err)

	err = suite.app.LegacyAmino().UnmarshalJSON(res, &totalFeeBurn)
	suite.Require().NoError(err)

	suite.Require().Equal(sdk.NewDec(1000), totalFeeBurn)
}
