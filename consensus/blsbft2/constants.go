package blsbftv2

import (
	"errors"
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

const (
	proposePhase  = "PROPOSE"
	listenPhase   = "LISTEN"
	votePhase     = "VOTE"
	commitPhase   = "COMMIT"
	newphase      = "NEWPHASE"
	consensusName = common.BlsConsensus2
)

//
const (
	committeeCacheCleanupTime = 40 * time.Minute
)

var (
	UnexpectedError                               = errors.New("unexpected error")
	BlockAlreadyFinalizedError                    = errors.New("Block Already Finalized Error")
	AlreadyReceivedProposedBlockError             = errors.New("Already Received Proposed Block Error")
	BlockIsMissingFromBlockConsensusInstanceError = errors.New("Block Is Missing From BlockConsensusInstance Error")
)
