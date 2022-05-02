package tests

import (
	"fmt"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"github.com/tharsis/evmos/v3/x/erc20/types"
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

func (suite KeeperTestSuite) TestToggleRelay() {
	var (
		contractAddr common.Address
		id           []byte
		pair         types.TokenPair
	)

	testCases := []struct {
		name         string
		malleate     func()
		expPass      bool
		relayEnabled bool
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
			"disable relay",
			func() {
				contractAddr = suite.setupRegisterERC20Pair(contractMinterBurner)
				id = suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, _ = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
			},
			true,
			false,
		},
		{
			"disable and enable relay",
			func() {
				contractAddr = suite.setupRegisterERC20Pair(contractMinterBurner)
				id = suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, _ = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				pair, _ = suite.app.Erc20Keeper.ToggleRelay(suite.ctx, contractAddr.String())
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
			pair, err = suite.app.Erc20Keeper.ToggleRelay(suite.ctx, contractAddr.String())
			// Request the pair using the GetPairToken func to make sure that is updated on the db
			pair, _ = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				if tc.relayEnabled {
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

func (suite KeeperTestSuite) TestUpdateTokenPairERC20() {
	var (
		contractAddr    common.Address
		contractAddr2   common.Address
		pair            types.TokenPair
		metadata        banktypes.Metadata
		newContractAddr common.Address
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
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
		},
		{
			"token not registered - Metadata not found",
			func() {
				contractAddr, err := suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				suite.Require().NoError(err)
				suite.Commit()
				pair = types.NewTokenPair(contractAddr, cosmosTokenBase, true, types.OWNER_MODULE)

				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())
			},
			false,
		},
		{
			"newErc20 not found",
			func() {
				contractAddr = suite.setupRegisterERC20Pair(contractMinterBurner)
				newContractAddr = common.Address{}
			},
			false,
		},
		{
			"empty denom units",
			func() {
				var found bool
				contractAddr = suite.setupRegisterERC20Pair(contractMinterBurner)
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, banktypes.Metadata{Base: pair.Denom})
				suite.Commit()

				// Deploy a new contract with the same values
				var err error
				newContractAddr, err = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"metadata ERC20 details mismatch",
			func() {
				var found bool
				contractAddr = suite.setupRegisterERC20Pair(contractMinterBurner)
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata := banktypes.Metadata{Base: pair.Denom, DenomUnits: []*banktypes.DenomUnit{{}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata)
				suite.Commit()

				// Deploy a new contract with the same values
				var err error
				newContractAddr, err = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"no denom unit with ERC20 name",
			func() {
				var found bool
				contractAddr = suite.setupRegisterERC20Pair(contractMinterBurner)
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata := banktypes.Metadata{Base: pair.Denom, Display: erc20Name, Description: types.CreateDenomDescription(contractAddr.String()), Symbol: erc20Symbol, DenomUnits: []*banktypes.DenomUnit{{}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata)
				suite.Commit()

				// Deploy a new contract with the same values
				var err error
				newContractAddr, err = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"denom unit and ERC20 decimals mismatch",
			func() {
				var found bool
				contractAddr = suite.setupRegisterERC20Pair(contractMinterBurner)
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata := banktypes.Metadata{Base: pair.Denom, Display: erc20Name, Description: types.CreateDenomDescription(contractAddr.String()), Symbol: erc20Symbol, DenomUnits: []*banktypes.DenomUnit{{Denom: erc20Name}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata)
				suite.Commit()

				// Deploy a new contract with the same values
				var err error
				newContractAddr, err = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"ok",
			func() {
				var found bool
				contractAddr = suite.setupRegisterERC20Pair(contractMinterBurner)
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata := banktypes.Metadata{Base: pair.Denom, Display: erc20Name, Description: types.CreateDenomDescription(contractAddr.String()), Symbol: erc20Symbol, DenomUnits: []*banktypes.DenomUnit{{Denom: erc20Name, Exponent: 18}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata)
				suite.Commit()

				// Deploy a new contract with the same values
				var err error
				newContractAddr, err = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"metadata details (display, symbol, description) don't match",
			func() {
				var found bool
				contractAddr, contractAddr2 = suite.setupRegisterERC20TwoPair(contractMinterBurner)
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata := banktypes.Metadata{Base: pair.Denom, Display: erc20Name, Description: types.CreateDenomDescription(contractAddr.String()), Symbol: erc20Symbol, DenomUnits: []*banktypes.DenomUnit{{Denom: erc20Name, Exponent: 18}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata)
				suite.Commit()

				id = suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr2.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata2 := banktypes.Metadata{Base: pair.Denom, Display: erc20Name2, Description: types.CreateDenomDescription(contractAddr2.String()), Symbol: erc20Symbol2, DenomUnits: []*banktypes.DenomUnit{{Denom: erc20Name2, Exponent: 18}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata2)
				suite.Commit()

				// Deploy a new contract with the same values
				var err error
				newContractAddr, err = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				suite.Require().NoError(err)

				pair, err = suite.app.Erc20Keeper.UpdateTokenPairERC20(suite.ctx, contractAddr, newContractAddr)
				suite.Require().NoError(err)
				_, err2 := suite.app.Erc20Keeper.UpdateTokenPairERC20(suite.ctx, contractAddr2, newContractAddr)
				suite.Require().Error(err2)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			var err error
			pair, err = suite.app.Erc20Keeper.UpdateTokenPairERC20(suite.ctx, contractAddr, newContractAddr)
			metadata, _ = suite.app.BankKeeper.GetDenomMetaData(suite.ctx, types.CreateDenom(contractAddr.String()))

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(newContractAddr.String(), pair.Erc20Address)
				suite.Require().Equal(types.CreateDenomDescription(newContractAddr.String()), metadata.Description)
			} else {
				suite.Require().Error(err, tc.name)
				if suite.app.Erc20Keeper.IsTokenPairRegistered(suite.ctx, pair.GetID()) {
					suite.Require().Equal(contractAddr.String(), pair.Erc20Address, "check pair")
					suite.Require().Equal(types.CreateDenomDescription(contractAddr.String()), metadata.Description, "check metadata")
				}
			}
		})
	}
}

func (suite KeeperTestSuite) TestUpdateTokenPairERC20_NewContractAddr() {
	var (
		contractAddr    common.Address
		contractAddr2   common.Address
		pair            types.TokenPair
		newContractAddr common.Address
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"newContractAddr already registered",
			func() {
				var found bool
				contractAddr, contractAddr2 = suite.setupRegisterERC20TwoPairSameErc20NameSymbol(contractMinterBurner)
				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata := banktypes.Metadata{Base: pair.Denom, Display: erc20Name, Description: types.CreateDenomDescription(contractAddr.String()), Symbol: erc20Symbol, DenomUnits: []*banktypes.DenomUnit{{Denom: erc20Name, Exponent: 18}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata)
				suite.Commit()

				id = suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr2.String())
				pair, found = suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(found)
				metadata2 := banktypes.Metadata{Base: pair.Denom, Display: erc20Name, Description: types.CreateDenomDescription(contractAddr2.String()), Symbol: erc20Symbol, DenomUnits: []*banktypes.DenomUnit{{Denom: erc20Name, Exponent: 18}}}
				suite.app.BankKeeper.SetDenomMetaData(suite.ctx, metadata2)
				suite.Commit()

				// Deploy a new contract with the same values
				var err error
				newContractAddr, err = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				suite.Require().NoError(err)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			var err error
			pair, err = suite.app.Erc20Keeper.UpdateTokenPairERC20(suite.ctx, contractAddr, newContractAddr)
			suite.Require().NoError(err)
			pair2, err2 := suite.app.Erc20Keeper.UpdateTokenPairERC20(suite.ctx, contractAddr2, newContractAddr)
			suite.Require().NoError(err2)
			metadata, _ := suite.app.BankKeeper.GetDenomMetaData(suite.ctx, types.CreateDenom(contractAddr.String()))
			metadata2, _ := suite.app.BankKeeper.GetDenomMetaData(suite.ctx, types.CreateDenom(contractAddr2.String()))

			suite.Require().Equal(pair.GetErc20Address(), pair2.GetErc20Address())
			suite.Require().NotEqual(pair.GetDenom(), pair2.GetDenom())

			suite.Require().Equal(metadata.GetDescription(), metadata2.GetDescription())
			suite.Require().NotEqual(metadata.GetBase(), metadata2.GetBase())

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(newContractAddr.String(), pair.Erc20Address)
				suite.Require().Equal(types.CreateDenomDescription(newContractAddr.String()), metadata.Description)
			} else {
				suite.Require().Error(err, tc.name)
				if suite.app.Erc20Keeper.IsTokenPairRegistered(suite.ctx, pair.GetID()) {
					suite.Require().Equal(contractAddr.String(), pair.Erc20Address, "check pair")
					suite.Require().Equal(types.CreateDenomDescription(contractAddr.String()), metadata.Description, "check metadata")
				}
			}
		})
	}
}
