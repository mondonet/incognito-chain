package blsbftv2

import (
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
)

type blockConsensusInstance struct {
	Engine         *BLSBFT
	View           consensus.ChainViewInterface
	ConsensusCfg   consensusConfig
	Block          common.BlockInterface
	ValidationData ValidationData
	Votes          map[string]*BFTVote
	lockVote       sync.RWMutex
	Timeslot       uint64
	Phase          string
	Committee      committeeDecode
}
