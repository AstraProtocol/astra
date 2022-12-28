package tests

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/app"
	"github.com/AstraProtocol/astra/v2/app/ante"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/encoding"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
	"time"

	"github.com/AstraProtocol/astra/v2/testutil"
	"github.com/evmos/ethermint/tests"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v6/x/vesting/types"
)

func TestVestingTestingSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Example:
// 21/10 Employee joins Astra and vesting starts
// 22/03 Mainnet launch
// 22/09 Cliff ends
// 23/02 Lock ends

func (suite *KeeperTestSuite) TestCrawbackVestingAccountsTokens() {
	// Monthly vesting period
	stakeDenom := config.BaseDenom
	amt := sdk.NewInt(1)
	vestingLength := int64(60 * 60 * 24 * 30) // in seconds
	vestingAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
	vestingPeriod := sdkvesting.Period{Length: vestingLength, Amount: vestingAmt}

	// 4 years vesting total
	periodsTotal := int64(48)
	vestingTotal := amt.Mul(sdk.NewInt(periodsTotal))
	vestingAmtTotal := sdk.NewCoins(sdk.NewCoin(stakeDenom, vestingTotal))

	// 6 month cliff
	cliff := int64(6)
	cliffLength := vestingLength * cliff
	cliffAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(cliff))))
	cliffPeriod := sdkvesting.Period{Length: cliffLength, Amount: cliffAmt}

	// 12 month lockup
	lockup := int64(12) // 12 year
	lockupLength := vestingLength * lockup
	lockupPeriod := sdkvesting.Period{Length: lockupLength, Amount: vestingAmtTotal}
	lockupPeriods := sdkvesting.Periods{lockupPeriod}

	// Create vesting periods with initial cliff
	vestingPeriods := sdkvesting.Periods{cliffPeriod}
	for p := int64(1); p <= periodsTotal-cliff; p++ {
		vestingPeriods = append(vestingPeriods, vestingPeriod)
	}

	var (
		clawbackAccount *types.ClawbackVestingAccount
		vesting         sdk.Coins
		vested          sdk.Coins
		unlocked        sdk.Coins
	)
	grantee := sdk.AccAddress(tests.GenerateAddress().Bytes())
	funder := sdk.AccAddress(tests.GenerateAddress().Bytes())
	dest := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"should claw back unvested amount before cliff",
			func() {
				ctx := sdk.WrapSDKContext(suite.ctx)

				balanceFunder := suite.app.BankKeeper.GetBalance(suite.ctx, funder, stakeDenom)
				balanceGrantee := suite.app.BankKeeper.GetBalance(suite.ctx, grantee, stakeDenom)
				balanceDest := suite.app.BankKeeper.GetBalance(suite.ctx, dest, stakeDenom)

				// Perform clawback before cliff
				msg := types.NewMsgClawback(funder, grantee, dest)
				_, err := suite.app.VestingKeeper.Clawback(ctx, msg)

				suite.Require().NoError(err)

				// All initial vesting amount goes to dest
				bF := suite.app.BankKeeper.GetBalance(suite.ctx, funder, stakeDenom)
				bG := suite.app.BankKeeper.GetBalance(suite.ctx, grantee, stakeDenom)
				bD := suite.app.BankKeeper.GetBalance(suite.ctx, dest, stakeDenom)

				suite.Require().Equal(bF, balanceFunder)
				suite.Require().Equal(balanceGrantee.Sub(vestingAmtTotal[0]).Amount.Uint64(), bG.Amount.Uint64())
				suite.Require().Equal(balanceDest.Add(vestingAmtTotal[0]).Amount.Uint64(), bD.Amount.Uint64())
			},
		},
		{
			"should claw back any unvested amount after cliff before unlocking",
			func() {
				// Surpass cliff but not lockup duration
				cliffDuration := time.Duration(cliffLength)
				suite.CommitAfter(cliffDuration * time.Second)

				// Check that all tokens are locked and some, but not all tokens are vested
				vested = clawbackAccount.GetVestedOnly(suite.ctx.BlockTime())
				unlocked = clawbackAccount.GetUnlockedOnly(suite.ctx.BlockTime())
				free := clawbackAccount.GetVestedCoins(suite.ctx.BlockTime())
				vesting = clawbackAccount.GetVestingCoins(suite.ctx.BlockTime())
				expVestedAmount := amt.Mul(sdk.NewInt(cliff))
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))

				suite.Require().Equal(expVested, vested)
				suite.Require().True(expVestedAmount.GT(sdk.NewInt(0)))
				suite.Require().True(free.IsZero())
				suite.Require().Equal(vesting, vestingAmtTotal)

				balanceFunder := suite.app.BankKeeper.GetBalance(suite.ctx, funder, stakeDenom)
				balanceGrantee := suite.app.BankKeeper.GetBalance(suite.ctx, grantee, stakeDenom)
				balanceDest := suite.app.BankKeeper.GetBalance(suite.ctx, dest, stakeDenom)

				// Perform clawback
				msg := types.NewMsgClawback(funder, grantee, dest)
				ctx := sdk.WrapSDKContext(suite.ctx)
				_, err := suite.app.VestingKeeper.Clawback(ctx, msg)
				suite.Require().NoError(err)

				bF := suite.app.BankKeeper.GetBalance(suite.ctx, funder, stakeDenom)
				bG := suite.app.BankKeeper.GetBalance(suite.ctx, grantee, stakeDenom)
				bD := suite.app.BankKeeper.GetBalance(suite.ctx, dest, stakeDenom)

				expClawback := clawbackAccount.GetUnvestedOnly(suite.ctx.BlockTime())

				// Any unvested amount is clawed back
				suite.Require().Equal(balanceFunder, bF)
				suite.Require().Equal(balanceGrantee.Sub(expClawback[0]).Amount.Uint64(), bG.Amount.Uint64())
				suite.Require().Equal(balanceDest.Add(expClawback[0]).Amount.Uint64(), bD.Amount.Uint64())
			},
		},
		{
			"should claw back any unvested amount after cliff and unlocking",
			func() {
				// Surpass lockup duration
				// A strict `if t < clawbackTime` comparison is used in ComputeClawback
				// so, we increment the duration with 1 for the free token calculation to match
				lockupDuration := time.Duration(lockupLength + 1)
				suite.CommitAfter(lockupDuration * time.Second)

				// Check if some, but not all tokens are vested and unlocked
				vested = clawbackAccount.GetVestedOnly(suite.ctx.BlockTime())
				unlocked = clawbackAccount.GetUnlockedOnly(suite.ctx.BlockTime())
				free := clawbackAccount.GetVestedCoins(suite.ctx.BlockTime())
				vesting = clawbackAccount.GetVestingCoins(suite.ctx.BlockTime())
				expVestedAmount := amt.Mul(sdk.NewInt(lockup))
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
				unvested := vestingAmtTotal.Sub(vested)
				suite.Require().Equal(free, vested)
				suite.Require().Equal(expVested, vested)
				suite.Require().True(expVestedAmount.GT(sdk.NewInt(0)))
				suite.Require().Equal(vesting, unvested)

				balanceFunder := suite.app.BankKeeper.GetBalance(suite.ctx, funder, stakeDenom)
				balanceGrantee := suite.app.BankKeeper.GetBalance(suite.ctx, grantee, stakeDenom)
				balanceDest := suite.app.BankKeeper.GetBalance(suite.ctx, dest, stakeDenom)

				// Perform clawback
				msg := types.NewMsgClawback(funder, grantee, dest)
				ctx := sdk.WrapSDKContext(suite.ctx)
				_, err := suite.app.VestingKeeper.Clawback(ctx, msg)
				suite.Require().NoError(err)

				bF := suite.app.BankKeeper.GetBalance(suite.ctx, funder, stakeDenom)
				bG := suite.app.BankKeeper.GetBalance(suite.ctx, grantee, stakeDenom)
				bD := suite.app.BankKeeper.GetBalance(suite.ctx, dest, stakeDenom)

				// Any unvested amount is clawed back
				suite.Require().Equal(balanceFunder, bF)
				suite.Require().Equal(balanceGrantee.Sub(vesting[0]).Amount.Uint64(), bG.Amount.Uint64())
				suite.Require().Equal(balanceDest.Add(vesting[0]).Amount.Uint64(), bD.Amount.Uint64())
			},
		},
		{
			"should not claw back any amount after vesting periods end",
			func() {
				// Surpass vesting periods
				vestingDuration := time.Duration(periodsTotal*vestingLength + 1)
				suite.CommitAfter(vestingDuration * time.Second)

				// Check if some, but not all tokens are vested and unlocked
				vested = clawbackAccount.GetVestedOnly(suite.ctx.BlockTime())
				unlocked = clawbackAccount.GetUnlockedOnly(suite.ctx.BlockTime())
				free := clawbackAccount.GetVestedCoins(suite.ctx.BlockTime())
				vesting = clawbackAccount.GetVestingCoins(suite.ctx.BlockTime())
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(periodsTotal))))
				unvested := vestingAmtTotal.Sub(vested)
				suite.Require().Equal(free, vested)
				suite.Require().Equal(expVested, vested)
				suite.Require().Equal(expVested, vestingAmtTotal)
				suite.Require().Equal(unlocked, vestingAmtTotal)
				suite.Require().Equal(vesting, unvested)
				suite.Require().True(vesting.IsZero())

				balanceFunder := suite.app.BankKeeper.GetBalance(suite.ctx, funder, stakeDenom)
				balanceGrantee := suite.app.BankKeeper.GetBalance(suite.ctx, grantee, stakeDenom)
				balanceDest := suite.app.BankKeeper.GetBalance(suite.ctx, dest, stakeDenom)

				// Perform clawback
				msg := types.NewMsgClawback(funder, grantee, dest)
				ctx := sdk.WrapSDKContext(suite.ctx)
				_, err := suite.app.VestingKeeper.Clawback(ctx, msg)
				suite.Require().NoError(err)

				bF := suite.app.BankKeeper.GetBalance(suite.ctx, funder, stakeDenom)
				bG := suite.app.BankKeeper.GetBalance(suite.ctx, grantee, stakeDenom)
				bD := suite.app.BankKeeper.GetBalance(suite.ctx, dest, stakeDenom)

				// No amount is clawed back
				suite.Require().Equal(balanceFunder, bF)
				suite.Require().Equal(balanceGrantee, bG)
				suite.Require().Equal(balanceDest, bD)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			ctx := sdk.WrapSDKContext(suite.ctx)

			// Create and fund periodic vesting account
			vestingStart := suite.ctx.BlockTime()
			testutil.FundAccount(suite.app.BankKeeper, suite.ctx, funder, vestingAmtTotal)

			balanceFunder := suite.app.BankKeeper.GetBalance(suite.ctx, funder, stakeDenom)
			balanceGrantee := suite.app.BankKeeper.GetBalance(suite.ctx, grantee, stakeDenom)
			balanceDest := suite.app.BankKeeper.GetBalance(suite.ctx, dest, stakeDenom)
			suite.Require().True(balanceFunder.IsGTE(vestingAmtTotal[0]))
			suite.Require().Equal(balanceGrantee, sdk.NewInt64Coin(stakeDenom, 0))
			suite.Require().Equal(balanceDest, sdk.NewInt64Coin(stakeDenom, 0))

			msg := types.NewMsgCreateClawbackVestingAccount(funder, grantee, vestingStart, lockupPeriods, vestingPeriods)

			_, err := suite.app.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)
			suite.Require().NoError(err)

			acc := suite.app.AccountKeeper.GetAccount(suite.ctx, grantee)
			clawbackAccount, _ = acc.(*types.ClawbackVestingAccount)

			// Check if all tokens are unvested and locked at vestingStart
			vesting = clawbackAccount.GetVestingCoins(suite.ctx.BlockTime())
			vested = clawbackAccount.GetVestedOnly(suite.ctx.BlockTime())
			unlocked = clawbackAccount.GetUnlockedOnly(suite.ctx.BlockTime())
			suite.Require().Equal(vestingAmtTotal, vesting)
			suite.Require().True(vested.IsZero())
			suite.Require().True(unlocked.IsZero())

			bF := suite.app.BankKeeper.GetBalance(suite.ctx, funder, stakeDenom)
			balanceGrantee = suite.app.BankKeeper.GetBalance(suite.ctx, grantee, stakeDenom)
			balanceDest = suite.app.BankKeeper.GetBalance(suite.ctx, dest, stakeDenom)

			suite.Require().True(bF.IsGTE(balanceFunder.Sub(vestingAmtTotal[0])))
			suite.Require().True(balanceGrantee.IsGTE(vestingAmtTotal[0]))
			suite.Require().Equal(balanceDest, sdk.NewInt64Coin(stakeDenom, 0))
			tc.malleate()
		})
	}
}

