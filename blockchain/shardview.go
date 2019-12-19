package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
)

// BestState houses information about the current best block and other info
// related to the state of the main chain as it exists from the point of view of
// the current best block.
//
// The BestSnapshot method can be used to obtain access to this information
// in a concurrent safe manner and the data will not be changed out from under
// the caller when chain state changes occur as the function name implies.
// However, the returned snapshot must be treated as immutable since it is
// shared by all callers.

type ShardView struct {
	TipBlock               *ShardBlock                       `json:"TipBlock"` // block data
	ShardID                byte                              `json:"ShardID"`
	MaxCommitteeSize       int                               `json:"MaxCommitteeSize"`
	MinCommitteeSize       int                               `json:"MinCommitteeSize"`
	CommitteeHash          string                            `json:"CommitteeHash"`
	Committee              []incognitokey.CommitteePublicKey `json:"Committee"`
	PendingValidator       []incognitokey.CommitteePublicKey `json:"ShardPendingValidator"`
	BestCrossShard         map[byte]uint64                   `json:"BestCrossShard"` // Best cross shard block by heigh
	StakingTx              map[string]string                 `json:"StakingTx"`
	NumTxns                uint64                            `json:"NumTxns"`                // The number of txns in the block.
	TotalTxns              uint64                            `json:"TotalTxns"`              // The total number of txns in the chain.
	TotalTxnsExcludeSalary uint64                            `json:"TotalTxnsExcludeSalary"` // for testing and benchmark
	ActiveShards           int                               `json:"ActiveShards"`
	ConsensusAlgorithm     string                            `json:"ConsensusAlgorithm"`
	ConsensusConfig        string                            `json:"ConsensusConfig"`

	// Number of blocks produced by producers in epoch
	NumOfBlocksByProducers map[string]uint64 `json:"NumOfBlocksByProducers"`

	BlockInterval      time.Duration `json:"BlockInterval"`
	BlockMaxCreateTime time.Duration `json:"BlockMaxCreateTime"`

	GenesisTime int64 `json:"GenesisTime"` //use for consensus to get timeslot
	isBest      bool  `json:"IsBest"`
	// MetricBlockHeight uint64
	lock sync.RWMutex
}

func NewShardView() *ShardView {
	var view ShardView
	return &view
}
func NewShardViewWithConfig(shardID byte, netparam *Params) *ShardView {
	var view ShardView
	view.TipBlock = nil
	view.Committee = []incognitokey.CommitteePublicKey{}
	view.MaxCommitteeSize = netparam.MaxShardCommitteeSize
	view.MinCommitteeSize = netparam.MinShardCommitteeSize
	view.PendingValidator = []incognitokey.CommitteePublicKey{}
	view.ActiveShards = netparam.ActiveShards
	view.BestCrossShard = make(map[byte]uint64)
	view.StakingTx = make(map[string]string)
	view.BlockInterval = netparam.MinShardBlockInterval
	view.BlockMaxCreateTime = netparam.MaxShardBlockCreation
	view.GenesisTime = netparam.GenesisShardBlock.GetBlockTimestamp()
	view.isBest = false
	view.NumOfBlocksByProducers = make(map[string]uint64)
	return &view
}

// Get role of a public key base on best state shard
func (view *ShardView) GetBytes() []byte {
	res := []byte{}
	res = append(res, view.TipBlock.Hash().GetBytes()...)
	res = append(res, view.ShardID)
	CommitteeSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(CommitteeSizeBytes, uint32(view.MaxCommitteeSize))
	res = append(res, CommitteeSizeBytes...)
	MinCommitteeSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(MinCommitteeSizeBytes, uint32(view.MinCommitteeSize))
	res = append(res, MinCommitteeSizeBytes...)
	for _, value := range view.Committee {
		valueBytes, err := value.Bytes()
		if err != nil {
			return nil
		}
		res = append(res, valueBytes...)
	}
	for _, value := range view.PendingValidator {
		valueBytes, err := value.Bytes()
		if err != nil {
			return nil
		}
		res = append(res, valueBytes...)
	}
	keys := []int{}
	for k := range view.BestCrossShard {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		value := view.BestCrossShard[byte(shardID)]
		valueBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(valueBytes, value)
		res = append(res, valueBytes...)
	}
	keystr := []string{}
	for _, k := range view.StakingTx {
		keystr = append(keystr, k)
	}
	sort.Strings(keystr)
	for _, key := range keystr {
		value := view.StakingTx[key]
		res = append(res, []byte(key)...)
		res = append(res, []byte(value)...)
	}
	numTxnsBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(numTxnsBytes, view.NumTxns)
	res = append(res, numTxnsBytes...)
	totalTxnsBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(totalTxnsBytes, view.TotalTxns)
	res = append(res, totalTxnsBytes...)
	activeShardsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(activeShardsBytes, uint32(view.ActiveShards))
	res = append(res, activeShardsBytes...)
	return res
}

