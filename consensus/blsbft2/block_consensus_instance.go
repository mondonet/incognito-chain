package blsbftv2

import (
	"github.com/incognitochain/incognito-chain/consensus"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
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
