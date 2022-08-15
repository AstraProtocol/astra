package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

	Describe("Committing a block", func() {
		initSupply := s.app.InflationKeeper.GetCirculatingSupply(s.ctx)
		genesisProvision := sdk.MustNewDecFromStr("304410958904109589041095.000000000000000000")
		Context("with inflation param enabled", func() {
			BeforeEach(func() {
				params := s.app.InflationKeeper.GetParams(s.ctx)
				params.EnableInflation = true
				s.app.InflationKeeper.SetParams(s.ctx, params)
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
			})

			Context("after an epoch ends", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute)    // Start Epoch
					s.CommitAfter(time.Hour * 25) // End Epoch
				})

				It("should release token to block reward", func() {
					supply := s.app.InflationKeeper.GetCirculatingSupply(s.ctx)
					Expect(supply).To(Equal(genesisProvision))
				})
			})

			Context("after two epoch ends", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute)    // Start Epoch
					s.CommitAfter(time.Hour * 24) // End 1 Epoch
					s.CommitAfter(time.Hour * 50) // End 2 Epoch
				})

				It("should release token to block reward 304410958904109589041095 * 2", func() {
					supply := s.app.InflationKeeper.GetCirculatingSupply(s.ctx)
					supplyAfter2Epoch := sdk.MustNewDecFromStr("608821917808219178082190")
					Expect(supply).To(Equal(supplyAfter2Epoch))
				})
			})
			Context("after 365 epoch ends", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute) // Start Epoch
					for i := 1; i < 366; i++ {
						t := 24 * i
						s.CommitAfter(time.Hour * time.Duration(t)) // End Epoch i
					}
				})

				It("should release token to block reward 304410958904109589041095 * 365", func() {
					supply := s.app.InflationKeeper.GetCirculatingSupply(s.ctx)
					supplyAfter2Epoch := sdk.MustNewDecFromStr("111109999999999999999999675")
					Expect(supply).To(Equal(supplyAfter2Epoch))
				})
			})
			Context("after 366 epoch ends", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute) // Start Epoch
					for i := 1; i < 367; i++ {
						t := 24 * i
						s.CommitAfter(time.Hour * time.Duration(t)) // End Epoch i
					}
				})

				It("should release token to block reward, 304410958904109589041095 * 365 + 304410958904109589041095 * 0.95", func() {
					supply := s.app.InflationKeeper.GetCirculatingSupply(s.ctx)
					supplyAfter2Epoch := sdk.MustNewDecFromStr("111399190410958904109588716")
					Expect(supply).To(Equal(supplyAfter2Epoch))
				})
			})
		})

		Context("with inflation param disabled", func() {
			BeforeEach(func() {
				params := s.app.InflationKeeper.GetParams(s.ctx)
				params.EnableInflation = false
				s.app.InflationKeeper.SetParams(s.ctx, params)
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
							Expect(provisionAfter).To(Equal(sdk.MustNewDecFromStr("105554500000000000000000000")))
						})
					})
				})
			})
		})
	})
})