func (view *ShardView) Hash() common.Hash {
	view.lock.RLock()
	defer view.lock.RUnlock()
	return common.HashH(view.GetBytes())
}

func (view *ShardView) SetMaxCommitteeSize(MaxCommitteeSize int) bool {
	view.lock.Lock()
	defer view.lock.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if MaxCommitteeSize < MinCommitteeSize {
		return false
	}
	// max committee size can't be lower than current min committee size
	if MaxCommitteeSize >= view.MinCommitteeSize {
		view.MaxCommitteeSize = MaxCommitteeSize
		return true
	}
	return false
}

func (view *ShardView) SetMinCommitteeSize(MinCommitteeSize int) bool {
	view.lock.Lock()
	defer view.lock.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if MinCommitteeSize < MinCommitteeSize {
		return false
	}
	// min committee size can't be greater than current min committee size
	if MinCommitteeSize <= view.MaxCommitteeSize {
		view.MinCommitteeSize = MinCommitteeSize
		return true
	}
	return false
}

func (view *ShardView) GetPubkeyRole(pubkey string, round int) (string, byte) {
	// keyList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(view.Committee, view.ConsensusAlgorithm)
	// // fmt.Printf("pubkey %v key list %v\n\n\n\n", pubkey, keyList)
	// found := common.IndexOfStr(pubkey, keyList)
	// if found > -1 {
	// 	tmpID := (view.ShardProposerIdx + round) % len(keyList)
	// 	if found == tmpID {
	// 		return common.ProposerRole
	// 	} else {
	// 		return common.ValidatorRole
	// 	}
	// }

	// keyList, _ = incognitokey.ExtractPublickeysFromCommitteeKeyList(view.ShardPendingValidator, view.ConsensusAlgorithm)
	// found = common.IndexOfStr(pubkey, keyList)
	// if found > -1 {
	// 	return common.PendingRole
	// }
	return common.EmptyString, 0
}

func (view *ShardView) MarshalJSON() ([]byte, error) {
	view.lock.RLock()
	defer view.lock.RUnlock()
	type Alias ShardView
	b, err := json.Marshal(&struct {
		*Alias
	}{
		(*Alias)(view),
	})
	if err != nil {
		Logger.log.Error(err)
	}
	return b, err
}

func (view ShardView) GetBeaconHeight() uint64 {
	return view.TipBlock.GetBeaconHeight()
}

func (view *ShardView) cloneShardViewFrom(target ChainViewInterface) error {
	tempMarshal, err := json.Marshal(target)
	if err != nil {
		return NewBlockChainError(MashallJsonShardBestStateError, fmt.Errorf("Shard Best State %+v get %+v", target.GetHeight(), err))
	}
	err = json.Unmarshal(tempMarshal, view)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBestStateError, fmt.Errorf("Clone Shard Best State %+v get %+v", target.GetHeight(), err))
	}
	if reflect.DeepEqual(*view, ShardView{}) {
		return NewBlockChainError(CloneShardBestStateError, fmt.Errorf("Shard Best State %+v clone failed", target.GetHeight()))
	}
	return nil
}
func (view *ShardView) GetStakingTx() map[string]string {
	view.lock.RLock()
	defer view.lock.RUnlock()
	m := make(map[string]string)
	for k, v := range view.StakingTx {
		m[k] = v
	}
	return m
}