// Example:
// 21/10 Employee joins Astra and vesting starts
// 22/03 Mainnet launch
// 22/09 Cliff ends
// 23/02 Lock ends
func (suite *KeeperTestSuite) TestCrawbackVestingAccounts() {
	// Monthly vesting period
	stakeDenom := config.BaseDenom
	amt := sdk.NewInt(1)
	vestingLength := int64(60 * 60 * 24 * 30) // in seconds
	vestingAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
	vestingPeriod := sdkvesting.Period{Length: vestingLength, Amount: vestingAmt}

	// 4 years vesting total
	periodsTotal := int64(48)
	vestingAmtTotal := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(periodsTotal))))

	// 6 month cliff
	cliff := int64(6)
	cliffLength := vestingLength * cliff
	cliffAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(cliff))))
	cliffPeriod := sdkvesting.Period{Length: cliffLength, Amount: cliffAmt}

	// 12 month lockup
	lockup := int64(12) // 12 year
	lockupLength := vestingLength * lockup
	lockupPeriod := sdkvesting.Period{Length: lockupLength, Amount: vestingAmtTotal}
	lockupPeriods := sdkvesting.Periods{lockupPeriod}

	// Create vesting periods with initial cliff
	vestingPeriods := sdkvesting.Periods{cliffPeriod}
	for p := int64(1); p <= periodsTotal-cliff; p++ {
		vestingPeriods = append(vestingPeriods, vestingPeriod)
	}

	var (
		clawbackAccount *types.ClawbackVestingAccount
		unvested        sdk.Coins
		vested          sdk.Coins
	)

	addr, _ := sdk.AccAddressFromBech32("astra1gcgds44pfkkc46tecucgtlrwjmrawy44wlpk3g")

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"before first vesting period",
			func() {
				err := suite.delegate(clawbackAccount, 100)
				suite.Require().Error(err, "cannot delegate tokens")

				err = suite.app.BankKeeper.SendCoins(
					suite.ctx,
					addr,
					tests.GenerateAddress().Bytes(),
					unvested,
				)
				suite.Require().Error(err, "cannot transfer tokens")

				err = suite.performEthTx(clawbackAccount)
				suite.Require().Error(err, "cannot perform Ethereum tx")
			},
		},
		{
			"after first vesting period and before lockup",
			func() {
				// Surpass cliff but not lockup duration
				cliffDuration := time.Duration(cliffLength)
				suite.CommitAfter(cliffDuration * time.Second)

				// Check if some, but not all tokens are vested
				vested = clawbackAccount.GetVestedOnly(suite.ctx.BlockTime())
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(cliff))))
				suite.Require().NotEqual(vestingAmtTotal, vested)
				suite.Require().Equal(expVested, vested)

				err := suite.delegate(clawbackAccount, 1)
				suite.Require().NoError(err, "can delegate vested tokens")

				err = suite.app.BankKeeper.SendCoins(
					suite.ctx,
					addr,
					tests.GenerateAddress().Bytes(),
					vested,
				)

				suite.Require().Error(err, "cannot transfer vested tokens")

				err = suite.performEthTx(clawbackAccount)

				suite.Require().Error(err, "cannot perform Ethereum tx")
			},
		},
		{
			"after first vesting period and lockup",
			func() {
				// Surpass lockup duration
				lockupDuration := time.Duration(lockupLength)
				suite.CommitAfter(lockupDuration * time.Second)

				// Check if some, but not all tokens are vested
				unvested = clawbackAccount.GetUnvestedOnly(suite.ctx.BlockTime())
				vested = clawbackAccount.GetVestedOnly(suite.ctx.BlockTime())
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(lockup))))
				suite.Require().NotEqual(vestingAmtTotal, vested)
				suite.Require().Equal(expVested, vested)

				err := suite.delegate(clawbackAccount, 1)
				suite.Require().NoError(err, "can delegate vested tokens")

				err = suite.delegate(clawbackAccount, 30)
				suite.Require().Error(err, "cannot delegate unvested tokens")

				err = suite.app.BankKeeper.SendCoins(
					suite.ctx,
					addr,
					tests.GenerateAddress().Bytes(),
					vested,
				)
				suite.Require().NoError(err, "can transfer vested tokens")

				err = suite.app.BankKeeper.SendCoins(
					suite.ctx,
					addr,
					tests.GenerateAddress().Bytes(),
					unvested,
				)
				suite.Require().Error(err, "cannot transfer unvested tokens")

				err = suite.performEthTx(clawbackAccount)
				suite.Require().NoError(err, "can perform ethereum tx")
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			// Create and fund periodic vesting account
			vestingStart := suite.ctx.BlockTime()
			baseAccount := authtypes.NewBaseAccountWithAddress(addr)
			funder := sdk.AccAddress(types.ModuleName)
			clawbackAccount = types.NewClawbackVestingAccount(
				baseAccount,
				funder,
				vestingAmtTotal,
				vestingStart,
				lockupPeriods,
				vestingPeriods,
			)
			err := testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, vestingAmtTotal)
			suite.Require().NoError(err)
			acc := suite.app.AccountKeeper.NewAccount(suite.ctx, clawbackAccount)
			suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

			// Check if all tokens are unvested at vestingStart
			unvested = clawbackAccount.GetUnvestedOnly(suite.ctx.BlockTime())
			vested = clawbackAccount.GetVestedOnly(suite.ctx.BlockTime())
			suite.Require().Equal(vestingAmtTotal, unvested)
			suite.Require().True(vested.IsZero())

			tc.malleate()
		})
	}
}

