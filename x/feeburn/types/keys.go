package types

const (
	// ModuleName defines the module name
	ModuleName = "feeburn"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_feeburn"

	QueryParameters   = "parameters"
	QueryTotalFeeBurn = "total_fee_burn"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	prefixTotalFeeBurn = iota
)

var (
	KeyPrefixTotalFeeBurn = []byte{prefixTotalFeeBurn}
)
