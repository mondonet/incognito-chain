package common

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"time"
)

type BlockInterface interface {
	GetHeight() uint64
	Hash() *Hash
	GetProducer() string
	GetValidationField() string
	GetRound() int
	GetRoundKey() string
	GetInstructions() [][]string
	GetConsensusType() string
	GetEpoch() uint64
	GetPreviousViewHash() *Hash
	GetPreviousBlockHash() *Hash
	GetTimeslot() uint64
	GetBlockTimestamp() int64
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

	UnmarshalBlock(blockString []byte) (BlockInterface, error)
	ValidateBlockSignatures(block BlockInterface, committee []incognitokey.CommitteePublicKey) error

	GetBestView() ChainViewInterface
	GetFinalView() ChainViewInterface
	GetAllViews() map[string]ChainViewInterface
	GetViewByHash(*common.Hash) (ChainViewInterface, error)
	GetAllTipBlocksHash() []*common.Hash
	AddView(view ChainViewInterface) error
	ConnectBlockAndAddView(block BlockInterface) error
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
	GetTipBlock() BlockInterface
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
	UpdateViewWithBlock(block BlockInterface) error
	CloneViewFrom(view *ChainViewInterface) error

	ValidateBlock(block BlockInterface, isPreSign bool) error
	CreateNewBlock(timeslot uint64) (BlockInterface, error)
	ConnectBlockAndCreateView(block BlockInterface) (ChainViewInterface, error)
}
