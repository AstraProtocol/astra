package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	// check calculated epochMintProvision at genesis
	epochMintProvision, _ := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
	expMintProvision := sdk.MustNewDecFromStr("304410958904109589041095.000000000000000000")
	suite.Require().Equal(expMintProvision, epochMintProvision)
}
