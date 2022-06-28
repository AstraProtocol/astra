package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	// check calculated epochMintProvision at genesis
	epochMintProvision, _ := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
	expMintProvision := sdk.MustNewDecFromStr("608821917808219178082192.000000000000000000")
	suite.Require().Equal(expMintProvision, epochMintProvision)
}
