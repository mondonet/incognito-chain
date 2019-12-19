package blockchain

import (
	"encoding/json"
	"fmt"

	// "github.com/incognitochain/incognito-chain/blockchain/chain"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"
)

type BeaconChain struct {
	BlockGen   *BlockGenerator
	Blockchain *BlockChain
	ChainName  string
	lock       sync.RWMutex

	views     map[string]*BeaconView
	bestView  *BeaconView
	finalView *BeaconView
}

func (chain *BeaconChain) GetGenesisTime() int64 {
	return chain.bestView.GenesisTime
}

func (chain *BeaconChain) GetTimeStamp() int64 {
	return chain.finalView.BestBlock.Header.Timestamp
}

func (chain *BeaconChain) GetMinBlkInterval() time.Duration {
	return chain.finalView.BlockInterval
}

func (chain *BeaconChain) GetMaxBlkCreateTime() time.Duration {
	return chain.finalView.BlockMaxCreateTime
}

func (chain *BeaconChain) IsReady() bool {
	return chain.Blockchain.Synker.IsLatest(false, 0)
}

func (chain *BeaconChain) GetHeight() uint64 {
	return chain.finalView.BestBlock.Header.Height
}

func (chain *BeaconChain) GetCommittee() []incognitokey.CommitteePublicKey {
	return chain.finalView.GetBeaconCommittee()
}

func (chain *BeaconChain) GetCommitteeSize() int {
	return len(chain.finalView.BeaconCommittee)
}

func (chain *BeaconChain) GetPubKeyCommitteeIndex(pubkey string) int {
	for index, key := range chain.finalView.GetBeaconCommittee() {
		if key.GetMiningKeyBase58(chain.finalView.ConsensusAlgorithm) == pubkey {
			return index
		}
	}
	return -1
}

func (chain *BeaconChain) GetLastProposerIndex() int {
	return chain.finalView.BeaconProposerIndex
}

func (chain *BeaconChain) CreateNewBlock(round int) (common.BlockInterface, error) {
	// chain.lock.Lock()
	// defer chain.lock.Unlock()
	newBlock, err := chain.BlockGen.NewBlockBeacon(round, chain.Blockchain.Synker.GetClosestShardToBeaconPoolState())
	if err != nil {
		return nil, err
	}
	return newBlock, nil
}

func (chain *BeaconChain) InsertBlk(block common.BlockInterface) error {
	if chain.Blockchain.config.ConsensusEngine.IsOngoing(common.BeaconChainKey) {
		return NewBlockChainError(ConsensusIsOngoingError, errors.New(fmt.Sprint(common.BeaconChainKey, block.Hash())))
	}
	return chain.Blockchain.InsertBeaconBlock(block.(*BeaconBlock), true)
}

func (chain *BeaconChain) InsertAndBroadcastBlock(block common.BlockInterface) error {
	go chain.Blockchain.config.Server.PushBlockToAll(block, true)
	err := chain.Blockchain.InsertBeaconBlock(block.(*BeaconBlock), true)
	if err != nil {
		return err
	}
	return nil
}

func (chain *BeaconChain) GetActiveShardNumber() int {
	return chain.finalView.ActiveShards
}

func (chain *BeaconChain) GetChainName() string {
	return chain.ChainName
}

func (chain *BeaconChain) GetPubkeyRole(pubkey string, round int) (string, byte) {
	return chain.finalView.GetPubkeyRole(pubkey, round)
}

func (chain *BeaconChain) ValidatePreSignBlock(block common.BlockInterface) error {
	return chain.Blockchain.VerifyPreSignBeaconBlock(block.(*BeaconBlock), true)
}

func (chain *BeaconChain) ValidateBlockSignatures(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		return err
	}
	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committee, chain.GetConsensusType()); err != nil {
		return nil
	}
	return nil
}

func (chain *BeaconChain) GetConsensusType() string {
	return chain.finalView.ConsensusAlgorithm
}

func (chain *BeaconChain) GetShardID() int {
	return -1
}

func (chain *BeaconChain) GetAllCommittees() map[string]map[string][]incognitokey.CommitteePublicKey {
	var result map[string]map[string][]incognitokey.CommitteePublicKey
	result = make(map[string]map[string][]incognitokey.CommitteePublicKey)
	result[chain.finalView.ConsensusAlgorithm] = make(map[string][]incognitokey.CommitteePublicKey)
	result[chain.finalView.ConsensusAlgorithm][common.BeaconChainKey] = append([]incognitokey.CommitteePublicKey{}, chain.finalView.BeaconCommittee...)
	for shardID, consensusType := range chain.finalView.GetShardConsensusAlgorithm() {
		if _, ok := result[consensusType]; !ok {
			result[consensusType] = make(map[string][]incognitokey.CommitteePublicKey)
		}
		result[consensusType][common.GetShardChainKey(shardID)] = append([]incognitokey.CommitteePublicKey{}, chain.finalView.ShardCommittee[shardID]...)
	}
	return result
}