func (view *ShardView) GetConsensusType() string {
	return view.ConsensusAlgorithm
}
func (view *ShardView) CreateNewBlock(timeslot uint64) (common.BlockInterface, error) {
	return nil, nil
}

// func (view *ShardView) InsertBlk(block common.BlockInterface) error {
// 	return nil
// }
// func (view *ShardView) InsertAndBroadcastBlock(block common.BlockInterface) error {
// 	return nil
// }
// func (view *ShardView) ValidatePreSignBlock(block common.BlockInterface) error {
// 	return nil
// }

func (view *ShardView) DeleteView() error {
	return nil
}

func (view ShardView) GetConsensusConfig() string {
	return view.ConsensusConfig
}

func (view ShardView) GetHeight() uint64 {
	return view.TipBlock.GetHeight()
}

func (view ShardView) GetBlkMaxCreateTime() time.Duration {
	return view.BlockMaxCreateTime
}

func (view ShardView) GetBlkMinInterval() time.Duration {
	return view.BlockInterval
}

func (view ShardView) GetTimeStamp() int64 {
	return view.TipBlock.GetBlockTimestamp()
}
func (view ShardView) GetCommittee() []incognitokey.CommitteePublicKey {
	result := []incognitokey.CommitteePublicKey{}
	result = append([]incognitokey.CommitteePublicKey{}, view.Committee...)
	return result
}

func (view *ShardView) SetCommittee(newCommittee []incognitokey.CommitteePublicKey) error {
	if len(newCommittee) > view.MaxCommitteeSize {
		return NewBlockChainError(ShardCommitteeLengthAndCommitteeIndexError, fmt.Errorf("newCommittee lenght: %v MaxCommitteeSize: %v", len(newCommittee), view.MaxCommitteeSize))
	}
	if len(newCommittee) < view.MinCommitteeSize {
		return NewBlockChainError(ShardCommitteeLengthAndCommitteeIndexError, fmt.Errorf("newCommittee lenght: %v MaxCommitteeSize: %v", len(newCommittee), view.MinCommitteeSize))
	}
	res := []byte{}
	for _, value := range newCommittee {
		valueBytes, err := value.Bytes()
		if err != nil {
			return err
		}
		res = append(res, valueBytes...)
	}

	view.Committee = append([]incognitokey.CommitteePublicKey{}, newCommittee...)
	view.CommitteeHash = common.HashH(res).String()

	return nil
}

func (view ShardView) GetTimeslot() uint64 {
	return view.TipBlock.GetTimeslot()
}

func (view ShardView) GetEpoch() uint64 {
	return view.TipBlock.GetEpoch()
}

func (view ShardView) GetTxsInView() []metadata.Transaction {
	view.lock.RLock()
	defer view.lock.RUnlock()
	var result []metadata.Transaction
	copy(result, view.TipBlock.Body.Transactions)
	return result
}

func (view *ShardView) IsBestView() bool {
	return view.isBest
}

func (view *ShardView) SetViewIsBest(isBest bool) {
	view.lock.Lock()
	defer view.lock.Unlock()
	view.isBest = isBest
}

func (view *ShardView) GetTipBlock() common.BlockInterface {
	return view.TipBlock
}

func (view *ShardView) GetCommitteeHash() *common.Hash {
	view.lock.RLock()
	defer view.lock.RUnlock()
	result, err := common.Hash{}.NewHashFromStr(view.CommitteeHash)
	if err != nil {
		panic(err)
	}
	return result
}

func (view ShardView) GetGenesisTime() int64 {
	return view.GenesisTime
}

func (view ShardView) GetCommitteeIndex(string) int {
	return 0
}

func (view ShardView) GetPreviousViewHash() *common.Hash {
	return nil
}

func (view ShardView) UpdateViewWithBlock(block common.BlockInterface) error {
	return nil
}
func (view ShardView) CloneViewFrom(vue *common.ChainManagerInterface) error {
	return nil
}

func (view ShardView) ValidateBlock(block common.BlockInterface, isPreSign bool) error {
	return nil
}

func (view ShardView) ConnectBlockAndCreateView(block common.BlockInterface) (common.ChainViewInterface, error) {
	return nil, nil
}

func (view ShardView) GetActiveShardNumber() int {
	panic("implement me")
}
