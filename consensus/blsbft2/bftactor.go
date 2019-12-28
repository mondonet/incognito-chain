package blsbftv2

import (
	"errors"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/patrickmn/go-cache"
)

type BLSBFT struct {
	Chain    consensus.ChainManagerInterface
	Node     consensus.NodeInterface
	ChainKey string
	PeerID   string

	UserKeySet   *MiningKey
	BFTMessageCh chan wire.MessageBFT
	isStarted    bool
	StopCh       chan struct{}
	Logger       common.Logger

	bestProposeBlockOfView map[string]string
	onGoingBlocks          map[string]*blockConsensusInstance
	lockOnGoingBlocks      sync.RWMutex

	proposedBlockOnView struct {
		ViewHash  string
		BlockHash string
	}

	viewCommitteesCache *cache.Cache // [committeeHash]:committeeDecodeStruct
}

type consensusConfig struct {
	Slottime string
}
type committeeDecode struct {
	Committee  []incognitokey.CommitteePublicKey
	StringList []string
	ByteList   []blsmultisig.PublicKey
}

func (e *BLSBFT) GetConsensusName() string {
	return consensusName
}

func (e *BLSBFT) Stop() error {
	if e.isStarted {
		select {
		case <-e.StopCh:
			return nil
		default:
			close(e.StopCh)
		}
		e.isStarted = false
	}
	return consensus.NewConsensusError(consensus.ConsensusAlreadyStoppedError, errors.New(e.ChainKey))
}

func (e *BLSBFT) Start() error {
	if e.isStarted {
		return consensus.NewConsensusError(consensus.ConsensusAlreadyStartedError, errors.New(e.ChainKey))
	}
	e.isStarted = true
	e.StopCh = make(chan struct{})
	e.bestProposeBlockOfView = make(map[string]string)
	e.onGoingBlocks = make(map[string]*blockConsensusInstance)

	//init view maps
	e.Logger.Info("start bls-bftv2 consensus for chain", e.ChainKey)
	go func() {
		e.viewWatcher(e.Chain.GetBestView())
	}()
	return nil
}

func (e BLSBFT) NewInstance(chain consensus.ChainManagerInterface, chainKey string, node consensus.NodeInterface, logger common.Logger) consensus.ConsensusInterface {
	var newInstance BLSBFT
	newInstance.Chain = chain
	newInstance.ChainKey = chainKey
	newInstance.Node = node
	newInstance.UserKeySet = e.UserKeySet
	newInstance.Logger = logger
	return &newInstance
}

func init() {
	consensus.RegisterConsensus(common.BlsConsensus2, &BLSBFT{})
}

func (e *BLSBFT) deleteOnGoingBlock(blockHash string) {
	e.lockOnGoingBlocks.Lock()
	if _, ok := e.bestProposeBlockOfView[blockHash]; ok {
		delete(e.bestProposeBlockOfView, blockHash)
	}
	delete(e.onGoingBlocks, blockHash)
	e.lockOnGoingBlocks.Unlock()
}

func (e *BLSBFT) chainWatcher() {
	var updateCh chan consensus.ChainUpdateInfo
	updateCh = make(chan consensus.ChainUpdateInfo)
	e.Chain.SubscribeChainUpdate("BLSBFT", updateCh)

	for {
		update := <-updateCh
		switch update.Action {
		case common.VIEWADDED:
			go func() {
				e.viewWatcher(update.View)
			}()
		case common.CHAINFINALIZED:
			e.lockOnGoingBlocks.Lock()
			defer e.lockOnGoingBlocks.Unlock()
			finalizedHeight := e.Chain.GetBestView().GetHeight()
			for blockHash, bcsi := range e.onGoingBlocks {
				if bcsi.Height <= finalizedHeight {
					delete(e.onGoingBlocks, blockHash)
				}
			}
		default:
			continue
		}
	}
}

//viewWatcher - check whether view is best and in timeslot to propose block
func (e *BLSBFT) viewWatcher(view consensus.ChainViewInterface) {
	viewHash := view.Hash().String()
	var timeoutCh chan struct{}
	timeoutCh = make(chan struct{})
	for {
		isBestView := (e.Chain.GetBestView().Hash().String() == viewHash)
		currentTime := time.Now().Unix()
		consensusCfg, _ := parseConsensusConfig(view.GetConsensusConfig())
		consensusSlottime, _ := time.ParseDuration(consensusCfg.Slottime)
		timeSlot, nextTimeSlotIn := getTimeSlot(view.GetGenesisTime(), currentTime, int64(consensusSlottime.Seconds()))
		if isBestView {
			if err := e.proposeBlock(timeSlot); err != nil {
				e.Logger.Critical(consensus.UnExpectedError, errors.New("can't propose block"))
			}
		} else {
			//clean up other related stuff
			if e.proposedBlockOnView.ViewHash == viewHash {
				e.proposedBlockOnView.ViewHash = ""
				e.proposedBlockOnView.BlockHash = ""
			}
			return
		}
		timer := time.AfterFunc(nextTimeSlotIn, func() {
			timeoutCh <- struct{}{}
		})
		<-timeoutCh
		timer.Stop()
	}
}