func (chain *BeaconChain) GetBeaconPendingList() []incognitokey.CommitteePublicKey {
	var result []incognitokey.CommitteePublicKey
	result = append(result, chain.finalView.BeaconPendingValidator...)
	return result
}

func (chain *BeaconChain) GetShardsPendingList() map[string]map[string][]incognitokey.CommitteePublicKey {
	var result map[string]map[string][]incognitokey.CommitteePublicKey
	result = make(map[string]map[string][]incognitokey.CommitteePublicKey)
	for shardID, consensusType := range chain.finalView.GetShardConsensusAlgorithm() {
		if _, ok := result[consensusType]; !ok {
			result[consensusType] = make(map[string][]incognitokey.CommitteePublicKey)
		}
		result[consensusType][common.GetShardChainKey(shardID)] = append([]incognitokey.CommitteePublicKey{}, chain.finalView.ShardPendingValidator[shardID]...)
	}
	return result
}

func (chain *BeaconChain) GetShardsWaitingList() []incognitokey.CommitteePublicKey {
	var result []incognitokey.CommitteePublicKey
	result = append(result, chain.finalView.CandidateShardWaitingForNextRandom...)
	result = append(result, chain.finalView.CandidateShardWaitingForCurrentRandom...)
	return result
}

func (chain *BeaconChain) GetBeaconWaitingList() []incognitokey.CommitteePublicKey {
	var result []incognitokey.CommitteePublicKey
	result = append(result, chain.finalView.CandidateBeaconWaitingForNextRandom...)
	result = append(result, chain.finalView.CandidateBeaconWaitingForCurrentRandom...)
	return result
}

func (chain *BeaconChain) UnmarshalBlock(blockString []byte) (common.BlockInterface, error) {
	var beaconBlk BeaconBlock
	err := json.Unmarshal(blockString, &beaconBlk)
	if err != nil {
		return nil, err
	}
	return &beaconBlk, nil
}

func (chain *BeaconChain) GetFinalViewConsensusType() string {
	return ""
}
func (chain *BeaconChain) GetFinalViewLastBlockTimeStamp() int64 {
	return 0
}
func (chain *BeaconChain) GetFinalViewMinBlkInterval() time.Duration {
	return 0
}
func (chain *BeaconChain) GetFinalViewMaxBlkCreateTime() time.Duration {
	return 0
}
func (chain *BeaconChain) CurrentFinalViewHeight() uint64 {
	return 0
}
func (chain *BeaconChain) GetFinalViewCommitteeSize() int {
	return 0
}
func (chain *BeaconChain) GetFinalViewCommittee() []incognitokey.CommitteePublicKey {
	return nil
}
func (chain *BeaconChain) GetFinalViewPubKeyCommitteeIndex(string) int {
	return 0
}
func (chain *BeaconChain) GetFinalViewLastProposerIndex() int {
	return 0
}

func (chain *BeaconChain) GetBestView() consensus.ChainViewInterface {
	return nil
}
func (chain *BeaconChain) GetAllViews() map[string]consensus.ChainViewInterface {
	return nil
}

func (chain *BeaconChain) GetViewByHash(hash *common.Hash) (consensus.ChainViewInterface, error) {
	return nil, nil
}

func (chain *BeaconChain) GetFinalView() consensus.ChainViewInterface {
	return nil
}

func (chain *BeaconChain) storeView() error {
	return nil
}

func (chain *BeaconChain) deleteView(view consensus.ChainViewInterface) error {
	return nil
}

func (chain *BeaconChain) loadView() error {
	return nil
}

func (chain *BeaconChain) GetAllTipBlocksHash() []*common.Hash {
	result := []*common.Hash{}
	chain.lock.RLock()
	defer chain.lock.RUnlock()

	for _, view := range chain.views {
		result = append(result, view.GetTipBlock().Hash())
	}
	return result
}

func (chain *BeaconChain) AddView(view consensus.ChainViewInterface) error {
	chain.lock.Lock()
	defer chain.lock.Unlock()
	chain.addView(view.(*BeaconView))
	return nil
}

func (chain *BeaconChain) addView(view consensus.ChainViewInterface) error {
	chain.views[view.Hash().String()] = view.(*BeaconView)
	return nil
}

func (chain *BeaconChain) ConnectBlockAndAddView(block common.BlockInterface) error {
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
	chain.views[newView.Hash().String()] = newView.(*BeaconView)
	return nil
}
