package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/chain"
	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/common"
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

type BeaconView struct {
	BestBlockHash                          common.Hash                                `json:"BestBlockHash"`         // The hash of the block.
	PreviousBestBlockHash                  common.Hash                                `json:"PreviousBestBlockHash"` // The hash of the block.
	BestBlock                              BeaconBlock                                `json:"BestBlock"`             // The block.
	BestShardHash                          map[byte]common.Hash                       `json:"BestShardHash"`
	BestShardHeight                        map[byte]uint64                            `json:"BestShardHeight"`
	Epoch                                  uint64                                     `json:"Epoch"`
	BeaconHeight                           uint64                                     `json:"BeaconHeight"`
	BeaconProposerIndex                    int                                        `json:"BeaconProposerIndex"`
	BeaconCommitteeHash                    string                                     `json:"BeaconCommitteeHash"`
	BeaconCommittee                        []incognitokey.CommitteePublicKey          `json:"BeaconCommittee"`
	BeaconPendingValidator                 []incognitokey.CommitteePublicKey          `json:"BeaconPendingValidator"`
	CandidateShardWaitingForCurrentRandom  []incognitokey.CommitteePublicKey          `json:"CandidateShardWaitingForCurrentRandom"` // snapshot shard candidate list, waiting to be shuffled in this current epoch
	CandidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey          `json:"CandidateBeaconWaitingForCurrentRandom"`
	CandidateShardWaitingForNextRandom     []incognitokey.CommitteePublicKey          `json:"CandidateShardWaitingForNextRandom"` // shard candidate list, waiting to be shuffled in next epoch
	CandidateBeaconWaitingForNextRandom    []incognitokey.CommitteePublicKey          `json:"CandidateBeaconWaitingForNextRandom"`
	ShardCommittee                         map[byte][]incognitokey.CommitteePublicKey `json:"ShardCommittee"`        // current committee and validator of all shard
	ShardPendingValidator                  map[byte][]incognitokey.CommitteePublicKey `json:"ShardPendingValidator"` // pending candidate waiting for swap to get in committee of all shard
	AutoStaking                            map[string]bool                            `json:"AutoStaking"`
	CurrentRandomNumber                    int64                                      `json:"CurrentRandomNumber"`
	CurrentRandomTimeStamp                 int64                                      `json:"CurrentRandomTimeStamp"` // random timestamp for this epoch
	IsGetRandomNumber                      bool                                       `json:"IsGetRandomNumber"`
	MaxBeaconCommitteeSize                 int                                        `json:"MaxBeaconCommitteeSize"`
	MinBeaconCommitteeSize                 int                                        `json:"MinBeaconCommitteeSize"`
	MaxShardCommitteeSize                  int                                        `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize                  int                                        `json:"MinShardCommitteeSize"`
	ActiveShards                           int                                        `json:"ActiveShards"`
	ConsensusAlgorithm                     string                                     `json:"ConsensusAlgorithm"`
	ConsensusConfig                        string                                     `json:"ConsensusConfig"`

	ShardConsensusAlgorithm map[byte]string `json:"ShardConsensusAlgorithm"`
	// key: public key of committee, value: payment address reward receiver
	RewardReceiver map[string]string `json:"RewardReceiver"` // map incognito public key -> reward receiver (payment address)
	// cross shard state for all the shard. from shardID -> to crossShard shardID -> last height
	// e.g 1 -> 2 -> 3 // shard 1 send cross shard to shard 2 at  height 3
	// e.g 1 -> 3 -> 2 // shard 1 send cross shard to shard 3 at  height 2
	LastCrossShardState map[byte]map[byte]uint64 `json:"LastCrossShardState"`
	ShardHandle         map[byte]bool            `json:"ShardHandle"` // lock sync.RWMutex

	// Number of blocks produced by producers in epoch
	NumOfBlocksByProducers map[string]uint64 `json:"NumOfBlocksByProducers"`

	lock sync.RWMutex

	BlockInterval      time.Duration `json:"BlockInterval"`
	BlockMaxCreateTime time.Duration `json:"BlockMaxCreateTime"`
	GenesisTime        int64         `json:"GenesisTime"` //use for consensus to get timeslot
	isBest             bool
}

func NewBeaconView() *BeaconView {
	var view BeaconView
	return &view
}

func NewBeaconViewWithConfig(netparam *Params) *BeaconView {
	var view BeaconView
	view.BestBlockHash.SetBytes(make([]byte, 32))
	view.BestBlockHash.SetBytes(make([]byte, 32))
	view.BestShardHash = make(map[byte]common.Hash)
	view.BestShardHeight = make(map[byte]uint64)
	view.BeaconHeight = 0
	view.BeaconCommittee = []incognitokey.CommitteePublicKey{}
	view.BeaconPendingValidator = []incognitokey.CommitteePublicKey{}
	view.CandidateShardWaitingForCurrentRandom = []incognitokey.CommitteePublicKey{}
	view.CandidateBeaconWaitingForCurrentRandom = []incognitokey.CommitteePublicKey{}
	view.CandidateShardWaitingForNextRandom = []incognitokey.CommitteePublicKey{}
	view.CandidateBeaconWaitingForNextRandom = []incognitokey.CommitteePublicKey{}
	view.RewardReceiver = make(map[string]string)
	view.ShardCommittee = make(map[byte][]incognitokey.CommitteePublicKey)
	view.ShardPendingValidator = make(map[byte][]incognitokey.CommitteePublicKey)
	view.AutoStaking = make(map[string]bool)
	view.CurrentRandomNumber = -1
	view.MaxBeaconCommitteeSize = netparam.MaxBeaconCommitteeSize
	view.MinBeaconCommitteeSize = netparam.MinBeaconCommitteeSize
	view.MaxShardCommitteeSize = netparam.MaxShardCommitteeSize
	view.MinShardCommitteeSize = netparam.MinShardCommitteeSize
	view.ActiveShards = netparam.ActiveShards
	view.LastCrossShardState = make(map[byte]map[byte]uint64)
	view.BlockInterval = netparam.MinBeaconBlockInterval
	view.BlockMaxCreateTime = netparam.MaxBeaconBlockCreation
	return &view
}

func (view *BeaconView) MarshalJSON() ([]byte, error) {
	view.lock.RLock()
	defer view.lock.RUnlock()
	type Alias BeaconView
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

func (view *BeaconView) SetBestShardHeight(shardID byte, height uint64) {
	view.lock.RLock()
	defer view.lock.RUnlock()
	view.BestShardHeight[shardID] = height
}

func (view *BeaconView) GetShardConsensusAlgorithm() map[byte]string {
	view.lock.RLock()
	defer view.lock.RUnlock()
	res := make(map[byte]string)
	for index, element := range view.ShardConsensusAlgorithm {
		res[index] = element
	}
	return res
}

func (view *BeaconView) GetBestShardHash() map[byte]common.Hash {
	view.lock.RLock()
	defer view.lock.RUnlock()
	res := make(map[byte]common.Hash)
	for index, element := range view.BestShardHash {
		res[index] = element
	}
	return res
}

func (view *BeaconView) GetBestShardHeight() map[byte]uint64 {
	view.lock.RLock()
	defer view.lock.RUnlock()
	res := make(map[byte]uint64)
	for index, element := range view.BestShardHeight {
		res[index] = element
	}
	return res
}

func (view *BeaconView) GetBestHeightOfShard(shardID byte) uint64 {
	view.lock.RLock()
	defer view.lock.RUnlock()
	return view.BestShardHeight[shardID]
}

// GetAShardCommittee TODO
func (view *BeaconView) GetAShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	view.lock.RLock()
	defer view.lock.RUnlock()
	return view.ShardCommittee[shardID]
}

// GetShardCommittee TODO
func (view *BeaconView) GetShardCommittee() (res map[byte][]incognitokey.CommitteePublicKey) {
	view.lock.RLock()
	defer view.lock.RUnlock()
	res = make(map[byte][]incognitokey.CommitteePublicKey)
	for index, element := range view.ShardCommittee {
		res[index] = element
	}
	return res
}

func (view *BeaconView) GetAShardPendingValidator(shardID byte) []incognitokey.CommitteePublicKey {
	view.lock.RLock()
	defer view.lock.RUnlock()
	return view.ShardPendingValidator[shardID]
}

func (view *BeaconView) GetShardPendingValidator() (res map[byte][]incognitokey.CommitteePublicKey) {
	view.lock.RLock()
	defer view.lock.RUnlock()
	res = make(map[byte][]incognitokey.CommitteePublicKey)
	for index, element := range view.ShardPendingValidator {
		res[index] = element
	}
	return res
}

func (view *BeaconView) GetCurrentShard() byte {
	view.lock.RLock()
	defer view.lock.RUnlock()
	for shardID, isCurrent := range view.ShardHandle {
		if isCurrent {
			return shardID
		}
	}
	return 0
}

func (view *BeaconView) SetMaxShardCommitteeSize(maxShardCommitteeSize int) bool {
	view.lock.Lock()
	defer view.lock.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if maxShardCommitteeSize < MinCommitteeSize {
		return false
	}
	// max committee size can't be lower than current min committee size
	if maxShardCommitteeSize >= view.MinShardCommitteeSize {
		view.MaxShardCommitteeSize = maxShardCommitteeSize
		return true
	}
	return false
}

func (view *BeaconView) SetMinShardCommitteeSize(minShardCommitteeSize int) bool {
	view.lock.Lock()
	defer view.lock.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if minShardCommitteeSize < MinCommitteeSize {
		return false
	}
	// min committee size can't be greater than current min committee size
	if minShardCommitteeSize <= view.MaxShardCommitteeSize {
		view.MinShardCommitteeSize = minShardCommitteeSize
		return true
	}
	return false
}

func (view *BeaconView) SetMaxBeaconCommitteeSize(maxBeaconCommitteeSize int) bool {
	view.lock.Lock()
	defer view.lock.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if maxBeaconCommitteeSize < MinCommitteeSize {
		return false
	}
	// max committee size can't be lower than current min committee size
	if maxBeaconCommitteeSize >= view.MinBeaconCommitteeSize {
		view.MaxBeaconCommitteeSize = maxBeaconCommitteeSize
		return true
	}
	return false
}

func (view *BeaconView) SetMinBeaconCommitteeSize(minBeaconCommitteeSize int) bool {
	view.lock.Lock()
	defer view.lock.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if minBeaconCommitteeSize < MinCommitteeSize {
		return false
	}
	// min committee size can't be greater than current min committee size
	if minBeaconCommitteeSize <= view.MaxBeaconCommitteeSize {
		view.MinBeaconCommitteeSize = minBeaconCommitteeSize
		return true
	}
	return false
}
func (view *BeaconView) CheckCommitteeSize() error {
	if view.MaxBeaconCommitteeSize < MinCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect max beacon size %+v equal or greater than min size %+v", view.MaxBeaconCommitteeSize, MinCommitteeSize))
	}
	if view.MinBeaconCommitteeSize < MinCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect min beacon size %+v equal or greater than min size %+v", view.MinBeaconCommitteeSize, MinCommitteeSize))
	}
	if view.MaxShardCommitteeSize < MinCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect max shard size %+v equal or greater than min size %+v", view.MaxShardCommitteeSize, MinCommitteeSize))
	}
	if view.MinShardCommitteeSize < MinCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect min shard size %+v equal or greater than min size %+v", view.MinShardCommitteeSize, MinCommitteeSize))
	}
	if view.MaxBeaconCommitteeSize < view.MinBeaconCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect Max beacon size is higher than min beacon size but max is %+v and min is %+v", view.MaxBeaconCommitteeSize, view.MinBeaconCommitteeSize))
	}
	if view.MaxShardCommitteeSize < view.MinShardCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect Max beacon size is higher than min beacon size but max is %+v and min is %+v", view.MaxBeaconCommitteeSize, view.MinBeaconCommitteeSize))
	}
	return nil
}

func (view *BeaconView) GetBytes() []byte {
	view.lock.RLock()
	defer view.lock.RUnlock()
	var keys []int
	res := []byte{}
	res = append(res, view.BestBlockHash.GetBytes()...)
	res = append(res, view.PreviousBestBlockHash.GetBytes()...)
	res = append(res, view.BestBlock.Hash().GetBytes()...)
	res = append(res, view.BestBlock.Header.PreviousBlockHash.GetBytes()...)
	for k := range view.BestShardHash {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		hash := view.BestShardHash[byte(shardID)]
		res = append(res, hash.GetBytes()...)
	}
	keys = []int{}
	for k := range view.BestShardHeight {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		height := view.BestShardHeight[byte(shardID)]
		res = append(res, byte(height))
	}
	EpochBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(EpochBytes, view.Epoch)
	res = append(res, EpochBytes...)
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, view.BeaconHeight)
	res = append(res, heightBytes...)
	res = append(res, []byte(strconv.Itoa(view.BeaconProposerIndex))...)
	for _, value := range view.BeaconCommittee {
		valueBytes, err := value.Bytes()
		if err != nil {
			return nil
		}
		res = append(res, valueBytes...)
	}
	for _, value := range view.BeaconPendingValidator {
		valueBytes, err := value.Bytes()
		if err != nil {
			return nil
		}
		res = append(res, valueBytes...)
	}
	for _, value := range view.CandidateBeaconWaitingForCurrentRandom {
		valueBytes, err := value.Bytes()
		if err != nil {
			return nil
		}
		res = append(res, valueBytes...)
	}
	for _, value := range view.CandidateBeaconWaitingForNextRandom {
		valueBytes, err := value.Bytes()
		if err != nil {
			return nil
		}
		res = append(res, valueBytes...)
	}
	for _, value := range view.CandidateShardWaitingForCurrentRandom {
		valueBytes, err := value.Bytes()
		if err != nil {
			return nil
		}
		res = append(res, valueBytes...)
	}
	for _, value := range view.CandidateShardWaitingForNextRandom {
		valueBytes, err := value.Bytes()
		if err != nil {
			return nil
		}
		res = append(res, valueBytes...)
	}
	keys = []int{}
	for k := range view.ShardCommittee {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range view.ShardCommittee[byte(shardID)] {
			valueBytes, err := value.Bytes()
			if err != nil {
				return nil
			}
			res = append(res, valueBytes...)
		}
	}
	keys = []int{}
	for k := range view.ShardPendingValidator {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range view.ShardPendingValidator[byte(shardID)] {
			valueBytes, err := value.Bytes()
			if err != nil {
				return nil
			}
			res = append(res, valueBytes...)
		}
	}
	keysStrs2 := []string{}
	for k := range view.AutoStaking {
		keysStrs2 = append(keysStrs2, k)
	}
	sort.Strings(keysStrs2)
	for _, key := range keysStrs2 {
		if view.AutoStaking[key] {
			res = append(res, []byte("true")...)
		} else {
			res = append(res, []byte("false")...)
		}
	}
	randomNumBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(randomNumBytes, uint64(view.CurrentRandomNumber))
	res = append(res, randomNumBytes...)

	randomTimeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(randomTimeBytes, uint64(view.CurrentRandomTimeStamp))
	res = append(res, randomTimeBytes...)

	if view.IsGetRandomNumber {
		res = append(res, []byte("true")...)
	} else {
		res = append(res, []byte("false")...)
	}
	keys = []int{}
	for k := range view.ShardHandle {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		shardHandleItem := view.ShardHandle[byte(shardID)]
		if shardHandleItem {
			res = append(res, []byte("true")...)
		} else {
			res = append(res, []byte("false")...)
		}
	}
	res = append(res, []byte(strconv.Itoa(view.MaxBeaconCommitteeSize))...)
	res = append(res, []byte(strconv.Itoa(view.MinBeaconCommitteeSize))...)
	res = append(res, []byte(strconv.Itoa(view.MaxShardCommitteeSize))...)
	res = append(res, []byte(strconv.Itoa(view.MinShardCommitteeSize))...)
	res = append(res, []byte(strconv.Itoa(view.ActiveShards))...)

	keys = []int{}
	for k := range view.LastCrossShardState {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, fromShard := range keys {
		fromShardMap := view.LastCrossShardState[byte(fromShard)]
		newKeys := []int{}
		for k := range fromShardMap {
			newKeys = append(newKeys, int(k))
		}
		sort.Ints(newKeys)
		for _, toShard := range newKeys {
			value := fromShardMap[byte(toShard)]
			valueBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(valueBytes, value)
			res = append(res, valueBytes...)
		}
	}
	return res
}
func (view *BeaconView) Hash() common.Hash {
	return common.HashH(view.GetBytes())
}

// Get role of a public key base on best state beacond
// return node-role, <shardID>
func (view *BeaconView) GetPubkeyRole(pubkey string, round int) (string, byte) {
	view.lock.RLock()
	defer view.lock.RUnlock()
	for shardID, pubkeyArr := range view.ShardPendingValidator {
		keyList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(pubkeyArr, view.ShardConsensusAlgorithm[shardID])
		found := common.IndexOfStr(pubkey, keyList)
		if found > -1 {
			return common.ShardRole, shardID
		}
	}

	for shardID, pubkeyArr := range view.ShardCommittee {
		keyList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(pubkeyArr, view.ShardConsensusAlgorithm[shardID])
		found := common.IndexOfStr(pubkey, keyList)
		if found > -1 {
			return common.ShardRole, shardID
		}
	}

	keyList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(view.BeaconCommittee, view.ConsensusAlgorithm)
	found := common.IndexOfStr(pubkey, keyList)
	if found > -1 {
		tmpID := (view.BeaconProposerIndex + round) % len(view.BeaconCommittee)
		if found == tmpID {
			return common.ProposerRole, 0
		}
		return common.ValidatorRole, 0
	}

	keyList, _ = incognitokey.ExtractPublickeysFromCommitteeKeyList(view.BeaconPendingValidator, view.ConsensusAlgorithm)
	found = common.IndexOfStr(pubkey, keyList)
	if found > -1 {
		return common.PendingRole, 0
	}

	return common.EmptyString, 0
}

func (view *BeaconView) GetShardCandidate() []incognitokey.CommitteePublicKey {
	view.lock.RLock()
	defer view.lock.RUnlock()
	return append(view.CandidateShardWaitingForCurrentRandom, view.CandidateShardWaitingForNextRandom...)
}
func (view *BeaconView) GetBeaconCandidate() []incognitokey.CommitteePublicKey {
	view.lock.RLock()
	defer view.lock.RUnlock()
	return append(view.CandidateBeaconWaitingForCurrentRandom, view.CandidateBeaconWaitingForNextRandom...)
}
func (view *BeaconView) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	view.lock.RLock()
	defer view.lock.RUnlock()
	result := []incognitokey.CommitteePublicKey{}
	return append(result, view.BeaconCommittee...)
}
func (view *BeaconView) GetBeaconPendingValidator() []incognitokey.CommitteePublicKey {
	view.lock.RLock()
	defer view.lock.RUnlock()
	return view.BeaconPendingValidator
}
func (view *BeaconView) cloneBeaconViewFrom(target chain.ChainViewInterface) error {
	tempMarshal, err := json.Marshal(target)
	if err != nil {
		return NewBlockChainError(MashallJsonBeaconBestStateError, fmt.Errorf("Shard Best State %+v get %+v", view.BeaconHeight, err))
	}
	err = json.Unmarshal(tempMarshal, view)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBeaconBestStateError, fmt.Errorf("Clone Shard Best State %+v get %+v", view.BeaconHeight, err))
	}
	plainBeaconView := NewBeaconView()
	if reflect.DeepEqual(*view, plainBeaconView) {
		return NewBlockChainError(CloneBeaconBestStateError, fmt.Errorf("Shard Best State %+v clone failed", view.BeaconHeight))
	}
	return nil
}

func (view *BeaconView) CloneBeaconViewFrom(target chain.ChainViewInterface) error {
	return view.cloneBeaconViewFrom(target)
}
func (view *BeaconView) updateLastCrossShardState(shardStates map[byte][]ShardState) {
	lastCrossShardState := view.LastCrossShardState
	for fromShard, shardBlocks := range shardStates {
		for _, shardBlock := range shardBlocks {
			for _, toShard := range shardBlock.CrossShard {
				if fromShard == toShard {
					continue
				}
				if lastCrossShardState[fromShard] == nil {
					lastCrossShardState[fromShard] = make(map[byte]uint64)
				}
				waitHeight := shardBlock.Height
				lastCrossShardState[fromShard][toShard] = waitHeight
			}
		}
	}
}
func (view *BeaconView) UpdateLastCrossShardState(shardStates map[byte][]ShardState) {
	view.lock.Lock()
	defer view.lock.Unlock()
	view.updateLastCrossShardState(shardStates)
}

func (view *BeaconView) GetAutoStakingList() map[string]bool {
	view.lock.RLock()
	defer view.lock.RUnlock()
	m := make(map[string]bool)
	for k, v := range view.AutoStaking {
		m[k] = v
	}
	return m
}
func (view *BeaconView) GetAllCommitteeValidatorCandidateFlattenList() []string {
	view.lock.RLock()
	defer view.lock.RUnlock()
	return view.getAllCommitteeValidatorCandidateFlattenList()
}
func (view *BeaconView) getAllCommitteeValidatorCandidateFlattenList() []string {
	res := []string{}
	for _, committee := range view.ShardCommittee {
		committeeStr, err := incognitokey.CommitteeKeyListToString(committee)
		if err != nil {
			panic(err)
		}
		res = append(res, committeeStr...)
	}
	for _, pendingValidator := range view.ShardPendingValidator {
		pendingValidatorStr, err := incognitokey.CommitteeKeyListToString(pendingValidator)
		if err != nil {
			panic(err)
		}
		res = append(res, pendingValidatorStr...)
	}

	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(view.BeaconCommittee)
	if err != nil {
		panic(err)
	}
	res = append(res, beaconCommitteeStr...)

	beaconPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(view.BeaconPendingValidator)
	if err != nil {
		panic(err)
	}
	res = append(res, beaconPendingValidatorStr...)

	candidateBeaconWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(view.CandidateBeaconWaitingForCurrentRandom)
	if err != nil {
		panic(err)
	}
	res = append(res, candidateBeaconWaitingForCurrentRandomStr...)

	candidateBeaconWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(view.CandidateBeaconWaitingForNextRandom)
	if err != nil {
		panic(err)
	}
	res = append(res, candidateBeaconWaitingForNextRandomStr...)

	candidateShardWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(view.CandidateShardWaitingForCurrentRandom)
	if err != nil {
		panic(err)
	}
	res = append(res, candidateShardWaitingForCurrentRandomStr...)

	candidateShardWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(view.CandidateShardWaitingForNextRandom)
	if err != nil {
		panic(err)
	}
	res = append(res, candidateShardWaitingForNextRandomStr...)
	return res
}

func (view *BeaconView) GetConsensusType() string {
	return ""
}
func (view *BeaconView) GetCommitteeIndex(string) int {
	return 0
}
func (view *BeaconView) GetLastProposerIndex() int {
	return 0
}
func (view *BeaconView) CreateNewBlock(timeslot uint64) (common.BlockInterface, error) {
	return nil, nil
}
func (view *BeaconView) InsertBlk(block common.BlockInterface) error {
	return nil
}
func (view *BeaconView) InsertAndBroadcastBlock(block common.BlockInterface) error {
	return nil
}
func (view *BeaconView) ValidatePreSignBlock(block common.BlockInterface) error {
	return nil
}

func (view *BeaconView) DeleteView() error {
	return nil
}

func (view *BeaconView) StoreView() error {
	return nil
}

func (view *BeaconView) GetConsensusConfig() string {
	return view.ConsensusConfig
}

func (view *BeaconView) GetHeight() uint64 {
	return view.BestBlock.GetHeight()
}

func (view *BeaconView) GetBlkMaxCreateTime() time.Duration {
	return 0
}

func (view *BeaconView) GetBlkMinInterval() time.Duration {
	return 0
}

func (view *BeaconView) GetTimeStamp() int64 {
	return 0
}
func (view *BeaconView) GetCommittee() []incognitokey.CommitteePublicKey {
	result := []incognitokey.CommitteePublicKey{}
	result = append([]incognitokey.CommitteePublicKey{}, view.BeaconCommittee...)
	return result
}

func (view *BeaconView) GetTimeslot() uint64 { return 0 }

func (view *BeaconView) GetEpoch() uint64 {
	return 0
}

func (view *BeaconView) IsBestView() bool {
	return view.isBest
}

func (view *BeaconView) SetViewIsBest(isBest bool) {
	view.lock.Lock()
	defer view.lock.Unlock()
	view.isBest = isBest
}

func (view *BeaconView) GetTipBlock() common.BlockInterface {
	return view.BestBlock
}

func (view *BeaconView) GetCommitteeHash() *common.Hash {
	view.lock.RLock()
	defer view.lock.RUnlock()
	result, err := common.Hash{}.NewHashFromStr(view.BeaconCommitteeHash)
	if err != nil {
		panic(err)
	}
	return result
}

func (view *BeaconView) GetGenesisTime() int64 {
	return view.GenesisTime
}

func (view *BeaconView) GetPreviousViewHash() *common.Hash {
	return nil
}

func (view *BeaconView) UpdateViewWithBlock(block common.BlockInterface) error {
	return nil
}
func (view *BeaconView) CloneViewFrom(viewToClone *chain.ChainViewInterface) error {
	return nil
}

func (view *BeaconView) ValidateBlock(block common.BlockInterface, isPreSign bool) error {
	return nil
}
func (view *BeaconView) ConnectBlockAndCreateView(block common.BlockInterface) (chain.ChainViewInterface, error) {
	return nil, nil
}

func (view BeaconView) GetActiveShardNumber() int {
	panic("implement me")
}
