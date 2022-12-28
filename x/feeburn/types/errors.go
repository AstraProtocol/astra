package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ErrFeeBurnSend x/feeburn module sentinel errors
var (
	ErrFeeBurnSend = sdkerrors.Register(ModuleName, 1, "feeburn send error")
)
