package keeper_test

import (
	"github.com/AstraProtocol/astra/v2/app"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	"github.com/AstraProtocol/astra/v2/x/mint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evm "github.com/evmos/ethermint/x/evm/types"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
	"testing"
	"time"
)

var denomMint = types.DefaultInflationDenom

type KeeperTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	app            *app.Astra
	queryClientEvm evm.QueryClient
	queryClient    types.QueryClient
	consAddress    sdk.ConsAddress
}

var s *KeeperTestSuite

const (
	numTests = 1000
)

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	gomega.RegisterFailHandler(ginkgo.Fail)

	suite.Run(t, s)
	ginkgo.RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) SetupTest() {
	config.SetBech32Prefixes(sdk.GetConfig())
	suite.DoSetupTest(suite.T())
}

// Set-up test
func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	checkTx := false

	// init app
	suite.app = app.Setup(checkTx, nil)

	// setup context
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         app.TestnetChainID + "-1",
		Time:            time.Now().UTC(),
		ProposerAddress: suite.consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	// setup query helpers
	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.MintKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

// CommitAndBeginBlock commits the current state and begins the next block (without committing).
func (suite *KeeperTestSuite) CommitAndBeginBlock() {
	suite.CommitAndBeginBlocks(1)
}

// CommitAndBeginBlocks commits the current state and begins subsequent blocks.
func (suite *KeeperTestSuite) CommitAndBeginBlocks(numBlocks int64) {
	for i := int64(0); i < numBlocks; i++ {
		_ = suite.app.Commit()
		header := suite.ctx.BlockHeader()
		header.Height += 1
		header.Time = header.Time.Add(3 * time.Second)
		suite.app.BeginBlock(abci.RequestBeginBlock{
			Header: header,
		})

		// update ctx
		suite.ctx = suite.app.BaseApp.NewContext(false, header)

		queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
		evm.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
		suite.queryClientEvm = evm.NewQueryClient(queryHelper)
	}
}
