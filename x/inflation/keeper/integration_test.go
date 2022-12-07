package keeper_test

import (
	"github.com/AstraProtocol/astra/v2/x/inflation/types"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	epochstypes "github.com/evmos/evmos/v6/x/epochs/types"
)

var (
	epochNumber int64
	skipped     uint64
	provision   sdk.Dec
	found       bool
)

var _ = Describe("Inflation", Ordered, func() {
	BeforeEach(func() {
		s.SetupTest()
	})

	var params types.Params
	var foundationAddr sdk.AccAddress
	var stakingModuleAddr sdk.AccAddress
	var communityAddr sdk.AccAddress
	var oldFoundationBalance sdk.Coin
	var oldStakingModuleBalance sdk.Coin
	var oldCommunityBalance sdk.Coin

	Describe("Committing a block", func() {
		var initSupply sdk.Dec
		genesisProvision := sdk.MustNewDecFromStr("569863013698630136986301.000000000000000000")
		Context("with inflation param enabled", func() {
			BeforeEach(func() {
				initSupply = s.app.InflationKeeper.GetCirculatingSupply(s.ctx)
				params = s.app.InflationKeeper.GetParams(s.ctx)
				params.EnableInflation = true
				s.app.InflationKeeper.SetParams(s.ctx, params)

				foundationAddr = sdk.MustAccAddressFromBech32(params.FoundationAddress)
				stakingModuleAddr = s.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
				communityAddr = s.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
			})

			Context("before an epoch ends", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute)    // Start Epoch
					s.CommitAfter(time.Hour * 23) // End Epoch
				})
				It("should equal init supply", func() {
					supply := s.app.InflationKeeper.GetCirculatingSupply(s.ctx)
					Expect(supply).To(Equal(initSupply))
				})

				It("should allocate no provision for inflation distribution component", func() {
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, denomMint).Amount.Uint64(),
					).To(Equal(uint64(0)))

					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, denomMint).Amount.Uint64(),
					).To(Equal(uint64(0)))

					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, communityAddr, denomMint).Amount.Uint64(),
					).To(Equal(uint64(0)))
				})
			})

			Context("after an epoch ends", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute) // Start Epoch

					oldFoundationBalance = s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, denomMint)
					oldStakingModuleBalance = s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, denomMint)
					oldCommunityBalance = s.app.BankKeeper.GetBalance(s.ctx, communityAddr, denomMint)

					s.CommitAfter(time.Hour*24 - time.Minute + 1) // End Epoch
				})

				It("should release token to block reward", func() {
					supply := s.app.InflationKeeper.GetCirculatingSupply(s.ctx)
					Expect(supply).To(Equal(genesisProvision.Add(initSupply)))
				})

				It("total minted token must be equal block reward", func() {
					totalMintedProvision := s.app.InflationKeeper.GetTotalMintProvision(s.ctx)
					Expect(totalMintedProvision).To(Equal(genesisProvision))
				})

				It("should allocate correct portions for inflation distribution with period = 0", func() {
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, denomMint).Sub(oldFoundationBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("56986301369863013698630").TruncateInt())))

					// The staking reward will be sent to the `distribution` module afterwards via the function `AllocateTokens`
					// called in `BeginBlocker` of the `distribution module (i.e, inflation.AfterEpochEnd -> distribution.BeginBlocker).
					// Therefore, the distribution for `community` also accounts for the amount sent to the `feeCollector` modue,
					// and the balance of `feeCollector` is always 0.
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, communityAddr, denomMint).Sub(oldCommunityBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("512876712328767123287671").TruncateInt())))
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, denomMint).Sub(oldStakingModuleBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("0").TruncateInt())))
				})
			})

			Context("after two epoch ends", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute)    // Start Epoch
					s.CommitAfter(time.Hour * 24) // End 1 Epoch

					oldFoundationBalance = s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, denomMint)
					oldStakingModuleBalance = s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, denomMint)
					oldCommunityBalance = s.app.BankKeeper.GetBalance(s.ctx, communityAddr, denomMint)

					s.CommitAfter(time.Hour * 50) // End 2 Epoch
				})

				It("should release token to block reward 569863013698630136986301 * 2", func() {
					supply := s.app.InflationKeeper.GetCirculatingSupply(s.ctx)
					supplyAfter2Epoch := sdk.MustNewDecFromStr("1139726027397260273972602")
					Expect(supply).To(Equal(supplyAfter2Epoch))
				})

				It("total minted token must be correct", func() {
					totalMintedProvision := s.app.InflationKeeper.GetTotalMintProvision(s.ctx)
					Expect(totalMintedProvision).To(Equal(sdk.MustNewDecFromStr("1139726027397260273972602")))
				})

				It("should allocate correct portions for inflation distribution with period = 0", func() {
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, denomMint).Sub(oldFoundationBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("56986301369863013698630").TruncateInt())))

					// The staking reward will be sent to the `distribution` module afterwards via the function `AllocateTokens`
					// called in `BeginBlocker` of the `distribution module (i.e, inflation.AfterEpochEnd -> distribution.BeginBlocker).
					// Therefore, the distribution for `community` also accounts for the amount sent to the `feeCollector` modue,
					// and the balance of `feeCollector` is always 0.
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, communityAddr, denomMint).Sub(oldCommunityBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("512876712328767123287671").TruncateInt())))
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, denomMint).Sub(oldStakingModuleBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("0").TruncateInt())))
				})
			})
			Context("after 365 epoch ends", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute) // Start Epoch
					for i := 1; i < 366; i++ {
						t := 24 * i

						if i == 365 {
							oldFoundationBalance = s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, denomMint)
							oldStakingModuleBalance = s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, denomMint)
							oldCommunityBalance = s.app.BankKeeper.GetBalance(s.ctx, communityAddr, denomMint)
						}

						s.CommitAfter(time.Hour * time.Duration(t)) // End Epoch i
					}
				})

				It("should release token to block reward 569863013698630136986301 * 365", func() {
					supply := s.app.InflationKeeper.GetCirculatingSupply(s.ctx)
					supplyAfter2Epoch := sdk.MustNewDecFromStr("207999999999999999999999865")
					Expect(supply).To(Equal(supplyAfter2Epoch))
				})

				It("total minted token must be correct", func() {
					totalMintedProvision := s.app.InflationKeeper.GetTotalMintProvision(s.ctx)
					Expect(totalMintedProvision).To(Equal(sdk.MustNewDecFromStr("207999999999999999999999865")))
				})

				It("should allocate correct portions for inflation distribution with period = 0", func() {
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, denomMint).Sub(oldFoundationBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("56986301369863013698630").TruncateInt())))

					// The staking reward will be sent to the `distribution` module afterwards via the function `AllocateTokens`
					// called in `BeginBlocker` of the `distribution module (i.e, inflation.AfterEpochEnd -> distribution.BeginBlocker).
					// Therefore, the distribution for `community` also accounts for the amount sent to the `feeCollector` modue,
					// and the balance of `feeCollector` is always 0.
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, communityAddr, denomMint).Sub(oldCommunityBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("512876712328767123287671").TruncateInt())))
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, denomMint).Sub(oldStakingModuleBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("0").TruncateInt())))
				})
			})
			Context("after 366 epoch ends", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute) // Start Epoch
					for i := 1; i < 367; i++ {
						t := 24 * i

						if i == 366 {
							oldFoundationBalance = s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, denomMint)
							oldStakingModuleBalance = s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, denomMint)
							oldCommunityBalance = s.app.BankKeeper.GetBalance(s.ctx, communityAddr, denomMint)
						}

						s.CommitAfter(time.Hour * time.Duration(t)) // End Epoch i
					}
				})

				It("should release token to block reward, 569863013698630136986301 * 365 + 569863013698630136986301 * 0.74", func() {
					supply := s.app.InflationKeeper.GetCirculatingSupply(s.ctx)
					supplyAfter2Epoch := sdk.MustNewDecFromStr("208421698630136986301369728")
					Expect(supply).To(Equal(supplyAfter2Epoch))
				})

				It("total minted token must be correct", func() {
					totalMintedProvision := s.app.InflationKeeper.GetTotalMintProvision(s.ctx)
					Expect(totalMintedProvision).To(Equal(sdk.MustNewDecFromStr("208421698630136986301369728")))
				})

				It("should allocate correct portions for inflation distribution with period = 1", func() {
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, denomMint).Sub(oldFoundationBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("42169863013698630136986").TruncateInt())))

					// The staking reward will be sent to the `distribution` module afterwards via the function `AllocateTokens`
					// called in `BeginBlocker` of the `distribution module (i.e, inflation.AfterEpochEnd -> distribution.BeginBlocker).
					// Therefore, the distribution for `community` also accounts for the amount sent to the `feeCollector` modue,
					// and the balance of `feeCollector` is always 0.
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, communityAddr, denomMint).Sub(oldCommunityBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("379528767123287671232877").TruncateInt())))
					Expect(
						s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, denomMint).Sub(oldStakingModuleBalance),
					).To(Equal(sdk.NewCoin(denomMint, sdk.MustNewDecFromStr("0").TruncateInt())))
				})
			})
		})

		Context("with inflation param disabled", func() {
			BeforeEach(func() {
				params := s.app.InflationKeeper.GetParams(s.ctx)
				params.EnableInflation = false
				s.app.InflationKeeper.SetParams(s.ctx, params)

				foundationAddr = sdk.MustAccAddressFromBech32(params.FoundationAddress)
				stakingModuleAddr = s.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
				communityAddr = s.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
			})

			Context("after the network was offline for several days/epochs", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute)        // start initial epoch
					s.CommitAfter(time.Hour * 24 * 5) // end epoch after several days
				})
				When("the epoch start time has not caught up with the block time", func() {
					BeforeEach(func() {
						// commit next 3 blocks to trigger afterEpochEnd let EpochStartTime
						// catch up with BlockTime
						s.CommitAfter(time.Second * 3)
						s.CommitAfter(time.Second * 3)
						s.CommitAfter(time.Second * 3)

						epochInfo, found := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						s.Require().True(found)
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.app.InflationKeeper.GetSkippedEpochs(s.ctx)

						s.CommitAfter(time.Second * 6) // commit next block
					})
					It("should increase the epoch number ", func() {
						epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						Expect(epochInfo.CurrentEpoch).To(Equal(epochNumber + 1))
					})
					It("should not increase the skipped epochs number", func() {
						skippedAfter := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
						Expect(skippedAfter).To(Equal(skipped + 1))
					})
					It("should allocate no provision for inflation distribution component", func() {
						Expect(
							s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, denomMint).Amount.Uint64(),
						).To(Equal(uint64(0)))

						Expect(
							s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, denomMint).Amount.Uint64(),
						).To(Equal(uint64(0)))

						Expect(
							s.app.BankKeeper.GetBalance(s.ctx, communityAddr, denomMint).Amount.Uint64(),
						).To(Equal(uint64(0)))
					})
				})

				When("the epoch start time has caught up with the block time", func() {
					BeforeEach(func() {
						// commit next 4 blocks to trigger afterEpochEnd hook several times
						// and let EpochStartTime catch up with BlockTime
						s.CommitAfter(time.Second * 3)
						s.CommitAfter(time.Second * 3)
						s.CommitAfter(time.Second * 3)
						s.CommitAfter(time.Second * 3)

						epochInfo, found := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						s.Require().True(found)
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.app.InflationKeeper.GetSkippedEpochs(s.ctx)

						s.CommitAfter(time.Second * 3) // commit next block
					})
					It("should not increase the epoch number ", func() {
						epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						Expect(epochInfo.CurrentEpoch).To(Equal(epochNumber))
					})
					It("should not increase the skipped epochs number", func() {
						skippedAfter := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
						Expect(skippedAfter).To(Equal(skipped))
					})
					It("should allocate no provision for inflation distribution component", func() {
						Expect(
							s.app.BankKeeper.GetBalance(s.ctx, foundationAddr, denomMint).Amount.Uint64(),
						).To(Equal(uint64(0)))

						Expect(
							s.app.BankKeeper.GetBalance(s.ctx, stakingModuleAddr, denomMint).Amount.Uint64(),
						).To(Equal(uint64(0)))

						Expect(
							s.app.BankKeeper.GetBalance(s.ctx, communityAddr, denomMint).Amount.Uint64(),
						).To(Equal(uint64(0)))
					})

					When("epoch number passes epochsPerPeriod + skippedEpochs and inflation re-enabled", func() {
						BeforeEach(func() {
							params := s.app.InflationKeeper.GetParams(s.ctx)
							params.EnableInflation = true
							s.app.InflationKeeper.SetParams(s.ctx, params)

							epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
							epochNumber := epochInfo.CurrentEpoch // 6

							epochsPerPeriod := int64(1)
							s.app.InflationKeeper.SetEpochsPerPeriod(s.ctx, epochsPerPeriod)
							skipped := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
							s.Require().Equal(epochNumber, epochsPerPeriod+int64(skipped))

							provision, found = s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							s.Require().True(found)

							s.CommitAfter(time.Hour * 23) // commit before next full epoch
							provisionAfter, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							s.Require().Equal(provisionAfter, provision)

							s.CommitAfter(time.Hour * 2) // commit after next full epoch
						})

						It("should recalculate the EpochMintProvision", func() {
							provisionAfter, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							Expect(provisionAfter).ToNot(Equal(provision))
							Expect(provisionAfter).To(Equal(sdk.MustNewDecFromStr("153920000000000000000000000"))) // = periodProvision since epochsPerPeriod := int64(1)
						})
					})
				})
			})
		})
	})
})
