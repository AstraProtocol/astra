package keeper_test

import (
	"github.com/AstraProtocol/astra/v2/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tendermint/tendermint/libs/rand"
)

var _ = Describe("Mint", Ordered, func() {
	var params types.Params
	var initialSupply sdk.Int
	BeforeEach(func() {
		s.SetupTest()
		params = s.app.MintKeeper.GetParams(s.ctx)

		// we set an initial supply to equal 1200000000astra and send it to the foundation address.
		var ok bool
		initialSupply, ok = sdk.NewIntFromString("1200000000000000000000000000")
		Expect(ok).To(BeTrue())

		s.mintAndTransfer(params.MintDenom, initialSupply, sdk.AccAddress{}, initialSupply)
	})

	var foundationAddr sdk.AccAddress
	var stakingModuleAddr sdk.AccAddress
	var communityAddr sdk.AccAddress
	var mintModuleAddr sdk.AccAddress

	var oldMinter types.Minter
	var oldSupply sdk.Int
	var oldBondedRatio sdk.Dec
	var oldFoundationBalance sdk.Int
	var oldStakingModuleBalance sdk.Int
	var oldCommunityBalance sdk.Int

	Describe("Committing a block", func() {
		var initSupply sdk.Coin
		BeforeEach(func() {
			params = s.app.MintKeeper.GetParams(s.ctx)
			foundationAddr = sdk.MustAccAddressFromBech32(params.FoundationAddress)
			stakingModuleAddr = s.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			communityAddr = s.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
			mintModuleAddr = s.app.AccountKeeper.GetModuleAddress(types.ModuleName)

			initSupply = s.app.BankKeeper.GetSupply(s.ctx, denomMint)
			Expect(initSupply.Amount).To(Equal(initialSupply))

			Expect(s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, params.MintDenom).Amount).To(
				Equal(sdk.ZeroInt()),
			)
			Expect(s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, params.MintDenom).Amount).To(
				Equal(sdk.ZeroInt()),
			)
			Expect(s.app.BankKeeper.GetBalance(s.ctx, communityAddr, params.MintDenom).Amount).To(
				Equal(sdk.ZeroInt()),
			)

			Expect(s.app.MintKeeper.BondedRatio(s.ctx)).To(Equal(sdk.ZeroDec()))
		})

		Context("On the first block after genesis", func() {
			BeforeEach(func() {
				oldMinter = s.app.MintKeeper.GetMinter(s.ctx)
				oldSupply = s.app.BankKeeper.GetSupply(s.ctx, params.MintDenom).Amount
				oldBondedRatio = s.app.MintKeeper.BondedRatio(s.ctx)

				oldStakingModuleBalance = s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, params.MintDenom).Amount
				oldFoundationBalance = s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, params.MintDenom).Amount
				oldCommunityBalance = s.app.BankKeeper.GetBalance(s.ctx, communityAddr, params.MintDenom).Amount

				s.CommitAndBeginBlock()
			})

			It("total supply should change", func() {
				nextInflationRate := oldMinter.NextInflationRate(params, oldBondedRatio)
				expSupplyIncrease := sdk.NewDecFromInt(oldSupply).Mul(nextInflationRate).QuoInt64(
					int64(params.InflationParameters.BlocksPerYear)).TruncateInt()
				newSupply := s.app.BankKeeper.GetSupply(s.ctx, denomMint)

				Expect(newSupply.Amount.String()).To(Equal(oldSupply.Add(expSupplyIncrease).String()))
			})

			It("inflationRate and annualProvisions should change", func() {
				expectedInflationRate := oldMinter.NextInflationRate(params, oldBondedRatio)
				expectedAnnualProvision := expectedInflationRate.MulInt(oldSupply)
				newMinter := s.app.MintKeeper.GetMinter(s.ctx)

				Expect(newMinter.Inflation).To(Equal(expectedInflationRate))
				Expect(newMinter.AnnualProvisions).To(Equal(expectedAnnualProvision))
			})

			It("mint module balance must be equal to zero", func() {
				for i := 0; i < numTests; i++ {
					s.CommitAndBeginBlock()
					Expect(s.app.BankKeeper.GetBalance(s.ctx, mintModuleAddr, params.MintDenom).Amount).To(Equal(sdk.ZeroInt()))
				}
			})

			It("inflation should be distributed correctly", func() {
				nextInflationRate := oldMinter.NextInflationRate(params, oldBondedRatio)
				blockProvision := sdk.NewDecFromInt(oldSupply).Mul(nextInflationRate).QuoInt64(
					int64(params.InflationParameters.BlocksPerYear)).TruncateInt()

				expStakingIncreased := params.InflationDistribution.StakingRewards.MulInt(blockProvision).TruncateInt()
				Expect(s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, params.MintDenom).Amount).To(
					Equal(oldStakingModuleBalance.Add(expStakingIncreased)),
				)

				expFoundationIncreased := params.InflationDistribution.Foundation.MulInt(blockProvision).TruncateInt()
				Expect(s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, params.MintDenom).Amount).To(
					Equal(oldFoundationBalance.Add(expFoundationIncreased)),
				)

				expCommunityIncreased := blockProvision.Sub(expStakingIncreased).Sub(expFoundationIncreased)
				Expect(s.app.BankKeeper.GetBalance(s.ctx, communityAddr, params.MintDenom).Amount).To(
					Equal(oldCommunityBalance.Add(expCommunityIncreased)),
				)
			})
		})

		Context("bondedRatio < goalBonded", func() {
			BeforeEach(func() {
				s.bondWithRate(params.MintDenom, randInflationRate(sdk.ZeroDec(), params.InflationParameters.GoalBonded))
				s.CommitAndBeginBlocks(rand.Int63n(50))
				oldMinter = s.app.MintKeeper.GetMinter(s.ctx)

			})

			It("mint module balance must be equal to zero", func() {
				for i := 0; i < numTests; i++ {
					s.CommitAndBeginBlock()
					Expect(s.app.BankKeeper.GetBalance(s.ctx, mintModuleAddr, params.MintDenom).Amount).To(Equal(sdk.ZeroInt()))
				}
			})

			When("inflationRate < inflationMax", func() {
				It("inflationRate should increase until it reaches inflationMax", func() {
					// reset the new minter with inflationRate nearly reaches inflationMax
					oldMinter.Inflation = params.InflationParameters.InflationMax.Mul(sdk.MustNewDecFromStr("0.9999"))
					s.app.MintKeeper.SetMinter(s.ctx, oldMinter)
					for {
						s.CommitAndBeginBlocks(1)
						newMinter := s.app.MintKeeper.GetMinter(s.ctx)

						Expect(newMinter.Inflation.LTE(params.InflationParameters.InflationMax)).To(BeTrue())
						Expect(newMinter.Inflation.GTE(oldMinter.Inflation)).To(BeTrue())

						if newMinter.Inflation.Equal(oldMinter.Inflation) {
							break
						}
						oldMinter = newMinter
					}
				})
			})

			When("inflationRate = inflationMax", func() {
				It("inflationRate should stay unchanged", func() {
					// set inflationRate = inflationMax
					oldMinter.Inflation = params.InflationParameters.InflationMax
					s.app.MintKeeper.SetMinter(s.ctx, oldMinter)

					s.CommitAndBeginBlocks(1)
					newMinter := s.app.MintKeeper.GetMinter(s.ctx)
					Expect(newMinter.Inflation.Equal(params.InflationParameters.InflationMax)).To(BeTrue())
					Expect(newMinter.Inflation.Equal(oldMinter.Inflation)).To(BeTrue())
				})
			})

			When("increase bondedRatio", func() {
				It("should decrease inflationRateChange", func() {
					s.CommitAndBeginBlock()
					newMinter := s.app.MintKeeper.GetMinter(s.ctx)
					oldRateChange := newMinter.Inflation.Sub(oldMinter.Inflation)
					oldMinter = newMinter
					oldBondedRatio = s.app.MintKeeper.BondedRatio(s.ctx)
					for {
						// increase the bondedRatio
						newBondedRatio := randInflationRate(
							oldBondedRatio, sdk.OneDec(),
						)
						if newBondedRatio.Equal(params.InflationParameters.GoalBonded) {
							break
						}
						s.bondWithRate(params.MintDenom, newBondedRatio)
						s.CommitAndBeginBlock()

						newBondedRatio = s.app.MintKeeper.BondedRatio(s.ctx)
						if newBondedRatio.LT(oldBondedRatio) {
							break
						}

						// retrieve the new minter
						newMinter = s.app.MintKeeper.GetMinter(s.ctx)
						newRateChange := newMinter.Inflation.Sub(oldMinter.Inflation)
						Expect(newRateChange.LTE(oldRateChange)).To(BeTrue())

						// reset
						oldMinter = newMinter
						oldBondedRatio = newBondedRatio
						oldRateChange = newRateChange
					}
				})
			})

		})

		Context("bondedRatio > goalBonded", func() {
			BeforeEach(func() {
				s.bondWithRate(params.MintDenom, randInflationRate(params.InflationParameters.GoalBonded, sdk.OneDec()))
				s.CommitAndBeginBlocks(rand.Int63n(50))
				oldMinter = s.app.MintKeeper.GetMinter(s.ctx)

			})

			It("mint module balance must be equal to zero", func() {
				for i := 0; i < numTests; i++ {
					s.CommitAndBeginBlock()
					Expect(s.app.BankKeeper.GetBalance(s.ctx, mintModuleAddr, params.MintDenom).Amount).To(Equal(sdk.ZeroInt()))
				}
			})

			When("inflationRate > inflationMin", func() {
				It("inflationRate should decrease until it reaches inflationMin", func() {
					// reset the new minter with inflationRate nearly reaches inflationMax
					oldMinter.Inflation = params.InflationParameters.InflationMin.Mul(sdk.MustNewDecFromStr("1.00001"))
					s.app.MintKeeper.SetMinter(s.ctx, oldMinter)
					for {
						s.CommitAndBeginBlocks(1)
						newMinter := s.app.MintKeeper.GetMinter(s.ctx)

						Expect(newMinter.Inflation.GTE(params.InflationParameters.InflationMin)).To(BeTrue())
						Expect(newMinter.Inflation.LTE(oldMinter.Inflation)).To(BeTrue())

						if newMinter.Inflation.Equal(oldMinter.Inflation) {
							break
						}
						oldMinter = newMinter
					}
				})
			})

			When("inflationRate = inflationMin", func() {
				It("inflationRate should stay unchanged", func() {
					// set inflationRate = inflationMax
					oldMinter.Inflation = params.InflationParameters.InflationMin
					s.app.MintKeeper.SetMinter(s.ctx, oldMinter)

					s.CommitAndBeginBlocks(1)
					newMinter := s.app.MintKeeper.GetMinter(s.ctx)
					Expect(newMinter.Inflation.Equal(params.InflationParameters.InflationMin)).To(BeTrue())
					Expect(newMinter.Inflation.Equal(oldMinter.Inflation)).To(BeTrue())
				})
			})

			When("decrease bondedRatio", func() {
				It("should increase inflationRateChange", func() {
					s.CommitAndBeginBlock()
					newMinter := s.app.MintKeeper.GetMinter(s.ctx)
					oldRateChange := newMinter.Inflation.Sub(oldMinter.Inflation)
					oldMinter = newMinter
					oldBondedRatio = s.app.MintKeeper.BondedRatio(s.ctx)
					for {
						if oldBondedRatio.LT(params.InflationParameters.GoalBonded) {
							break
						}

						// increase the bondedRatio
						newBondedRatio := randInflationRate(
							params.InflationParameters.GoalBonded, oldBondedRatio,
						)
						if newBondedRatio.Equal(params.InflationParameters.GoalBonded) {
							break
						}
						s.bondWithRate(params.MintDenom, newBondedRatio)
						s.CommitAndBeginBlock()

						newBondedRatio = s.app.MintKeeper.BondedRatio(s.ctx)
						if newBondedRatio.GT(oldBondedRatio) {
							break
						}

						// retrieve the new minter
						newMinter = s.app.MintKeeper.GetMinter(s.ctx)
						newRateChange := newMinter.Inflation.Sub(oldMinter.Inflation)
						Expect(newRateChange.GTE(oldRateChange)).To(BeTrue())

						// reset
						oldMinter = newMinter
						oldBondedRatio = newBondedRatio
						oldRateChange = newRateChange
					}
				})
			})

		})

		Context("bondedRatio = goalBonded", func() {
			BeforeEach(func() {
				s.SetupTest()
				s.mintAndBondWithRate(params.MintDenom, initialSupply, params.InflationParameters.GoalBonded)
			})

			It("mint module balance must be equal to zero", func() {
				for i := 0; i < numTests; i++ {
					s.CommitAndBeginBlock()
					Expect(s.app.BankKeeper.GetBalance(s.ctx, mintModuleAddr, params.MintDenom).Amount).To(Equal(sdk.ZeroInt()))
				}
			})

			It("inflationRate mush stay unchanged", func() {
				oldMinter := s.app.MintKeeper.GetMinter(s.ctx)
				s.CommitAndBeginBlock()
				newMinter := s.app.MintKeeper.GetMinter(s.ctx)
				Expect(newMinter.Inflation).To(Equal(oldMinter.Inflation))
			})
		})
	})
})

