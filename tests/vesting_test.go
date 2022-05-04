package tests

import (
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/AstraProtocol/astra/v1/testutil"
	"github.com/tharsis/ethermint/tests"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tharsis/evmos/v3/x/vesting/types"
)

func TestVestingTestingSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestCrawbackVestingAccounts() {
	// Monthly vesting period
	stakeDenom := stakingtypes.DefaultParams().BondDenom
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

	s.SetupTest()
	ctx := sdk.WrapSDKContext(s.ctx)

	// Create and fund periodic vesting account
	vestingStart := s.ctx.BlockTime()
	testutil.FundAccount(s.app.BankKeeper, s.ctx, funder, vestingAmtTotal)

	balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
	balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
	balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
	s.Require().True(balanceFunder.IsGTE(vestingAmtTotal[0]))
	s.Require().Equal(balanceGrantee, sdk.NewInt64Coin(stakeDenom, 0))
	s.Require().Equal(balanceDest, sdk.NewInt64Coin(stakeDenom, 0))

	msg := types.NewMsgCreateClawbackVestingAccount(funder, grantee, vestingStart, lockupPeriods, vestingPeriods, true)

	_, err := s.app.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)
	s.Require().NoError(err)

	acc := s.app.AccountKeeper.GetAccount(s.ctx, grantee)
	clawbackAccount, _ = acc.(*types.ClawbackVestingAccount)

	// Check if all tokens are unvested and locked at vestingStart
	vesting = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
	vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
	unlocked = clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
	s.Require().Equal(vestingAmtTotal, vesting)
	s.Require().True(vested.IsZero())
	s.Require().True(unlocked.IsZero())

	bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
	balanceGrantee = s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
	balanceDest = s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

	s.Require().True(bF.IsGTE(balanceFunder.Sub(vestingAmtTotal[0])))
	s.Require().True(balanceGrantee.IsGTE(vestingAmtTotal[0]))
	s.Require().Equal(balanceDest, sdk.NewInt64Coin(stakeDenom, 0))

	ctx = sdk.WrapSDKContext(s.ctx)

	balanceFunder = s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
	balanceGrantee = s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
	balanceDest = s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

	// Perform clawback before cliff
	msg1 := types.NewMsgClawback(funder, grantee, dest)
	_, err = s.app.VestingKeeper.Clawback(ctx, msg1)
	Expect(err).To(BeNil())

	// All initial vesting amount goes to dest
	bF = s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
	bG := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
	bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

	s.Require().Equal(bF, balanceFunder)
	s.Require().Equal(balanceGrantee.Sub(vestingAmtTotal[0]).Amount.Uint64(), bG.Amount.Uint64())
	s.Require().Equal(balanceDest.Add(vestingAmtTotal[0]).Amount.Uint64(), bD.Amount.Uint64())
}

