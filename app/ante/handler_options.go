package ante

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	ibcante "github.com/cosmos/ibc-go/v6/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"
	cosmosante "github.com/evmos/evmos/v12/app/ante/cosmos"

	evmante "github.com/evmos/evmos/v12/app/ante/evm"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"

	feeburnante "github.com/AstraProtocol/astra/v3/x/feeburn/ante"
	feeburntypes "github.com/AstraProtocol/astra/v3/x/feeburn/types"
	anteutils "github.com/evmos/evmos/v12/app/ante/utils"
	vestingtypes "github.com/evmos/evmos/v12/x/vesting/types"
)

// HandlerOptions defines the list of module keepers required to run the Evmos
// AnteHandler decorators.
type HandlerOptions struct {
	AccountKeeper      evmtypes.AccountKeeper
	BankKeeper         evmtypes.BankKeeper
	BankKeeperFork     feeburntypes.BankKeeper
	IBCKeeper          *ibckeeper.Keeper
	FeeMarketKeeper    evmante.FeeMarketKeeper
	StakingKeeper      vestingtypes.StakingKeeper
	EvmKeeper          evmante.EVMKeeper
	DistributionKeeper anteutils.DistributionKeeper
	FeegrantKeeper     ante.FeegrantKeeper
	FeeBurnKeeper      feeburntypes.FeeBurnKeeper
	SignModeHandler    authsigning.SignModeHandler
	SigGasConsumer     func(meter sdk.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	Cdc                codec.BinaryCodec
	MaxTxGasWanted     uint64
}

// Validate checks if the keepers are defined
func (options HandlerOptions) Validate() error {
	if options.AccountKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.DistributionKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "distribution keeper is required for AnteHandler")
	}
	if options.FeeBurnKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "feeburn keeper is required for AnteHandler")
	}
	if options.StakingKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "staking keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.FeeMarketKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "fee market keeper is required for AnteHandler")
	}
	if options.EvmKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "evm keeper is required for AnteHandler")
	}
	return nil
}

// newCosmosAnteHandler creates the default ante handler for Ethereum transactions
func newEthAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		// outermost AnteDecorator. SetUpContext must be called first
		evmante.NewEthSetUpContextDecorator(options.EvmKeeper),
		// Check eth effective gas price against minimal-gas-prices
		evmante.NewEthMempoolFeeDecorator(options.EvmKeeper),
		// Check eth effective gas price against the global MinGasPrice
		evmante.NewEthMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper),
		evmante.NewEthValidateBasicDecorator(options.EvmKeeper),
		evmante.NewEthSigVerificationDecorator(options.EvmKeeper),
		evmante.NewEthAccountVerificationDecorator(options.AccountKeeper, options.EvmKeeper),
		evmante.NewCanTransferDecorator(options.EvmKeeper),
		NewEthVestingTransactionDecorator(options.AccountKeeper), // TODO(thanhnv): sync with evmos
		evmante.NewEthGasConsumeDecorator(options.BankKeeper, options.DistributionKeeper, options.EvmKeeper, options.StakingKeeper, options.MaxTxGasWanted),
		evmante.NewEthIncrementSenderSequenceDecorator(options.AccountKeeper),
		evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
		// emit eth tx hash and index at the very last ante handler.
		evmante.NewEthEmitEventDecorator(options.EvmKeeper),
	)
}

// newCosmosAnteHandler creates the default ante handler for Cosmos transactions
func newCosmosAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		// reject MsgEthereumTxs
		cosmosante.RejectMessagesDecorator{},
		// disable the Msg types that cannot be included on an authz.MsgExec msgs field
		cosmosante.NewAuthzLimiterDecorator(
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}),
		),
		ante.NewSetUpContextDecorator(),
		ante.NewExtensionOptionsDecorator(nil), // see https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md#api-breaking-changes-5
		cosmosante.NewMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper,
			options.FeegrantKeeper, nil), // see https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md#api-breaking-changes-5
		feeburnante.NewFeeBurnDecorator(options.BankKeeperFork, options.FeeBurnKeeper),
		NewVestingDelegationDecorator(options.AccountKeeper, options.StakingKeeper, options.Cdc), // TODO(thanhnv): sync with evmos
		NewValidatorCommissionDecorator(options.Cdc),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	)
}

// newCosmosAnteHandlerEip712 creates the ante handler for transactions signed with EIP712
func newCosmosAnteHandlerEip712(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		// reject MsgEthereumTxs
		cosmosante.RejectMessagesDecorator{},
		// disable the Msg types that cannot be included on an authz.MsgExec msgs field
		cosmosante.NewAuthzLimiterDecorator(
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}),
		),
		ante.NewSetUpContextDecorator(),
		ante.NewValidateBasicDecorator(),
		cosmosante.NewMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper,
			options.FeegrantKeeper, nil),
		feeburnante.NewFeeBurnDecorator(options.BankKeeperFork, options.FeeBurnKeeper),
		NewVestingDelegationDecorator(options.AccountKeeper, options.StakingKeeper, options.Cdc), //TODO(vesting delegation)
		NewValidatorCommissionDecorator(options.Cdc),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		// Note: signature verification uses EIP instead of the cosmos signature validator
		cosmosante.NewLegacyEip712SigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	)
}