// ======================================== helper functions ========================================

func (suite *KeeperTestSuite) mintAndTransfer(denom string, mintAmount sdk.Int, recipient sdk.AccAddress, transferAmount sdk.Int) {
	err := suite.app.MintKeeper.MintCoins(suite.ctx, sdk.NewCoin(denom, mintAmount))
	Expect(err).To(BeNil())
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, recipient,
		sdk.NewCoins(sdk.NewCoin(denom, transferAmount)))
	Expect(err).To(BeNil())
}

// bondWithRate mints more coins and transfers them to/or burns from the BondedPool module to make sure the bondedRatio ~~ rate.
func (suite *KeeperTestSuite) bondWithRate(denom string, rate sdk.Dec) {
	currentSupply := suite.app.BankKeeper.GetSupply(suite.ctx, denom)
	oldBonded := suite.app.StakingKeeper.TotalBondedTokens(suite.ctx)
	mintAmount := rate.MulInt(currentSupply.Amount).Sub(oldBonded.ToDec()).Quo(sdk.OneDec().Sub(rate)).TruncateInt()

	// should stake more
	if mintAmount.IsPositive() {
		err := suite.app.MintKeeper.MintCoins(suite.ctx, sdk.NewCoin(denom, mintAmount))
		Expect(err).To(BeNil())
		err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, types.ModuleName, stakingtypes.BondedPoolName,
			sdk.NewCoins(sdk.NewCoin(denom, mintAmount)))
		Expect(err).To(BeNil())
	} else { // should burn
		err := suite.app.BankKeeper.BurnCoins(suite.ctx, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(denom, mintAmount.Abs())))
		Expect(err).To(BeNil())
	}

	newSupply := suite.app.BankKeeper.GetSupply(suite.ctx, denom)
	Expect(newSupply.Amount).To(Equal(currentSupply.Amount.Add(mintAmount)))
}

