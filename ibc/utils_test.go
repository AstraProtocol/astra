package ibc

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("astra", "astrapub")
}

func TestGetTransferSenderRecipient(t *testing.T) {
	testCases := []struct {
		name         string
		packet       channeltypes.Packet
		expSender    string
		expRecipient string
		expError     bool
	}{
		{
			"empty packet",
			channeltypes.Packet{},
			"", "",
			true,
		},
		{
			"invalid packet data",
			channeltypes.Packet{
				Data: ibctesting.MockFailPacketData,
			},
			"", "",
			true,
		},
		{
			"empty FungibleTokenPacketData",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{},
				),
			},
			"", "",
			true,
		},
		{
			"invalid sender",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1",
						Receiver: "astra1x2w87cvt5mqjncav4lxy8yfreynn273xjldukc",
					},
				),
			},
			"", "",
			true,
		},
		{
			"invalid recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "astra1",
					},
				),
			},
			"", "",
			true,
		},
		{
			"valid - cosmos sender, astra recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "astra1x2w87cvt5mqjncav4lxy8yfreynn273xjldukc",
					},
				),
			},
			"astra1qql8ag4cluz6r4dz28p3w00dnc9w8ueuuz8try",
			"astra1x2w87cvt5mqjncav4lxy8yfreynn273xjldukc",
			false,
		},
		{
			"valid - astra sender, cosmos recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "astra1x2w87cvt5mqjncav4lxy8yfreynn273xjldukc",
						Receiver: "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
					},
				),
			},
			"astra1x2w87cvt5mqjncav4lxy8yfreynn273xjldukc",
			"astra1qql8ag4cluz6r4dz28p3w00dnc9w8ueuuz8try",
			false,
		},
		{
			"valid - osmosis sender, astra recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2",
						Receiver: "astra1x2w87cvt5mqjncav4lxy8yfreynn273xjldukc",
					},
				),
			},
			"astra1qql8ag4cluz6r4dz28p3w00dnc9w8ueuuz8try",
			"astra1x2w87cvt5mqjncav4lxy8yfreynn273xjldukc",
			false,
		},
	}

	for _, tc := range testCases {
		sender, recipient, _, _, err := GetTransferSenderRecipient(tc.packet)
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expSender, sender.String())
			require.Equal(t, tc.expRecipient, recipient.String())
		}
	}
}
