package keeper_test

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/evmos/ethermint/tests"
)

func (suite *KeeperTestSuite) TestBurnFeeCosmos() {
	suite.SetupTest()
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())
	totalSupply := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	accBalance := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(1000)}}
	fmt.Println("totalSupply", totalSupply)
	err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr, accBalance)
	suite.Require().NoError(err)
}