// mintAndBondWithRate mints `mintAmount` of coins and sends `rate * mintAmount` to the BondedPool module.
func (suite *KeeperTestSuite) mintAndBondWithRate(denom string, mintAmount sdk.Int, rate sdk.Dec) {
	if rate.GT(sdk.OneDec()) {
		rate = sdk.OneDec()
	}
	err := suite.app.MintKeeper.MintCoins(suite.ctx, sdk.NewCoin(denom, mintAmount))
	Expect(err).To(BeNil())
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, types.ModuleName, stakingtypes.BondedPoolName,
		sdk.NewCoins(sdk.NewCoin(denom, mintAmount.ToDec().Mul(rate).TruncateInt())))
	Expect(err).To(BeNil())

	Expect(suite.app.BankKeeper.GetSupply(suite.ctx, denom).Amount).To(Equal(mintAmount))
	Expect(suite.app.MintKeeper.BondedRatio(suite.ctx)).To(Equal(rate))
}

func randInflationRate(minRate sdk.Dec, maxRate sdk.Dec) sdk.Dec {
	base := int64(1000000000000000000)
	min := minRate.MulInt64(base).Ceil().TruncateInt64()
	max := maxRate.MulInt64(base).TruncateInt64()

	if max-min < 0 {
		return minRate
	}
	rate := min + rand.Int63n(max-min)

	return sdk.NewDec(rate).QuoInt64(base)
}
