package blockchain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
)

type ShardToBeaconPool interface {
	RemoveBlock(map[byte]uint64)
	//GetFinalBlock() map[byte][]ShardToBeaconBlock
	AddShardToBeaconBlock(*ShardToBeaconBlock) (uint64, uint64, error)
	//ValidateShardToBeaconBlock(ShardToBeaconBlock) error
	GetValidBlockHash() map[byte][]common.Hash
	GetValidBlock(map[byte]uint64) map[byte][]*ShardToBeaconBlock
	GetValidBlockHeight() map[byte][]uint64
	GetLatestValidPendingBlockHeight() map[byte]uint64
	GetBlockByHeight(shardID byte, height uint64) *ShardToBeaconBlock
	SetShardState(map[byte]uint64)
	GetAllBlockHeight() map[byte][]uint64
	RevertShardToBeaconPool(s byte, height uint64)
}

type CrossShardPool interface {
	AddCrossShardBlock(*CrossShardBlock) (map[byte]uint64, byte, error)
	GetValidBlock(map[byte]uint64) map[byte][]*CrossShardBlock
	GetLatestValidBlockHeight() map[byte]uint64
	GetValidBlockHeight() map[byte][]uint64
	GetBlockByHeight(_shardID byte, height uint64) *CrossShardBlock
	RemoveBlockByHeight(map[byte]uint64)
	UpdatePool() map[byte]uint64
	GetAllBlockHeight() map[byte][]uint64
	RevertCrossShardPool(uint64)
}

type ShardPool interface {
	RemoveBlock(height uint64)
	AddShardBlock(block *ShardBlock) error
	GetValidBlockHash() []common.Hash
	GetValidBlock() []*ShardBlock
	GetValidBlockHeight() []uint64
	GetLatestValidBlockHeight() uint64
	SetShardState(height uint64)
	RevertShardPool(uint64)
	GetAllBlockHeight() []uint64
	GetPendingBlockHeight() []uint64
	Start(chan struct{})
}

type BeaconPool interface {
	RemoveBlock(height uint64)
	AddBeaconBlock(block *BeaconBlock) error
	GetValidBlock() []*BeaconBlock
	GetValidBlockHeight() []uint64
	SetBeaconState(height uint64)
	GetBeaconState() uint64
	RevertBeconPool(height uint64)
	GetAllBlockHeight() []uint64
	Start(chan struct{})
	GetPendingBlockHeight() []uint64
}
type TxPool interface {
	// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
	LastUpdated() time.Time

	// MiningDescs returns a slice of mining descriptors for all the
	// transactions in the source pool.
	MiningDescs() []*metadata.TxDesc

	// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
	HaveTransaction(hash *common.Hash) bool

	// RemoveTx remove tx from tx resource
	RemoveTx(txs []metadata.Transaction, isInBlock bool)

	RemoveCandidateList([]string)

	RemoveTokenIDList([]string)

	EmptyPool() bool

	MaybeAcceptTransactionForBlockProducing(metadata.Transaction, int64) (*metadata.TxDesc, error)
	ValidateTxList(txs []metadata.Transaction) error
	//CheckTransactionFee
	// CheckTransactionFee(tx metadata.Transaction) (uint64, error)

	// Check tx validate by it self
	// ValidateTxByItSelf(tx metadata.Transaction) bool
}

type FeeEstimator interface {
	RegisterBlock(block *ShardBlock) error
}

type ChainManagerInterface interface {
	GetChainName() string
	GetShardID() int
	GetGenesisTime() int64

	// IsReady() bool                                         //TO_BE_DELETE
	// GetActiveShardNumber() int                             //TO_BE_DELETE
	// GetPubkeyRole(pubkey string, round int) (string, byte) //TO_BE_DELETE
	// GetConsensusType() string                              //TO_BE_DELETE
	// GetTimeStamp() int64                                   //TO_BE_DELETE
	// GetMinBlkInterval() time.Duration                      //TO_BE_DELETE
	// GetMaxBlkCreateTime() time.Duration                    //TO_BE_DELETE
	// GetHeight() uint64                                     //TO_BE_DELETE
	// GetCommitteeSize() int                                 //TO_BE_DELETE
	// GetCommittee() []incognitokey.CommitteePublicKey       //TO_BE_DELETE
	// GetPubKeyCommitteeIndex(string) int                    //TO_BE_DELETE
	// GetLastProposerIndex() int                             //TO_BE_DELETE

	UnmarshalBlock(blockString []byte) (common.BlockInterface, error)
	ValidateBlockSignatures(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error

	GetBestView() ChainViewInterface
	GetFinalView() ChainViewInterface
	GetAllViews() map[string]ChainViewInterface
	GetViewByHash(*common.Hash) (ChainViewInterface, error)
	GetAllTipBlocksHash() []*common.Hash
	AddView(view ChainViewInterface) error
	ConnectBlockAndAddView(block common.BlockInterface) error
}

type ChainViewInterface interface {
	GetGenesisTime() int64
	GetConsensusConfig() string
	GetConsensusType() string
	GetBlkMinInterval() time.Duration
	GetBlkMaxCreateTime() time.Duration
	GetPubkeyRole(pubkey string, round int) (string, byte)
	GetCommittee() []incognitokey.CommitteePublicKey
	GetCommitteeHash() *common.Hash
	GetCommitteeIndex(string) int
	GetTipBlock() common.BlockInterface
	GetHeight() uint64
	GetTimeStamp() int64
	GetTimeslot() uint64
	GetEpoch() uint64
	Hash() common.Hash
	GetPreviousViewHash() *common.Hash
	GetActiveShardNumber() int

	IsBestView() bool
	SetViewIsBest(isBest bool)

	DeleteView() error
	UpdateViewWithBlock(block common.BlockInterface) error
	CloneViewFrom(view *ChainViewInterface) error

	ValidateBlock(block common.BlockInterface, isPreSign bool) error
	CreateNewBlock(timeslot uint64) (common.BlockInterface, error)
	ConnectBlockAndCreateView(block common.BlockInterface) (ChainViewInterface, error)
}
