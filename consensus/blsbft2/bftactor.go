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
	Chain    common.ChainManagerInterface
	Node     consensus.NodeInterface
	ChainKey string
	PeerID   string

	UserKeySet   *MiningKey
	BFTMessageCh chan wire.MessageBFT
	isStarted    bool
	StopCh       chan struct{}
	Logger       common.Logger

	currentTimeslotOfViews map[string]uint64
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
	e.currentTimeslotOfViews = make(map[string]uint64)
	e.bestProposeBlockOfView = make(map[string]string)
	e.onGoingBlocks = make(map[string]*blockConsensusInstance)

	//init view maps
	ticker := time.Tick(1 * time.Second)
	e.Logger.Info("start bls-bftv2 consensus for chain", e.ChainKey)
	go func() {

		currentTime := time.Now().Unix()
		views := e.Chain.GetAllViews()
		for _, view := range views {
			consensusCfg, _ := parseConsensusConfig(view.GetConsensusConfig())
			consensusSlottime, _ := time.ParseDuration(consensusCfg.Slottime)
			timeSlot := getTimeSlot(view.GetGenesisTime(), currentTime, int64(consensusSlottime.Seconds()))
			e.currentTimeslotOfViews[view.Hash().String()] = timeSlot
		}

		for { //actor loop
			select {
			case <-e.StopCh:
				return
			case <-ticker:
				e.lockOnGoingBlocks.Lock()
				//check if is proposer of bestview
				bestView := e.Chain.GetBestView()
				bestViewHash := bestView.Hash().String()
				currentTime := time.Now().Unix()
				consensusCfg, _ := parseConsensusConfig(bestView.GetConsensusConfig())
				consensusSlottime, _ := time.ParseDuration(consensusCfg.Slottime)
				timeSlot := getTimeSlot(bestView.GetGenesisTime(), currentTime, int64(consensusSlottime.Seconds()))
				willProposeBlock := false
				if _, ok := e.currentTimeslotOfViews[bestViewHash]; ok {
					if e.currentTimeslotOfViews[bestViewHash]+1 == timeSlot {
						willProposeBlock = true
					}
				} else {
					willProposeBlock = true
				}
				if willProposeBlock {
					if err := e.proposeBlock(timeSlot); err != nil {
						e.Logger.Critical(consensus.UnExpectedError, errors.New("can't propose block"))
					}
				}
				//update timeslot of views
				//clean all view
				views := e.Chain.GetAllViews()
				for _, view := range views {
					_, ok := e.currentTimeslotOfViews[view.Hash().String()]
					if !ok {
						continue
					}
					consensusCfg, _ := parseConsensusConfig(view.GetConsensusConfig())
					consensusSlottime, _ := time.ParseDuration(consensusCfg.Slottime)
					timeSlot := getTimeSlot(view.GetGenesisTime(), currentTime, int64(consensusSlottime.Seconds()))
					e.currentTimeslotOfViews[view.Hash().String()] = timeSlot
				}

				for proposedBlockHash, proposedBlock := range e.onGoingBlocks {
					if proposedBlock.Phase == votePhase {
						if len(proposedBlock.Votes) > (2/3*len(proposedBlock.Committee.Committee))-1 {
							err := proposedBlock.FinalizeBlock()
							if err != nil {
								//weird thing happend
								panic(err)
							}
							go func(blockHash string) {
								e.deleteOnGoingBlock(blockHash)
							}(proposedBlockHash)
						}
					}
				}
				finalView := e.Chain.GetFinalView()
				finalViewHeight := finalView.GetHeight()
				for _, block := range e.onGoingBlocks {
					if block.Block.GetHeight() <= finalViewHeight {
						go func(blockHash string) {
							e.deleteOnGoingBlock(blockHash)
						}(block.Block.Hash().String())
					}
				}
				e.lockOnGoingBlocks.Unlock()
			}
		}
	}()
	return nil
}

func (e BLSBFT) NewInstance(chain common.ChainManagerInterface, chainKey string, node consensus.NodeInterface, logger common.Logger) consensus.ConsensusInterface {
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
