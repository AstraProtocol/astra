package keeper_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestSetGetTotalMintProvision() {
	totalMintProvision := sdk.NewDec(1_000_000)

	testCases := []struct {
		name     string
		malleate func()
		genesis  bool
	}{
		{
			"genesis EpochMintProvision",
			func() {},
			true,
		},
		{
			"newly set EpochMintProvision",
			func() {
				suite.app.InflationKeeper.SetTotalMintProvision(suite.ctx, totalMintProvision)
			},
			false,
		},
	}

	genesisProvision := sdk.MustNewDecFromStr("0.000000000000000000")

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			provision := suite.app.InflationKeeper.GetTotalMintProvision(suite.ctx)

			if tc.genesis {
				suite.Require().Equal(genesisProvision, provision, tc.name)
			} else {
				suite.Require().Equal(totalMintProvision, provision, tc.name)
			}
		})
	}
}
