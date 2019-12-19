package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/chain"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"
)

type ShardChain struct {
	BlockGen   *BlockGenerator
	Blockchain *BlockChain
	ChainName  string
	lock       sync.RWMutex

	views     map[string]*ShardView
	bestView  *ShardView
	finalView *ShardView
}

func (chain *ShardChain) GetGenesisTime() int64 {
	return chain.bestView.GenesisTime
}

func (chain *ShardChain) IsReady() bool {
	return chain.Blockchain.Synker.IsLatest(true, chain.finalView.ShardID)
}

// func (chain *ShardChain) GetHeight() uint64 {
// 	return chain.finalView.BestBlock.Header.Height
// }

// func (chain *ShardChain) GetCommittee() []incognitokey.CommitteePublicKey {
// 	result := []incognitokey.CommitteePublicKey{}
// 	return append(result, chain.finalView.ShardCommittee...)
// }

// func (chain *ShardChain) GetCommitteeSize() int {
// 	return len(chain.finalView.ShardCommittee)
// }

// func (chain *ShardChain) GetPubKeyCommitteeIndex(pubkey string) int {
// 	for index, key := range chain.finalView.ShardCommittee {
// 		if key.GetMiningKeyBase58(chain.finalView.ConsensusAlgorithm) == pubkey {
// 			return index
// 		}
// 	}
// 	return -1
// }

// func (chain *ShardChain) GetLastProposerIndex() int {
// 	return chain.finalView.ShardProposerIdx
// }

func (chain *ShardChain) CreateNewBlock(round int) (common.BlockInterface, error) {
	chain.lock.Lock()
	defer chain.lock.Unlock()
	start := time.Now()
	Logger.log.Infof("Begin Create New Block %+v", start)
	beaconHeight := chain.Blockchain.Synker.States.ClosestState.ClosestBeaconState
	// if chain.Blockchain.FinalView.Beacon.BeaconHeight < beaconHeight {
	// 	beaconHeight = chain.Blockchain.FinalView.Beacon.BeaconHeight
	// } else {
	// 	if beaconHeight < chain.Blockchain.FinalView.Shard[byte(chain.GetShardID())].BeaconHeight {
	// 		beaconHeight = chain.Blockchain.FinalView.Shard[byte(chain.GetShardID())].BeaconHeight
	// 	}
	// }
	Logger.log.Infof("Begin Enter New Block Shard %+v", time.Now())
	newBlock, err := chain.BlockGen.NewBlockShard(byte(chain.GetShardID()), round, chain.Blockchain.Synker.GetClosestCrossShardPoolState(), beaconHeight, start)
	Logger.log.Infof("Begin Finish New Block Shard %+v", time.Now())
	if err != nil {
		return nil, err
	}
	Logger.log.Infof("Finish Create New Block %+v", start)
	return newBlock, nil
}

func (chain *ShardChain) ValidateBlockSignatures(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		return err
	}
	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committee, chain.GetConsensusType()); err != nil {
		return nil
	}
	return nil
}

func (chain *ShardChain) InsertBlk(block common.BlockInterface) error {
	if chain.Blockchain.config.ConsensusEngine.IsOngoing(chain.ChainName) {
		return NewBlockChainError(ConsensusIsOngoingError, errors.New(fmt.Sprint(chain.ChainName, block.Hash())))
	}
	return chain.Blockchain.InsertShardBlock(block.(*ShardBlock), false)
}

func (chain *ShardChain) InsertAndBroadcastBlock(block common.BlockInterface) error {
	go chain.Blockchain.config.Server.PushBlockToAll(block, false)
	err := chain.Blockchain.InsertShardBlock(block.(*ShardBlock), true)
	if err != nil {
		return err
	}
	return nil
}

func (chain *ShardChain) GetActiveShardNumber() int {
	return 0
}

func (chain *ShardChain) GetChainName() string {
	return chain.ChainName
}

func (chain *ShardChain) GetConsensusType() string {
	return chain.finalView.ConsensusAlgorithm
}

func (chain *ShardChain) GetShardID() int {
	return int(chain.finalView.ShardID)
}

func (chain *ShardChain) GetPubkeyRole(pubkey string, round int) (string, byte) {
	return chain.finalView.GetPubkeyRole(pubkey, round)
}

func (chain *ShardChain) UnmarshalBlock(blockString []byte) (common.BlockInterface, error) {
	var shardBlk ShardBlock
	err := json.Unmarshal(blockString, &shardBlk)
	if err != nil {
		return nil, err
	}
	return &shardBlk, nil
}

func (chain *ShardChain) ValidatePreSignBlock(block common.BlockInterface) error {
	return chain.Blockchain.VerifyPreSignShardBlock(block.(*ShardBlock), chain.finalView.ShardID)
}

func (chain *ShardChain) GetFinalViewConsensusType() string {
	return ""
}
func (chain *ShardChain) GetFinalViewLastBlockTimeStamp() int64 {
	return 0
}
func (chain *ShardChain) GetFinalViewMinBlkInterval() time.Duration {
	return 0
}
func (chain *ShardChain) GetFinalViewMaxBlkCreateTime() time.Duration {
	return 0
}
func (chain *ShardChain) CurrentFinalViewHeight() uint64 {
	return 0
}
func (chain *ShardChain) GetFinalViewCommitteeSize() int {
	return 0
}
func (chain *ShardChain) GetFinalViewCommittee() []incognitokey.CommitteePublicKey {
	return nil
}
func (chain *ShardChain) GetFinalViewPubKeyCommitteeIndex(string) int {
	return 0
}
func (chain *ShardChain) GetFinalViewLastProposerIndex() int {
	return 0
}

func (chain *ShardChain) GetBestView() chain.ChainViewInterface {
	return nil
}
func (chain *ShardChain) GetAllViews() map[string]chain.ChainViewInterface {
	return nil
}

func (chain *ShardChain) GetViewByHash(hash *common.Hash) (chain.ChainViewInterface, error) {
	return nil, nil
}

func (chain *ShardChain) GetFinalView() chain.ChainViewInterface {
	return nil
}

func (chain *ShardChain) storeView() error {
	return nil
}

func (chain *ShardChain) deleteView(view chain.ChainViewInterface) error {
	return nil
}

func (chain *ShardChain) loadView() error {
	return nil
}

func (chain *ShardChain) GetAllTipBlocksHash() []*common.Hash {
	result := []*common.Hash{}
	chain.lock.RLock()
	defer chain.lock.RUnlock()

	for _, view := range chain.views {
		result = append(result, view.GetTipBlock().Hash())
	}
	return result
}

func (chain *ShardChain) AddView(view chain.ChainViewInterface) error {
	chain.lock.Lock()
	defer chain.lock.Unlock()
	chain.addView(view.(*ShardView))
	return nil
}

func (chain *ShardChain) addView(view *ShardView) error {
	chain.views[view.Hash().String()] = view
	return nil
}

func (chain *ShardChain) ConnectBlockAndAddView(block common.BlockInterface) error {
	chain.lock.Lock()
	defer chain.lock.Unlock()
	view, ok := chain.views[block.GetPreviousViewHash().String()]
	if !ok {
		return NewBlockChainError(ViewNotExistError, fmt.Errorf("View %v isn't exist", block.GetPreviousViewHash().String()))
	}
	newView, err := view.ConnectBlockAndCreateView(block)
	if err != nil {
		return NewBlockChainError(ConnectBlockError, fmt.Errorf("Can't connect block %v", block.Hash().String()))
	}
	chain.views[newView.Hash().String()] = newView.(*ShardView)
	return nil
}