//
//// Example:
//// 21/10 Employee joins Evmos and vesting starts
//// 22/03 Mainnet launch
//// 22/09 Cliff ends
//// 23/02 Lock ends
//var _ = Describe("Clawback Vesting Accounts - claw back tokens", Ordered, func() {
//
//	BeforeEach(func() {
//
//	})
//
//	It("should claw back unvested amount before cliff", func() {
//
//	})
//
//	It("should claw back any unvested amount after cliff before unlocking", func() {
//		// Surpass cliff but not lockup duration
//		cliffDuration := time.Duration(cliffLength)
//		s.CommitAfter(cliffDuration * time.Second)
//
//		// Check that all tokens are locked and some, but not all tokens are vested
//		vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
//		unlocked = clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
//		free = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
//		vesting = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
//		expVestedAmount := amt.Mul(sdk.NewInt(cliff))
//		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
//
//		s.Require().Equal(expVested, vested)
//		s.Require().True(expVestedAmount.GT(sdk.NewInt(0)))
//		s.Require().True(free.IsZero())
//		s.Require().Equal(vesting, vestingAmtTotal)
//
//		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
//		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
//		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
//
//		// Perform clawback
//		msg := types.NewMsgClawback(funder, grantee, dest)
//		ctx := sdk.WrapSDKContext(s.ctx)
//		_, err := s.app.VestingKeeper.Clawback(ctx, msg)
//		Expect(err).To(BeNil())
//
//		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
//		bG := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
//		bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
//
//		expClawback := clawbackAccount.GetUnvestedOnly(s.ctx.BlockTime())
//
//		// Any unvested amount is clawed back
//		s.Require().Equal(balanceFunder, bF)
//		s.Require().Equal(balanceGrantee.Sub(expClawback[0]).Amount.Uint64(), bG.Amount.Uint64())
//		s.Require().Equal(balanceDest.Add(expClawback[0]).Amount.Uint64(), bD.Amount.Uint64())
//	})
//
//	It("should claw back any unvested amount after cliff and unlocking", func() {
//		// Surpass lockup duration
//		// A strict `if t < clawbackTime` comparison is used in ComputeClawback
//		// so, we increment the duration with 1 for the free token calculation to match
//		lockupDuration := time.Duration(lockupLength + 1)
//		s.CommitAfter(lockupDuration * time.Second)
//
//		// Check if some, but not all tokens are vested and unlocked
//		vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
//		unlocked = clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
//		free = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
//		vesting = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
//		expVestedAmount := amt.Mul(sdk.NewInt(lockup))
//		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
//		unvested := vestingAmtTotal.Sub(vested)
//		s.Require().Equal(free, vested)
//		s.Require().Equal(expVested, vested)
//		s.Require().True(expVestedAmount.GT(sdk.NewInt(0)))
//		s.Require().Equal(vesting, unvested)
//
//		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
//		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
//		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
//
//		// Perform clawback
//		msg := types.NewMsgClawback(funder, grantee, dest)
//		ctx := sdk.WrapSDKContext(s.ctx)
//		_, err := s.app.VestingKeeper.Clawback(ctx, msg)
//		Expect(err).To(BeNil())
//
//		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
//		bG := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
//		bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
//
//		// Any unvested amount is clawed back
//		s.Require().Equal(balanceFunder, bF)
//		s.Require().Equal(balanceGrantee.Sub(vesting[0]).Amount.Uint64(), bG.Amount.Uint64())
//		s.Require().Equal(balanceDest.Add(vesting[0]).Amount.Uint64(), bD.Amount.Uint64())
//	})
//
//	It("should not claw back any amount after vesting periods end", func() {
//		// Surpass vesting periods
//		vestingDuration := time.Duration(periodsTotal*vestingLength + 1)
//		s.CommitAfter(vestingDuration * time.Second)
//
//		// Check if some, but not all tokens are vested and unlocked
//		vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
//		unlocked = clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
//		free = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
//		vesting = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
//		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(periodsTotal))))
//		unvested := vestingAmtTotal.Sub(vested)
//		s.Require().Equal(free, vested)
//		s.Require().Equal(expVested, vested)
//		s.Require().Equal(expVested, vestingAmtTotal)
//		s.Require().Equal(unlocked, vestingAmtTotal)
//		s.Require().Equal(vesting, unvested)
//		s.Require().True(vesting.IsZero())
//
//		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
//		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
//		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
//
//		// Perform clawback
//		msg := types.NewMsgClawback(funder, grantee, dest)
//		ctx := sdk.WrapSDKContext(s.ctx)
//		_, err := s.app.VestingKeeper.Clawback(ctx, msg)
//		Expect(err).To(BeNil())
//
//		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
//		bG := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
//		bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
//
//		// No amount is clawed back
//		s.Require().Equal(balanceFunder, bF)
//		s.Require().Equal(balanceGrantee, bG)
//		s.Require().Equal(balanceDest, bD)
//	})
//})
