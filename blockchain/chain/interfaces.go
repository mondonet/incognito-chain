package chain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

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
	CloneViewFrom(view ChainViewInterface) error

	ValidateBlock(block common.BlockInterface, isPreSign bool) error
	CreateNewBlock(timeslot uint64) (common.BlockInterface, error)
	ConnectBlockAndCreateView(block common.BlockInterface) (ChainViewInterface, error)
}
