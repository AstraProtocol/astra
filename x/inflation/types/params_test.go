package types

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	"github.com/AstraProtocol/astra/v2/types"
	"math"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"
)

type ParamsTestSuite struct {
	suite.Suite
}

func TestParamsTestSuite(t *testing.T) {
	config.SetBech32Prefixes(sdk.GetConfig())
	suite.Run(t, new(ParamsTestSuite))
}

func (suite *ParamsTestSuite) TestParamKeyTable() {
	suite.Require().IsType(paramtypes.KeyTable{}, ParamKeyTable())
}

func (suite *ParamsTestSuite) TestParamsValidate() {
	validInflationParameters := InflationParameters{
		MaxStakingRewards: sdk.NewDec(800000000),
		R:                 sdk.NewDecWithPrec(26, 2),
	}
	validInflationDistribution := InflationDistribution{
		StakingRewards: sdk.NewDecWithPrec(88, 2),
		Foundation:     sdk.NewDecWithPrec(10, 2),
		CommunityPool:  sdk.NewDecWithPrec(2, 2),
	}
	foundationAddr := "astra13wjs7d3z8hra6rp7vjmryuulwxjrd232sceuen"

	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{
			"default",
			DefaultParams(),
			false,
		},
		{
			"valid param",
			NewParams(
				config.DisplayDenom,
				validInflationParameters,
				validInflationDistribution,
				foundationAddr,
				true,
			),
			false,
		},
		{
			"valid param - enableInflation = false",
			NewParams(
				config.DisplayDenom,
				validInflationParameters,
				validInflationDistribution,
				foundationAddr,
				false,
			),
			false,
		},
		{
			"valid param literal",
			Params{
				MintDenom:             config.BaseDenom,
				InflationParameters:   validInflationParameters,
				InflationDistribution: validInflationDistribution,
				FoundationAddress:     foundationAddr,
				EnableInflation:       true,
			},
			false,
		},
		{
			"invalid - denom with backslash",
			NewParams(
				"/aastra",
				validInflationParameters,
				validInflationDistribution,
				foundationAddr,
				true,
			),
			true,
		},
		{
			"invalid - empty denom",
			Params{
				MintDenom:             "",
				InflationParameters:   validInflationParameters,
				InflationDistribution: validInflationDistribution,
				FoundationAddress:     foundationAddr,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - not allowed denom",
			Params{
				MintDenom:             "aaastra",
				InflationParameters:   validInflationParameters,
				InflationDistribution: validInflationDistribution,
				FoundationAddress:     foundationAddr,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - inflation parameters - R greater than 1",
			Params{
				MintDenom: config.BaseDenom,
				InflationParameters: InflationParameters{
					MaxStakingRewards: validInflationParameters.MaxStakingRewards,
					R:                 sdk.NewDecWithPrec(5, 0),
				},
				InflationDistribution: validInflationDistribution,
				FoundationAddress:     foundationAddr,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - inflation parameters - negative R",
			Params{
				MintDenom: config.BaseDenom,
				InflationParameters: InflationParameters{
					MaxStakingRewards: validInflationParameters.MaxStakingRewards,
					R:                 sdk.NewDecWithPrec(-5, 1),
				},
				InflationDistribution: validInflationDistribution,
				FoundationAddress:     foundationAddr,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - inflation parameters - negative maximum staking rewards < -1",
			Params{
				MintDenom: config.BaseDenom,
				InflationParameters: InflationParameters{
					MaxStakingRewards: sdk.NewDec(-100101010101),
					R:                 sdk.NewDecWithPrec(-5, 1),
				},
				InflationDistribution: validInflationDistribution,
				FoundationAddress:     foundationAddr,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - inflation parameters - maximum staking rewards == 0",
			Params{
				MintDenom: config.BaseDenom,
				InflationParameters: InflationParameters{
					MaxStakingRewards: sdk.NewDec(0),
					R:                 sdk.NewDecWithPrec(-5, 1),
				},
				InflationDistribution: validInflationDistribution,
				FoundationAddress:     foundationAddr,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - inflation distribution - staking reward < 0",
			Params{
				MintDenom:           config.BaseDenom,
				InflationParameters: validInflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(-1, 2),
					Foundation:     validInflationDistribution.Foundation,
					CommunityPool:  validInflationDistribution.CommunityPool,
				},
				FoundationAddress: foundationAddr,
				EnableInflation:   true,
			},
			true,
		},
		{
			"invalid - inflation distribution - staking reward > 1",
			Params{
				MintDenom:           config.BaseDenom,
				InflationParameters: validInflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(101, 2),
					Foundation:     validInflationDistribution.Foundation,
					CommunityPool:  validInflationDistribution.CommunityPool,
				},
				FoundationAddress: foundationAddr,
				EnableInflation:   true,
			},
			true,
		},
		{
			"invalid - inflation distribution - foundation < 0",
			Params{
				MintDenom:           config.BaseDenom,
				InflationParameters: validInflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: validInflationDistribution.StakingRewards,
					Foundation:     sdk.NewDecWithPrec(-1, 2),
					CommunityPool:  validInflationDistribution.CommunityPool,
				},
				FoundationAddress: foundationAddr,
				EnableInflation:   true,
			},
			true,
		},
		{
			"invalid - inflation distribution - foundation > 1",
			Params{
				MintDenom:           config.BaseDenom,
				InflationParameters: validInflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: validInflationDistribution.StakingRewards,
					Foundation:     sdk.NewDecWithPrec(101, 2),
					CommunityPool:  validInflationDistribution.CommunityPool,
				},
				FoundationAddress: foundationAddr,
				EnableInflation:   true,
			},
			true,
		},
		{
			"invalid - inflation distribution - community < 0",
			Params{
				MintDenom:           config.BaseDenom,
				InflationParameters: validInflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: validInflationDistribution.StakingRewards,
					Foundation:     validInflationDistribution.Foundation,
					CommunityPool:  sdk.NewDecWithPrec(-1, 2),
				},
				FoundationAddress: foundationAddr,
				EnableInflation:   true,
			},
			true,
		},
		{
			"invalid - inflation distribution - community > 1",
			Params{
				MintDenom:           config.BaseDenom,
				InflationParameters: validInflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: validInflationDistribution.StakingRewards,
					Foundation:     validInflationDistribution.Foundation,
					CommunityPool:  sdk.NewDecWithPrec(101, 2),
				},
				FoundationAddress: foundationAddr,
				EnableInflation:   true,
			},
			true,
		},
		{
			"invalid - inflation distribution - wrong total proportion",
			Params{
				MintDenom:           config.BaseDenom,
				InflationParameters: validInflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(80, 2),
					Foundation:     sdk.NewDecWithPrec(10, 2),
					CommunityPool:  sdk.NewDecWithPrec(20, 2),
				},
				FoundationAddress: foundationAddr,
				EnableInflation:   true,
			},
			true,
		},
		{
			"valid param - inflation distribution - correct total proportion",
			Params{
				MintDenom:           config.BaseDenom,
				InflationParameters: validInflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(50, 2),
					Foundation:     sdk.NewDecWithPrec(30, 2),
					CommunityPool:  sdk.NewDecWithPrec(20, 2),
				},
				FoundationAddress: foundationAddr,
				EnableInflation:   true,
			},
			false,
		},
		{
			"invalid - foundation address - empty address",
			Params{
				MintDenom:             config.BaseDenom,
				InflationParameters:   validInflationParameters,
				InflationDistribution: validInflationDistribution,
				FoundationAddress:     "",
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - foundation address - invalid prefix",
			Params{
				MintDenom:             config.BaseDenom,
				InflationParameters:   validInflationParameters,
				InflationDistribution: validInflationDistribution,
				FoundationAddress:     "cosmos13wjs7d3z8hra6rp7vjmryuulwxjrd232sceuen",
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - foundation address - invalid address",
			Params{
				MintDenom:             config.BaseDenom,
				InflationParameters:   validInflationParameters,
				InflationDistribution: validInflationDistribution,
				FoundationAddress:     "astra13wjs7d3z8hra6rp7vjmryuulwxjrd232scecen",
				EnableInflation:       true,
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}

	numRandTests := 1000
	// randomized tests
	for i := 0; i < numRandTests; i++ {
		validParam := newRandomValidParam()
		err := validParam.Validate()
		jsb, _ := json.Marshal(validParam)
		suite.Require().NoError(err, fmt.Sprintf("expect param to be valid: %v", string(jsb)))
	}

	for i := 0; i < numRandTests; i++ {
		invalidParam := newRandomInvalidParam()
		err := invalidParam.Validate()
		jsb, _ := json.Marshal(invalidParam)
		suite.Require().Error(err, fmt.Sprintf("expect param to be invalid: %v", string(jsb)))
	}
}

// newRandomValidParam creates a new randomized valid parameter.
func newRandomValidParam() Params {
	newValidParam := DefaultParams()
	r := randInt64(8, false)

	switch r {
	case 0:
		newValidParam.MintDenom = config.BaseDenom
	case 1:
		newValidParam.MintDenom = config.DisplayDenom
	case 2:
		newValidParam.InflationParameters.MaxStakingRewards = sdk.NewDec(randInt64(math.MaxInt64) + 1)
	case 3:
		newValidParam.InflationParameters.R = sdk.NewDecWithPrec(randInt64(99)+1, 2)
	case 4:
		for {
			stakingAmt := randInt64(100)
			foundationAmt := randInt64(100 - stakingAmt + 1)
			communityAmt := 100 - stakingAmt - foundationAmt
			if communityAmt >= 0 {
				newValidParam.InflationDistribution.StakingRewards = sdk.NewDecWithPrec(stakingAmt, 2)
				newValidParam.InflationDistribution.Foundation = sdk.NewDecWithPrec(foundationAmt, 2)
				newValidParam.InflationDistribution.CommunityPool = sdk.NewDecWithPrec(communityAmt, 2)
				break
			}
		}
	case 5:
		newAddr := sdk.AccAddress(randBytes(int(1 + randInt64(255))))
		newValidParam.FoundationAddress = newAddr.String()
	case 6:
		newValidParam.EnableInflation = true
	case 7:
		newValidParam.EnableInflation = false
	}

	return newValidParam
}

// newRandomInvalidParam creates a new randomized invalid parameter.
func newRandomInvalidParam() Params {
	invalidParam := DefaultParams()
	r := randInt64(11, false)

	switch r {
	case 0: // invalid denom
		for {
			denom := fmt.Sprintf("%s", randBytes(0))
			if denom != config.DisplayDenom && denom != config.BaseDenom {
				invalidParam.MintDenom = denom
				break
			}
		}
	case 1: // MaxStakingRewards <= 0
		invalidParam.InflationParameters.MaxStakingRewards = sdk.NewDec(randInt64(math.MaxInt64, true))
	case 2: // InflationParameters.R >= 1
		invalidParam.InflationParameters.R = sdk.NewDecWithPrec(100+randInt64(math.MaxInt64), 2)
	case 3: // InflationDistribution.StakingReward < 0
		invalidParam.InflationDistribution.StakingRewards = sdk.NewDecWithPrec(randInt64(math.MaxInt64, true)-1, 2)
	case 4: // InflationDistribution.StakingReward > 1
		invalidParam.InflationDistribution.StakingRewards = sdk.NewDecWithPrec(101+randInt64(math.MaxInt64), 2)
	case 5: // InflationDistribution.Foundation < 0
		invalidParam.InflationDistribution.Foundation = sdk.NewDecWithPrec(randInt64(math.MaxInt64, true)-1, 2)
	case 6: // InflationDistribution.Foundation > 1
		invalidParam.InflationDistribution.Foundation = sdk.NewDecWithPrec(101+randInt64(math.MaxInt64), 2)
	case 7: // InflationDistribution.CommunityPool < 0
		invalidParam.InflationDistribution.CommunityPool = sdk.NewDecWithPrec(randInt64(math.MaxInt64, true)-1, 2)
	case 8: // InflationDistribution.CommunityPool > 1
		invalidParam.InflationDistribution.CommunityPool = sdk.NewDecWithPrec(101+randInt64(math.MaxInt64), 2)
	case 9: // InflationDistribution.Total != 1
		for {
			stakingAmt := randInt64(math.MaxInt64, randInt64(2) == 0)
			foundationAmt := randInt64(math.MaxInt64, randInt64(2) == 0)
			communityAmt := randInt64(math.MaxInt64, randInt64(2) == 0)
			if stakingAmt+foundationAmt+communityAmt != 100 {
				invalidParam.InflationDistribution.StakingRewards = sdk.NewDecWithPrec(stakingAmt, 2)
				invalidParam.InflationDistribution.Foundation = sdk.NewDecWithPrec(foundationAmt, 2)
				invalidParam.InflationDistribution.CommunityPool = sdk.NewDecWithPrec(communityAmt, 2)
				break
			}
		}
	case 10: // Invalid FoundationAddress
		for {
			address := fmt.Sprintf("%s", randBytes(44))
			if _, err := types.GetAstraAddressFromBech32(address); err != nil {
				invalidParam.FoundationAddress = address
				break
			}
		}

	}

	return invalidParam
}

func randBytes(l int) []byte {
	if l == 0 {
		l = int(randInt64(32))
		l += 1
	}

	ret := make([]byte, l)
	_, _ = rand.Read(ret)

	return ret
}

func randInt64(max int64, negative ...bool) int64 {
	ret := new(big.Int)
	if max == 0 {
		ret, _ = rand.Int(rand.Reader, new(big.Int).SetUint64(math.MaxUint64))

	} else {
		ret, _ = rand.Int(rand.Reader, new(big.Int).SetUint64(uint64(max)))
	}

	if len(negative) > 0 && negative[0] {
		return -ret.Int64()
	}
	return ret.Int64()
}
