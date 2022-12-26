package keeper_test

import (
	"github.com/AstraProtocol/astra/v2/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
)

func randomMinter() types.Minter {
	return types.Minter{
		Inflation:        sdk.NewDecWithPrec(1+rand.Int63n(100), 2),
		AnnualProvisions: sdk.NewDec(1 + rand.Int63()),
	}
}

func (suite *KeeperTestSuite) TestMinter() {
	minter := suite.app.MintKeeper.GetMinter(suite.ctx)
	expMinter := types.DefaultInitialMinter()

	suite.Require().Equal(expMinter, minter)

	for i := 0; i < numRandTests; i++ {
		minter = randomMinter()
		suite.app.MintKeeper.SetMinter(suite.ctx, minter)
		newMinter := suite.app.MintKeeper.GetMinter(suite.ctx)
		suite.Require().Equal(newMinter, minter)
	}
}