func nextFn(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}

func (suite *KeeperTestSuite) delegate(clawbackAccount *types.ClawbackVestingAccount, amount int64) error {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	addr, err := sdk.AccAddressFromBech32(clawbackAccount.Address)
	suite.Require().NoError(err)
	//
	val, err := sdk.ValAddressFromBech32("astravaloper1z3t55m0l9h0eupuz3dp5t5cypyv674jj6flkt5")
	suite.Require().NoError(err)
	delegateMsg := stakingtypes.NewMsgDelegate(addr, val, sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, sdk.NewInt(amount)))
	txBuilder.SetMsgs(delegateMsg)
	tx := txBuilder.GetTx()

	dec := ante.NewVestingDelegationDecorator(suite.app.AccountKeeper, suite.app.StakingKeeper, types.ModuleCdc)
	_, err = dec.AnteHandle(suite.ctx, tx, false, nextFn)
	return err
}

func (suite *KeeperTestSuite) performEthTx(clawbackAccount *types.ClawbackVestingAccount) error {
	addr, err := sdk.AccAddressFromBech32(clawbackAccount.Address)
	suite.Require().NoError(err)
	chainID := suite.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(addr.Bytes())
	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, from)

	msgEthereumTx := evmtypes.NewTx(chainID, nonce, &from, nil, 100000, nil,
		suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx), big.NewInt(1), nil, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msgEthereumTx)
	tx := txBuilder.GetTx()

	// Call Ante decorator
	dec := ante.NewEthVestingTransactionDecorator(suite.app.AccountKeeper)
	_, err = dec.AnteHandle(suite.ctx, tx, false, nextFn)
	return err
}
