package tests

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"github.com/tharsis/evmos/v4/x/erc20/types"
	"testing"
)

func TestErc20ProposalsTestingSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite KeeperTestSuite) TestRegisterERC20() {
	var (
		contractAddr common.Address
		pair         types.TokenPair
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"intrarelaying is disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"token ERC20 already registered",
			func() {
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), pair.GetID())
			},
			false,
		},
		{
			"denom already registered",
			func() {
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())
			},
			false,
		},
		{
			"meta data already stored",
			func() {
				suite.app.Erc20Keeper.CreateCoinMetadata(suite.ctx, contractAddr)
			},
			false,
		},
		{
			"ok",
			func() {},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			var err error
			contractAddr, err = suite.DeployContract(erc20Name, erc20Symbol, cosmosDecimals)
			suite.Require().NoError(err)
			suite.Commit()
			coinName := types.CreateDenom(contractAddr.String())
			pair = types.NewTokenPair(contractAddr, coinName, true, types.OWNER_EXTERNAL)

			tc.malleate()

			_, err = suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
			metadata, found := suite.app.BankKeeper.GetDenomMetaData(suite.ctx, coinName)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				// Metadata variables
				suite.Require().True(found)
				suite.Require().Equal(coinName, metadata.Base)
				suite.Require().Equal(coinName, metadata.Name)
				suite.Require().Equal(types.SanitizeERC20Name(erc20Name), metadata.Display)
				suite.Require().Equal(erc20Symbol, metadata.Symbol)
				// Denom units
				suite.Require().Equal(len(metadata.DenomUnits), 2)
				suite.Require().Equal(coinName, metadata.DenomUnits[0].Denom)
				suite.Require().Equal(uint32(zeroExponent), metadata.DenomUnits[0].Exponent)
				suite.Require().Equal(types.SanitizeERC20Name(erc20Name), metadata.DenomUnits[1].Denom)
				// Custom exponent at contract creation matches coin with token
				suite.Require().Equal(metadata.DenomUnits[1].Exponent, uint32(cosmosDecimals))
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite KeeperTestSuite) TestToggleConversion() {
	var (
		contractAddr common.Address
		id           []byte
		pair         types.TokenPair
	)

	testCases := []struct {
		name              string
		malleate          func()
		expPass           bool
		conversionEnabled bool
	}{
		{
			"token not registered",
			func() {
				contractAddr, err := suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				suite.Require().NoError(err)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenBase, true, types.OWNER_MODULE)
			},
			false,
			false,
		},
		{
			"token not registered - pair not found",
			func() {
				contractAddr, err := suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				suite.Require().NoError(err)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenBase, true, types.OWNER_MODULE)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())
			},
			false,
			false,
		},
		{
			"disable conversion",
			func() {
				contractAddr = suite.setupRegisterERC20Pair(contractMinterBurner)
				id = suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, _ = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
			},
			true,
			false,
		},
		{
			"disable and enable conversion",
			func() {
				contractAddr = suite.setupRegisterERC20Pair(contractMinterBurner)
				id = suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, _ = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				pair, _ = suite.app.Erc20Keeper.ToggleConversion(suite.ctx, contractAddr.String())
			},
			true,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			var err error
			pair, err = suite.app.Erc20Keeper.ToggleConversion(suite.ctx, contractAddr.String())
			// Request the pair using the GetPairToken func to make sure that is updated on the db
			pair, _ = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				if tc.conversionEnabled {
					suite.Require().True(pair.Enabled)
				} else {
					suite.Require().False(pair.Enabled)
				}
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}
