package keeper

import (
	"github.com/AstraProtocol/astra/v1/x/inflation/types"
)

var _ types.QueryServer = Keeper{}
