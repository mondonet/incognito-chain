package blockchain

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/btc"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
)

/*
	// This function should receives block in consensus round
	// It verify validity of this function before sign it
	// This should be verify in the first round of consensus

	Step:
	1. Verify Pre proccessing data
	2. Retrieve FinalView for new block, store in local variable
	3. Update: process local FinalView with new block
	4. Verify Post processing: updated local FinalView and newblock

	Return:
	- No error: valid and can be sign
	- Error: invalid new block
*/
func (blockchain *BlockChain) VerifyPreSignBeaconBlock(beaconBlock *BeaconBlock, isPreSign bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	// Verify block only
	// Logger.log.Infof("BEACON | Verify block for signing process %d, with hash %+v", beaconBlock.Header.Height, *beaconBlock.Hash())
	// if err := blockchain.verifyPreProcessingBeaconBlock(beaconBlock, isPreSign); err != nil {
	// 	return err
	// }
	// // Verify block with previous best state
	// // Get FinalView of previous block == previous best state
	// // Clone best state value into new variable
	// beaconView := NewBeaconView()
	// if err := beaconView.cloneBeaconViewFrom(blockchain.FinalView.Beacon); err != nil {
	// 	return err
	// }
	// // Verify block with previous best state
	// // not verify agg signature in this function
	// if err := beaconView.verifyFinalViewWithBeaconBlock(beaconBlock, false, blockchain.config.ChainParams.Epoch); err != nil {
	// 	return err
	// }
	// // Update best state with new block
	// if err := beaconView.updateBeaconView(beaconBlock, blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.AssignOffset, blockchain.config.ChainParams.RandomTime); err != nil {
	// 	return err
	// }
	// // Post verififcation: verify new beaconstate with corresponding block
	// if err := beaconView.verifyPostProcessingBeaconBlock(beaconBlock, blockchain.config.RandomClient); err != nil {
	// 	return err
	// }
	// Logger.log.Infof("BEACON | Block %d, with hash %+v is VALID to be 🖊 signed", beaconBlock.Header.Height, *beaconBlock.Hash())
	return nil
}

func (blockchain *BlockChain) InsertBeaconBlock(beaconBlock *BeaconBlock, isValidated bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()

	// currentBeaconView := blockchain.FinalView.Beacon
	// if currentBeaconView.BeaconHeight == beaconBlock.Header.Height && currentBeaconView.BestBlock.Header.Timestamp < beaconBlock.Header.Timestamp && currentBeaconView.BestBlock.Header.Round < beaconBlock.Header.Round {
	// 	currentBeaconHeight, currentBeaconHash := currentBeaconView.BeaconHeight, currentBeaconView.BestBlockHash
	// 	Logger.log.Infof("FORK BEACON, Current Beacon Block Height %+v, Hash %+v | Try to Insert New Beacon Block Height %+v, Hash %+v", currentBeaconView.BeaconHeight, currentBeaconView.BestBlockHash, beaconBlock.Header.Height, beaconBlock.Header.Hash())
	// 	if err := blockchain.ValidateBlockWithPrevBeaconView(beaconBlock); err != nil {
	// 		Logger.log.Error(err)
	// 		return err
	// 	}
	// 	if err := blockchain.RevertBeaconState(); err != nil {
	// 		panic(err)
	// 	}
	// 	blockchain.FinalView.Beacon.lock.Unlock()
	// 	Logger.log.Infof("REVERTED BEACON, Revert Current Beacon Block Height %+v, Hash %+v", currentBeaconHeight, currentBeaconHash)
	// }

	// if beaconBlock.Header.Height != currentBeaconView.BeaconHeight+1 {
	// 	return errors.New("Not expected height")
	// }

	// blockHash := beaconBlock.Header.Hash()
	// Logger.log.Infof("BEACON | Begin insert new Beacon Block height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// Logger.log.Infof("BEACON | Check Beacon Block existence before insert block height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// isExist, _ := blockchain.config.DataBase.HasBeaconBlock(beaconBlock.Header.Hash())
	// if isExist {
	// 	return NewBlockChainError(DuplicateShardBlockError, errors.New("This beaconBlock has been stored already"))
	// }
	// Logger.log.Infof("BEACON | Begin Insert new Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// if !isValidated {
	// 	Logger.log.Infof("BEACON | Verify Pre Processing, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// 	if err := blockchain.verifyPreProcessingBeaconBlock(beaconBlock, false); err != nil {
	// 		return err
	// 	}
	// } else {
	// 	Logger.log.Infof("BEACON | SKIP Verify Pre Processing, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// }
	// // Verify beaconBlock with previous best state
	// if !isValidated {
	// 	Logger.log.Infof("BEACON | Verify Best State With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// 	// Verify beaconBlock with previous best state
	// 	if err := blockchain.FinalView.Beacon.verifyFinalViewWithBeaconBlock(beaconBlock, true, blockchain.config.ChainParams.Epoch); err != nil {
	// 		return err
	// 	}
	// } else {
	// 	Logger.log.Infof("BEACON | SKIP Verify Best State With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// }
	// // Backup finalView
	// err := blockchain.config.DataBase.CleanBackup(true, 0)
	// if err != nil {
	// 	return NewBlockChainError(CleanBackUpError, err)
	// }
	// err = blockchain.BackupCurrentBeaconState(beaconBlock)
	// if err != nil {
	// 	return NewBlockChainError(BackUpFinalViewError, err)
	// }
	// // process for slashing, make sure this one is called before update best state
	// // since we'd like to process with old committee, not updated committee
	// slashErr := blockchain.processForSlashing(beaconBlock)
	// if slashErr != nil {
	// 	Logger.log.Errorf("Failed to process slashing with error: %+v", NewBlockChainError(ProcessSlashingError, slashErr))
	// }

	// // snapshot current beacon committee and shard committee
	// snapshotBeaconCommittee, snapshotAllShardCommittee, err := snapshotCommittee(blockchain.FinalView.Beacon.BeaconCommittee, blockchain.FinalView.Beacon.ShardCommittee)
	// if err != nil {
	// 	return NewBlockChainError(SnapshotCommitteeError, err)
	// }
	// _, snapshotAllShardPending, err := snapshotCommittee([]incognitokey.CommitteePublicKey{}, blockchain.FinalView.Beacon.ShardPendingValidator)
	// if err != nil {
	// 	return NewBlockChainError(SnapshotCommitteeError, err)
	// }

	// snapshotShardWaiting := append([]incognitokey.CommitteePublicKey{}, blockchain.FinalView.Beacon.CandidateShardWaitingForNextRandom...)
	// snapshotShardWaiting = append(snapshotShardWaiting, blockchain.FinalView.Beacon.CandidateBeaconWaitingForCurrentRandom...)

	// snapshotRewardReceiver, err := snapshotRewardReceiver(blockchain.FinalView.Beacon.RewardReceiver)
	// if err != nil {
	// 	return NewBlockChainError(SnapshotRewardReceiverError, err)
	// }
	// Logger.log.Infof("BEACON | Update FinalView With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// // Update best state with new beaconBlock

	// // NEWCONSENSUS dont need this
	// // if err := blockchain.FinalView.Beacon.updateBeaconView(beaconBlock, blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.AssignOffset, blockchain.config.ChainParams.RandomTime); err != nil {
	// // 	errRevert := blockchain.revertBeaconView()
	// // 	if errRevert != nil {
	// // 		return errors.WithStack(errRevert)
	// // 	}
	// // 	return err
	// // }
	// // updateNumOfBlocksByProducers updates number of blocks produced by producers
	// blockchain.FinalView.Beacon.updateNumOfBlocksByProducers(beaconBlock, blockchain.config.ChainParams.Epoch)

	// newBeaconCommittee, newAllShardCommittee, err := snapshotCommittee(blockchain.FinalView.Beacon.BeaconCommittee, blockchain.FinalView.Beacon.ShardCommittee)
	// if err != nil {
	// 	return NewBlockChainError(SnapshotCommitteeError, err)
	// }
	// _, newAllShardPending, err := snapshotCommittee([]incognitokey.CommitteePublicKey{}, blockchain.FinalView.Beacon.ShardPendingValidator)
	// if err != nil {
	// 	return NewBlockChainError(SnapshotCommitteeError, err)
	// }

	// Notify highway when there's a change (beacon/shard committee or beacon/shard pending members); for masternode only
	// TODO(@0xbunyip): check case changing pending beacon validators
	// notifyHighway := false

	// newShardWaiting := append([]incognitokey.CommitteePublicKey{}, blockchain.BestState.Beacon.CandidateShardWaitingForNextRandom...)
	// newShardWaiting = append(newShardWaiting, blockchain.BestState.Beacon.CandidateBeaconWaitingForCurrentRandom...)

	// isChanged := !reflect.DeepEqual(snapshotBeaconCommittee, newBeaconCommittee)
	// if isChanged {
	// 	go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
	// 	notifyHighway = true
	// }

	// isChanged = !reflect.DeepEqual(snapshotShardWaiting, newShardWaiting)
	// if isChanged {
	// 	go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
	// }

	//Check shard-pending
	// for shardID, committee := range newAllShardPending {
	// 	if _, ok := snapshotAllShardPending[shardID]; ok {
	// 		isChanged := !reflect.DeepEqual(snapshotAllShardPending[shardID], committee)
	// 		if isChanged {
	// 			go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
	// 			notifyHighway = true
	// 		}
	// 	} else {
	// 		go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
	// 		notifyHighway = true
	// 	}
	// }
	// //Check shard-committee
	// for shardID, committee := range newAllShardCommittee {
	// 	if _, ok := snapshotAllShardCommittee[shardID]; ok {
	// 		isChanged := !reflect.DeepEqual(snapshotAllShardCommittee[shardID], committee)
	// 		if isChanged {
	// 			go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
	// 			notifyHighway = true
	// 		}
	// 	} else {
	// 		go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
	// 		notifyHighway = true
	// 	}
	// }

	// if !isValidated {
	// 	Logger.log.Infof("BEACON | Verify Post Processing Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// 	// Post verification: verify new beacon best state with corresponding beacon block
	// 	if err := blockchain.FinalView.Beacon.verifyPostProcessingBeaconBlock(beaconBlock, blockchain.config.RandomClient); err != nil {
	// 		return err
	// 	}
	// } else {
	// 	Logger.log.Infof("BEACON | SKIP Verify Post Processing Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// }
	// Logger.log.Infof("Finish Insert new Beacon Block %+v, with hash %+v \n", beaconBlock.Header.Height, *beaconBlock.Hash())
	// if beaconBlock.Header.Height%50 == 0 {
	// 	BLogger.log.Debugf("Inserted beacon height: %d", beaconBlock.Header.Height)
	// }
	// go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewBeaconBlockTopic, beaconBlock))
	// go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.BeaconBeststateTopic, blockchain.BestState.Beacon))

	// // For masternode: broadcast new committee to highways
	// if notifyHighway {
	// 	go blockchain.config.Highway.BroadcastCommittee(
	// 		blockchain.config.ChainParams.Epoch,
	// 		newBeaconCommittee,
	// 		newAllShardCommittee,
	// 		newAllShardPending,
	// 	)
	// }
	return nil
}

// updateNumOfBlocksByProducers updates number of blocks produced by producers
func (beaconView *BeaconView) updateNumOfBlocksByProducers(beaconBlock *BeaconBlock, chainParamEpoch uint64) {
	producer := beaconBlock.GetProducerPubKeyStr()
	if beaconBlock.GetHeight()%chainParamEpoch == 1 {
		beaconView.NumOfBlocksByProducers = map[string]uint64{
			producer: 1,
		}
	}
	// Update number of blocks produced by producers in epoch
	numOfBlks, found := beaconView.NumOfBlocksByProducers[producer]
	if !found {
		beaconView.NumOfBlocksByProducers[producer] = 1
	} else {
		beaconView.NumOfBlocksByProducers[producer] = numOfBlks + 1
	}
}

func (blockchain *BlockChain) removeOldDataAfterProcessingBeaconBlock() {
	//=========Remove beacon beaconBlock in pool
	// go blockchain.config.BeaconPool.SetBeaconState(blockchain.FinalView.Beacon.BeaconHeight)
	// go blockchain.config.BeaconPool.RemoveBlock(blockchain.FinalView.Beacon.BeaconHeight)
	// //=========Remove shard to beacon beaconBlock in pool

	// go func() {
	// 	shardHeightMap := blockchain.FinalView.Beacon.GetBestShardHeight()
	// 	//force release readLock first, before execute the params in below function (which use same readLock).
	// 	//if writeLock occur before release, readLock will be block
	// 	blockchain.config.ShardToBeaconPool.SetShardState(shardHeightMap)
	// }()
}

/*
	VerifyPreProcessingBeaconBlock
	This function DOES NOT verify new block with best state
	DO NOT USE THIS with GENESIS BLOCK
	- Producer sanity data
	- Version: compatible with predefined version
	- Previous Block exist in database, fetch previous block by previous hash of new beacon block
	- Check new beacon block height is equal to previous block height + 1
	- Epoch = blockHeight % Epoch == 1 ? Previous Block Epoch + 1 : Previous Block Epoch
	- Timestamp of new beacon block is greater than previous beacon block timestamp
	- ShardStateHash: rebuild shard state hash from shard state body and compare with shard state hash in block header
	- InstructionHash: rebuild instruction hash from instruction body and compare with instruction hash in block header
	- InstructionMerkleRoot: rebuild instruction merkle root from instruction body and compare with instruction merkle root in block header
	- If verify block for signing then verifyPreProcessingBeaconBlockForSigning
*/
func (blockchain *BlockChain) verifyPreProcessingBeaconBlock(beaconBlock *BeaconBlock, isPreSign bool) error {
	// beaconLock := &blockchain.FinalView.Beacon.lock
	// beaconLock.RLock()
	// defer beaconLock.RUnlock()

	// // if len(beaconBlock.Header.Producer) == 0 {
	// // 	return NewBlockChainError(ProducerError, fmt.Errorf("Expect has length 66 but get %+v", len(beaconBlock.Header.Producer)))
	// // }
	// //verify version
	// if beaconBlock.Header.Version != BEACON_BLOCK_VERSION {
	// 	return NewBlockChainError(WrongVersionError, fmt.Errorf("Expect block version to be equal to %+v but get %+v", BEACON_BLOCK_VERSION, beaconBlock.Header.Version))
	// }
	//verify version
	// if beaconBlock.Header.Version != BEACON_BLOCK_VERSION {
	// 	return NewBlockChainError(WrongVersionError, fmt.Errorf("Expect block version to be equal to %+v but get %+v", BEACON_BLOCK_VERSION, beaconBlock.Header.Version))
	// }
	// // Verify parent hash exist or not
	// previousBlockHash := beaconBlock.Header.PreviousBlockHash
	// parentBlockBytes, err := blockchain.config.DataBase.FetchBeaconBlock(previousBlockHash)
	// if err != nil {
	// 	Logger.log.Criticalf("FORK BEACON DETECTED, New Beacon Block Height %+v, Hash %+v, Expected Previous Hash %+v, BUT Current Best State Height %+v and Hash %+v", beaconBlock.Header.Height, beaconBlock.Header.Hash(), beaconBlock.Header.PreviousBlockHash, blockchain.BestState.Beacon.BeaconHeight, blockchain.BestState.Beacon.BestBlockHash)
	// 	blockchain.Synker.SyncBlkBeacon(true, false, false, []common.Hash{previousBlockHash}, nil, 0, 0, "")
	// 	if !isPreSign {
	// 		revertErr := blockchain.revertBeaconState()
	// 		if revertErr != nil {
	// 			return errors.WithStack(revertErr)
	// 		}
	// 	}
	// 	return NewBlockChainError(FetchBeaconBlockError, err)
	// }
	// previousBeaconBlock := NewBeaconBlock()
	// err = json.Unmarshal(parentBlockBytes, previousBeaconBlock)
	// if err != nil {
	// 	return NewBlockChainError(UnmashallJsonBeaconBlockError, fmt.Errorf("Failed to unmarshall parent block of block height %+v", beaconBlock.Header.Height))
	// }
	// // Verify block height with parent block
	// if previousBeaconBlock.Header.Height+1 != beaconBlock.Header.Height {
	// 	return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Expect receive beacon block height %+v but get %+v", previousBeaconBlock.Header.Height+1, beaconBlock.Header.Height))
	// }
	// // Verify epoch with parent block
	// if (beaconBlock.Header.Height != 1) && (beaconBlock.Header.Height%blockchain.config.ChainParams.Epoch == 1) && (previousBeaconBlock.Header.Epoch != beaconBlock.Header.Epoch-1) {
	// 	return NewBlockChainError(WrongEpochError, fmt.Errorf("Expect receive beacon block epoch %+v greater than previous block epoch %+v, 1 value", beaconBlock.Header.Epoch, previousBeaconBlock.Header.Epoch))
	// }
	// // Verify timestamp with parent block
	// if beaconBlock.Header.Timestamp <= previousBeaconBlock.Header.Timestamp {
	// 	return NewBlockChainError(WrongTimestampError, fmt.Errorf("Expect receive beacon block with timestamp %+v greater than previous block timestamp %+v", beaconBlock.Header.Timestamp, previousBeaconBlock.Header.Timestamp))
	// }
	// if !verifyHashFromShardState(beaconBlock.Body.ShardState, beaconBlock.Header.ShardStateHash) {
	// 	return NewBlockChainError(ShardStateHashError, fmt.Errorf("Expect shard state hash to be %+v", beaconBlock.Header.ShardStateHash))
	// }
	// tempInstructionArr := []string{}
	// for _, strs := range beaconBlock.Body.Instructions {
	// 	tempInstructionArr = append(tempInstructionArr, strs...)
	// }
	// if hash, ok := verifyHashFromStringArray(tempInstructionArr, beaconBlock.Header.InstructionHash); !ok {
	// 	return NewBlockChainError(InstructionHashError, fmt.Errorf("Expect instruction hash to be %+v but get %+v", beaconBlock.Header.InstructionHash, hash))
	// }
	// // Shard state must in right format
	// // state[i].Height must less than state[i+1].Height and state[i+1].Height - state[i].Height = 1
	// for _, shardStates := range beaconBlock.Body.ShardState {
	// 	for i := 0; i < len(shardStates)-2; i++ {
	// 		if shardStates[i+1].Height-shardStates[i].Height != 1 {
	// 			return NewBlockChainError(ShardStateError, fmt.Errorf("Expect Shard State Height to be in the right format, %+v, %+v", shardStates[i+1].Height, shardStates[i].Height))
	// 		}
	// 	}
	// }
	// // Check if InstructionMerkleRoot is the root of merkle tree containing all instructions in this block
	// flattenInsts, err := FlattenAndConvertStringInst(beaconBlock.Body.Instructions)
	// if err != nil {
	// 	return NewBlockChainError(FlattenAndConvertStringInstError, err)
	// }
	// root := GetKeccak256MerkleRoot(flattenInsts)
	// if !bytes.Equal(root, beaconBlock.Header.InstructionMerkleRoot[:]) {
	// 	return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Expect Instruction Merkle Root in Beacon Block Header to be %+v but get %+v", string(beaconBlock.Header.InstructionMerkleRoot[:]), string(root)))
	// }
	// // if pool does not have one of needed block, fail to verify
	// if isPreSign {
	// 	if err := blockchain.verifyPreProcessingBeaconBlockForSigning(beaconBlock); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

/*
	verifyPreProcessingBeaconBlockForSigning
	Must pass these following condition:
	- Rebuild Reward By Epoch Instruction
	- Get All Shard To Beacon Block in Shard To Beacon Pool
	- For all Shard To Beacon Blocks in each Shard
		+ Compare all shard height of shard states in body and these Shard To Beacon Blocks (got from pool)
			* Must be have the same range of height
			* Compare CrossShardBitMap of each Shard To Beacon Block and Shard State in New Beacon Block Body
		+ After finish comparing these shard to beacon blocks with shard states in new beacon block body
			* Verifying Shard To Beacon Block Agg Signature
			* Only accept block in one epoch
		+ Get Instruction from these Shard To Beacon Blocks:
			* Stake Instruction
			* Swap Instruction
			* Bridge Instruction
			* Block Reward Instruction
		+ Generate Instruction Hash from all recently got instructions
		+ Compare just created Instruction Hash with Instruction Hash In Beacon Header
*/
func (blockchain *BlockChain) verifyPreProcessingBeaconBlockForSigning(beaconBlock *BeaconBlock) error {
	// var err error
	// rewardByEpochInstruction := [][]string{}
	// tempShardStates := make(map[byte][]ShardState)
	// stakeInstructions := [][]string{}
	// validStakePublicKeys := []string{}
	// swapInstructions := make(map[byte][][]string)
	// bridgeInstructions := [][]string{}
	// acceptedBlockRewardInstructions := [][]string{}
	// stopAutoStakingInstructions := [][]string{}
	// statefulActionsByShardID := map[byte][][]string{}
	// // Get Reward Instruction By Epoch
	// if beaconBlock.Header.Height%blockchain.config.ChainParams.Epoch == 1 {
	// 	rewardByEpochInstruction, err = blockchain.BuildRewardInstructionByEpoch(beaconBlock.Header.Height, beaconBlock.Header.Epoch-1)
	// 	if err != nil {
	// 		return NewBlockChainError(BuildRewardInstructionError, err)
	// 	}
	// }
	// // get shard to beacon blocks from pool
	// allShardBlocks := blockchain.config.ShardToBeaconPool.GetValidBlock(nil)

	// var keys []int
	// for k := range beaconBlock.Body.ShardState {
	// 	keys = append(keys, int(k))
	// }
	// sort.Ints(keys)

	// for _, value := range keys {
	// 	shardID := byte(value)
	// 	shardBlocks, ok := allShardBlocks[shardID]
	// 	shardStates := beaconBlock.Body.ShardState[shardID]
	// 	if !ok && len(shardStates) > 0 {
	// 		return NewBlockChainError(GetShardToBeaconBlocksError, fmt.Errorf("Expect to get from pool ShardToBeacon Block from Shard %+v but failed", shardID))
	// 	}
	// 	// repeatly compare each shard to beacon block and shard state in new beacon block body
	// 	if len(shardBlocks) >= len(shardStates) {
	// 		shardBlocks = shardBlocks[:len(beaconBlock.Body.ShardState[shardID])]
	// 		for index, shardState := range shardStates {
	// 			if shardBlocks[index].Header.Height != shardState.Height {
	// 				return NewBlockChainError(ShardStateHeightError, fmt.Errorf("Expect shard state height to be %+v but get %+v from pool", shardState.Height, shardBlocks[index].Header.Height))
	// 				return NewBlockChainError(ShardStateHeightError, fmt.Errorf("Expect shard state height to be %+v but get %+v from pool(shard %v)", shardState.Height, shardBlocks[index].Header.Height, shardID))
	// 			}
	// 			blockHash := shardBlocks[index].Header.Hash()
	// 			if !blockHash.IsEqual(&shardState.Hash) {
	// 				return NewBlockChainError(ShardStateHashError, fmt.Errorf("Expect shard state height %+v has hash %+v but get %+v from pool", shardState.Height, shardState.Hash, shardBlocks[index].Header.Hash()))
	// 			}
	// 			if !reflect.DeepEqual(shardBlocks[index].Header.CrossShardBitMap, shardState.CrossShard) {
	// 				return NewBlockChainError(ShardStateCrossShardBitMapError, fmt.Errorf("Expect shard state height %+v has bitmap %+v but get %+v from pool", shardState.Height, shardState.CrossShard, shardBlocks[index].Header.CrossShardBitMap))
	// 			}
	// 		}
	// 		// Only accept block in one epoch
	// 		for _, shardBlock := range shardBlocks {
	// 			currentCommittee := blockchain.FinalView.Beacon.GetAShardCommittee(shardID)
	// 			errValidation := blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(shardBlock, currentCommittee, blockchain.FinalView.Beacon.ShardConsensusAlgorithm[shardID])
	// 			if errValidation != nil {
	// 				return NewBlockChainError(ShardStateError, fmt.Errorf("Fail to verify with Shard To Beacon Block %+v, error %+v", shardBlock.Header.Height, err))
	// 			}
	// 		}
	// 		for _, shardBlock := range shardBlocks {
	// 			tempShardState, stakeInstruction, tempValidStakePublicKeys, swapInstruction, bridgeInstruction, acceptedBlockRewardInstruction, stopAutoStakingInstruction, statefulActions := blockchain.GetShardStateFromBlock(beaconBlock.Header.Height, shardBlock, shardID, false, validStakePublicKeys)
	// 			tempShardStates[shardID] = append(tempShardStates[shardID], tempShardState[shardID])
	// 			stakeInstructions = append(stakeInstructions, stakeInstruction...)
	// 			swapInstructions[shardID] = append(swapInstructions[shardID], swapInstruction[shardID]...)
	// 			bridgeInstructions = append(bridgeInstructions, bridgeInstruction...)
	// 			acceptedBlockRewardInstructions = append(acceptedBlockRewardInstructions, acceptedBlockRewardInstruction)
	// 			stopAutoStakingInstructions = append(stopAutoStakingInstructions, stopAutoStakingInstruction...)
	// 			validStakePublicKeys = append(validStakePublicKeys, tempValidStakePublicKeys...)

	// 			// group stateful actions by shardID
	// 			_, found := statefulActionsByShardID[shardID]
	// 			if !found {
	// 				statefulActionsByShardID[shardID] = statefulActions
	// 			} else {
	// 				statefulActionsByShardID[shardID] = append(statefulActionsByShardID[shardID], statefulActions...)
	// 			}
	// 		}
	// 	} else {
	// 		return NewBlockChainError(GetShardToBeaconBlocksError, fmt.Errorf("Expect to get more than %+v ShardToBeaconBlock but only get %+v (shard %v)", len(beaconBlock.Body.ShardState[shardID]), len(shardBlocks), shardID))
	// 	}
	// }
	// // build stateful instructions
	// statefulInsts := blockchain.buildStatefulInstructions(
	// 	statefulActionsByShardID,
	// 	beaconBlock.Header.Height,
	// 	blockchain.GetDatabase(),
	// )
	// bridgeInstructions = append(bridgeInstructions, statefulInsts...)

	// tempInstruction, err := blockchain.FinalView.Beacon.GenerateInstruction(beaconBlock.Header.Height,
	// 	stakeInstructions, swapInstructions, stopAutoStakingInstructions,
	// 	blockchain.FinalView.Beacon.CandidateShardWaitingForCurrentRandom,
	// 	bridgeInstructions, acceptedBlockRewardInstructions,
	// 	blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.RandomTime, blockchain)
	// if err != nil {
	// 	return err
	// }
	// if len(rewardByEpochInstruction) != 0 {
	// 	tempInstruction = append(tempInstruction, rewardByEpochInstruction...)
	// }
	// tempInstructionArr := []string{}
	// for _, strs := range tempInstruction {
	// 	tempInstructionArr = append(tempInstructionArr, strs...)
	// }
	// tempInstructionHash, err := generateHashFromStringArray(tempInstructionArr)
	// if err != nil {
	// 	return NewBlockChainError(GenerateInstructionHashError, fmt.Errorf("Fail to generate hash for instruction %+v", tempInstructionArr))
	// }
	// if !tempInstructionHash.IsEqual(&beaconBlock.Header.InstructionHash) {
	// 	return NewBlockChainError(InstructionHashError, fmt.Errorf("Expect Instruction Hash in Beacon Header to be %+v, but get %+v, validator instructions: %+v", beaconBlock.Header.InstructionHash, tempInstructionHash, tempInstruction))
	// }
	return nil
}

/*
	This function will verify the validation of a block with some best state in cache or current best state
	Get beacon state of this block
	For example, new blockHeight is 91 then beacon state of this block must have height 90
	OR new block has previous has is beacon best block hash
	- Get producer via index and compare with producer address in beacon block header
	- Validate public key and signature sanity
	- Validate Agg Signature
	- Beacon Best State has best block is previous block of new beacon block
	- Beacon Best State has height compatible with new beacon block
	- Beacon Best State has epoch compatible with new beacon block
	- Beacon Best State has best shard height compatible with shard state of new beacon block
	- New Stake public key must not found in beacon best state (candidate, pending validator, committee)
*/
func (beaconView *BeaconView) verifyFinalViewWithBeaconBlock(beaconBlock *BeaconBlock, isVerifySig bool, chainParamEpoch uint64) error {
	beaconView.lock.RLock()
	defer beaconView.lock.RUnlock()
	//verify producer via index
	producerPublicKey := beaconBlock.Header.Producer
	producerPosition := (beaconView.BeaconProposerIndex + beaconBlock.Header.Round) % len(beaconView.BeaconCommittee)
	tempProducer, err := beaconView.BeaconCommittee[producerPosition].ToBase58() //.GetMiningKeyBase58(common.BridgeConsensus)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	if strings.Compare(string(tempProducer), producerPublicKey) != 0 {
		return NewBlockChainError(BeaconBlockProducerError, fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", tempProducer, producerPublicKey))
	}

	//=============End Verify Aggegrate signature
	if !beaconView.BestBlockHash.IsEqual(&beaconBlock.Header.PreviousBlockHash) {
		return NewBlockChainError(BeaconBestStateBestBlockNotCompatibleError, errors.New("previous us block should be :"+beaconView.BestBlockHash.String()))
	}
	if beaconView.BeaconHeight+1 != beaconBlock.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(beaconBlock.Header.Height+1))))
	}
	if beaconBlock.Header.Height%chainParamEpoch == 1 && beaconView.Epoch+1 != beaconBlock.Header.Epoch {
		return NewBlockChainError(WrongEpochError, fmt.Errorf("Expect beacon block height %+v has epoch %+v but get %+v", beaconBlock.Header.Height, beaconView.Epoch+1, beaconBlock.Header.Epoch))
	}
	if beaconBlock.Header.Height%chainParamEpoch != 1 && beaconView.Epoch != beaconBlock.Header.Epoch {
		return NewBlockChainError(WrongEpochError, fmt.Errorf("Expect beacon block height %+v has epoch %+v but get %+v", beaconBlock.Header.Height, beaconView.Epoch, beaconBlock.Header.Epoch))
	}
	// check shard states of new beacon block and beacon best state
	// shard state of new beacon block must be greater or equal to current best shard height
	for shardID, shardStates := range beaconBlock.Body.ShardState {
		if bestShardHeight, ok := beaconView.BestShardHeight[shardID]; !ok {
			if shardStates[0].Height != 2 {
				return NewBlockChainError(BeaconBestStateBestShardHeightNotCompatibleError, fmt.Errorf("Shard %+v best height not found in beacon best state", shardID))
			}
		} else {
			if len(shardStates) > 0 {
				if bestShardHeight > shardStates[0].Height {
					return NewBlockChainError(BeaconBestStateBestShardHeightNotCompatibleError, fmt.Errorf("Expect Shard %+v has state greater than or equal to %+v but get %+v", shardID, bestShardHeight, shardStates[0].Height))
				}
				if bestShardHeight < shardStates[0].Height && bestShardHeight+1 != shardStates[0].Height {
					return NewBlockChainError(BeaconBestStateBestShardHeightNotCompatibleError, fmt.Errorf("Expect Shard %+v has state %+v but get %+v", shardID, bestShardHeight+1, shardStates[0].Height))
				}
			}
		}
	}
	//=============Verify Stake Public Key
	newBeaconCandidate, newShardCandidate := GetStakingCandidate(*beaconBlock)
	if !reflect.DeepEqual(newBeaconCandidate, []string{}) {
		validBeaconCandidate := beaconView.GetValidStakers(newBeaconCandidate)
		if !reflect.DeepEqual(validBeaconCandidate, newBeaconCandidate) {
			return NewBlockChainError(CandidateError, errors.New("beacon candidate list is INVALID"))
		}
	}
	if !reflect.DeepEqual(newShardCandidate, []string{}) {
		validShardCandidate := beaconView.GetValidStakers(newShardCandidate)
		if !reflect.DeepEqual(validShardCandidate, newShardCandidate) {
			return NewBlockChainError(CandidateError, errors.New("shard candidate list is INVALID"))
		}
	}
	//=============End Verify Stakers
	return nil
}

/* Verify Post-processing data
- Validator root: BeaconCommittee + BeaconPendingValidator
- Beacon Candidate root: CandidateBeaconWaitingForCurrentRandom + CandidateBeaconWaitingForNextRandom
- Shard Candidate root: CandidateShardWaitingForCurrentRandom + CandidateShardWaitingForNextRandom
- Shard Validator root: ShardCommittee + ShardPendingValidator
- Random number if have in instruction
*/
func (beaconView *BeaconView) verifyPostProcessingBeaconBlock(beaconBlock *BeaconBlock, randomClient btc.RandomClient) error {
	beaconView.lock.RLock()
	defer beaconView.lock.RUnlock()
	var (
		strs []string
	)

	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconView.BeaconCommittee)
	if err != nil {
		panic(err)
	}
	strs = append(strs, beaconCommitteeStr...)

	beaconPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(beaconView.BeaconPendingValidator)
	if err != nil {
		panic(err)
	}
	strs = append(strs, beaconPendingValidatorStr...)
	if hash, ok := verifyHashFromStringArray(strs, beaconBlock.Header.BeaconCommitteeAndValidatorRoot); !ok {
		return NewBlockChainError(BeaconCommitteeAndPendingValidatorRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v but get %+v", beaconBlock.Header.BeaconCommitteeAndValidatorRoot, hash))
	}
	strs = []string{}

	candidateBeaconWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(beaconView.CandidateBeaconWaitingForCurrentRandom)
	if err != nil {
		panic(err)
	}
	strs = append(strs, candidateBeaconWaitingForCurrentRandomStr...)

	candidateBeaconWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(beaconView.CandidateBeaconWaitingForNextRandom)
	if err != nil {
		panic(err)
	}
	strs = append(strs, candidateBeaconWaitingForNextRandomStr...)
	if hash, ok := verifyHashFromStringArray(strs, beaconBlock.Header.BeaconCandidateRoot); !ok {
		return NewBlockChainError(BeaconCandidateRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v but get %+v", beaconBlock.Header.BeaconCandidateRoot, hash))
	}
	strs = []string{}

	candidateShardWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(beaconView.CandidateShardWaitingForCurrentRandom)
	if err != nil {
		panic(err)
	}
	strs = append(strs, candidateShardWaitingForCurrentRandomStr...)

	candidateShardWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(beaconView.CandidateShardWaitingForNextRandom)
	if err != nil {
		panic(err)
	}
	strs = append(strs, candidateShardWaitingForNextRandomStr...)
	if hash, ok := verifyHashFromStringArray(strs, beaconBlock.Header.ShardCandidateRoot); !ok {
		return NewBlockChainError(ShardCandidateRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v but get %+v", beaconBlock.Header.ShardCandidateRoot, hash))
	}

	shardPendingValidator := make(map[byte][]string)
	for shardID, keyList := range beaconView.ShardPendingValidator {
		keyListStr, err := incognitokey.CommitteeKeyListToString(keyList)
		if err != nil {
			return err
		}
		shardPendingValidator[shardID] = keyListStr
	}

	shardCommittee := make(map[byte][]string)
	for shardID, keyList := range beaconView.ShardCommittee {
		keyListStr, err := incognitokey.CommitteeKeyListToString(keyList)
		if err != nil {
			return err
		}
		shardCommittee[shardID] = keyListStr
	}
	ok := verifyHashFromMapByteString(shardPendingValidator, shardCommittee, beaconBlock.Header.ShardCommitteeAndValidatorRoot)
	if !ok {
		return NewBlockChainError(ShardCommitteeAndPendingValidatorRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v", beaconBlock.Header.ShardCommitteeAndValidatorRoot))
	}
	if hash, ok := verifyHashFromMapStringBool(beaconView.AutoStaking, beaconBlock.Header.AutoStakingRoot); !ok {
		return NewBlockChainError(ShardCommitteeAndPendingValidatorRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v but get %+v", beaconBlock.Header.AutoStakingRoot, hash))
	}
	if !TestRandom {
		//COMMENT FOR TESTING
		instructions := beaconBlock.Body.Instructions
		for _, l := range instructions {
			if l[0] == "random" {
				startTime := time.Now()
				// ["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
				nonce, err := strconv.Atoi(l[1])
				if err != nil {
					Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
					return NewBlockChainError(UnExpectedError, err)
				}
				ok, err = randomClient.VerifyNonceWithTimestamp(startTime, beaconView.BlockMaxCreateTime, beaconView.CurrentRandomTimeStamp, int64(nonce))
				Logger.log.Infof("Verify Random number %+v", ok)
				if err != nil {
					Logger.log.Error("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
					return NewBlockChainError(UnExpectedError, err)
				}
				if !ok {
					return NewBlockChainError(RandomError, errors.New("Error verify random number"))
				}
			}
		}
	}
	return nil
}

/*
	Update FinalView with new Block
*/
func (beaconView *BeaconView) updateBeaconView(beaconBlock *BeaconBlock, chainParamEpoch uint64, chainParamAssignOffset int, randomTime uint64) error {
	beaconView.lock.Lock()
	defer beaconView.lock.Unlock()
	Logger.log.Debugf("Start processing new block at height %d, with hash %+v", beaconBlock.Header.Height, *beaconBlock.Hash())
	newBeaconCandidate := []incognitokey.CommitteePublicKey{}
	newShardCandidate := []incognitokey.CommitteePublicKey{}
	// Logger.log.Infof("Start processing new block at height %d, with hash %+v", newBlock.Header.Height, *newBlock.Hash())
	if beaconBlock == nil {
		return errors.New("null pointer")
	}
	// signal of random parameter from beacon block
	randomFlag := false
	// update BestShardHash, BestBlock, BestBlockHash
	beaconView.PreviousBestBlockHash = beaconView.BestBlockHash
	beaconView.BestBlockHash = *beaconBlock.Hash()
	beaconView.BestBlock = *beaconBlock
	beaconView.Epoch = beaconBlock.Header.Epoch
	beaconView.BeaconHeight = beaconBlock.Header.Height
	if beaconBlock.Header.Height == 1 {
		beaconView.BeaconProposerIndex = 0
	} else {
		beaconView.BeaconProposerIndex = (beaconView.BeaconProposerIndex + beaconBlock.Header.Round) % len(beaconView.BeaconCommittee)
	}
	if beaconView.BestShardHash == nil {
		beaconView.BestShardHash = make(map[byte]common.Hash)
	}
	if beaconView.BestShardHeight == nil {
		beaconView.BestShardHeight = make(map[byte]uint64)
	}
	// Update new best new block hash
	for shardID, shardStates := range beaconBlock.Body.ShardState {
		beaconView.BestShardHash[shardID] = shardStates[len(shardStates)-1].Hash
		beaconView.BestShardHeight[shardID] = shardStates[len(shardStates)-1].Height
	}
	// processing instruction
	for _, instruction := range beaconBlock.Body.Instructions {
		err, tempRandomFlag, tempNewBeaconCandidate, tempNewShardCandidate := beaconView.processInstruction(instruction)
		if err != nil {
			return err
		}
		if tempRandomFlag {
			randomFlag = tempRandomFlag
		}
		if len(tempNewBeaconCandidate) > 0 {
			newBeaconCandidate = append(newBeaconCandidate, tempNewBeaconCandidate...)
		}
		if len(tempNewShardCandidate) > 0 {
			newShardCandidate = append(newShardCandidate, tempNewShardCandidate...)
		}
	}
	// update candidate list after processing instructions
	beaconView.CandidateBeaconWaitingForNextRandom = append(beaconView.CandidateBeaconWaitingForNextRandom, newBeaconCandidate...)
	beaconView.CandidateShardWaitingForNextRandom = append(beaconView.CandidateShardWaitingForNextRandom, newShardCandidate...)

	if beaconView.BeaconHeight%chainParamEpoch == 1 && beaconView.BeaconHeight != 1 {
		// Begin of each epoch
		beaconView.IsGetRandomNumber = false
		// Before get random from bitcoin
	} else if beaconView.BeaconHeight%chainParamEpoch >= randomTime {
		// After get random from bitcoin
		if beaconView.BeaconHeight%chainParamEpoch == randomTime {
			// snapshot candidate list
			beaconView.CandidateShardWaitingForCurrentRandom = beaconView.CandidateShardWaitingForNextRandom
			beaconView.CandidateBeaconWaitingForCurrentRandom = beaconView.CandidateBeaconWaitingForNextRandom
			Logger.log.Info("Beacon Process: CandidateShardWaitingForCurrentRandom: ", beaconView.CandidateShardWaitingForCurrentRandom)
			Logger.log.Info("Beacon Process: CandidateBeaconWaitingForCurrentRandom: ", beaconView.CandidateBeaconWaitingForCurrentRandom)
			// reset candidate list
			beaconView.CandidateShardWaitingForNextRandom = []incognitokey.CommitteePublicKey{}
			beaconView.CandidateBeaconWaitingForNextRandom = []incognitokey.CommitteePublicKey{}
			// assign random timestamp
			beaconView.CurrentRandomTimeStamp = beaconBlock.Header.Timestamp
		}
		// if get new random number
		// Assign candidate to shard
		// assign CandidateShardWaitingForCurrentRandom to ShardPendingValidator with CurrentRandom
		if randomFlag {
			beaconView.IsGetRandomNumber = true
			numberOfPendingValidator := make(map[byte]int)
			for shardID, pendingValidators := range beaconView.ShardPendingValidator {
				numberOfPendingValidator[shardID] = len(pendingValidators)
			}
			shardCandidatesStr, err := incognitokey.CommitteeKeyListToString(beaconView.CandidateShardWaitingForCurrentRandom)
			if err != nil {
				panic(err)
			}
			remainShardCandidatesStr, assignedCandidates := assignShardCandidate(shardCandidatesStr, numberOfPendingValidator, beaconView.CurrentRandomNumber, chainParamAssignOffset, beaconView.ActiveShards)
			remainShardCandidates, err := incognitokey.CommitteeBase58KeyListToStruct(remainShardCandidatesStr)
			if err != nil {
				panic(err)
			}
			// append remain candidate into shard waiting for next random list
			beaconView.CandidateShardWaitingForNextRandom = append(beaconView.CandidateShardWaitingForNextRandom, remainShardCandidates...)
			// assign candidate into shard pending validator list
			for shardID, candidateListStr := range assignedCandidates {
				candidateList, err := incognitokey.CommitteeBase58KeyListToStruct(candidateListStr)
				if err != nil {
					panic(err)
				}
				beaconView.ShardPendingValidator[shardID] = append(beaconView.ShardPendingValidator[shardID], candidateList...)
			}
			// delete CandidateShardWaitingForCurrentRandom list
			beaconView.CandidateShardWaitingForCurrentRandom = []incognitokey.CommitteePublicKey{}
			// Shuffle candidate
			// shuffle CandidateBeaconWaitingForCurrentRandom with current random number
			newBeaconPendingValidator, err := ShuffleCandidate(beaconView.CandidateBeaconWaitingForCurrentRandom, beaconView.CurrentRandomNumber)
			if err != nil {
				return NewBlockChainError(ShuffleBeaconCandidateError, err)
			}
			beaconView.CandidateBeaconWaitingForCurrentRandom = []incognitokey.CommitteePublicKey{}
			beaconView.BeaconPendingValidator = append(beaconView.BeaconPendingValidator, newBeaconPendingValidator...)
		}
	}
	return nil
}

func (beaconView *BeaconView) initBeaconView(genesisBeaconBlock *BeaconBlock) error {
	var (
		newBeaconCandidate = []incognitokey.CommitteePublicKey{}
		newShardCandidate  = []incognitokey.CommitteePublicKey{}
	)
	Logger.log.Info("Process Update Beacon Best State With Beacon Genesis Block")
	beaconView.lock.Lock()
	defer beaconView.lock.Unlock()
	beaconView.PreviousBestBlockHash = beaconView.BestBlockHash
	beaconView.BestBlockHash = *genesisBeaconBlock.Hash()
	beaconView.BestBlock = *genesisBeaconBlock
	beaconView.Epoch = genesisBeaconBlock.Header.Epoch
	beaconView.BeaconHeight = genesisBeaconBlock.Header.Height
	beaconView.BeaconProposerIndex = 0
	beaconView.BestShardHash = make(map[byte]common.Hash)
	beaconView.BestShardHeight = make(map[byte]uint64)
	// Update new best new block hash
	for shardID, shardStates := range genesisBeaconBlock.Body.ShardState {
		beaconView.BestShardHash[shardID] = shardStates[len(shardStates)-1].Hash
		beaconView.BestShardHeight[shardID] = shardStates[len(shardStates)-1].Height
	}
	// update param
	for _, instruction := range genesisBeaconBlock.Body.Instructions {
		err, _, tempNewBeaconCandidate, tempNewShardCandidate := beaconView.processInstruction(instruction)
		if err != nil {
			return err
		}
		newBeaconCandidate = append(newBeaconCandidate, tempNewBeaconCandidate...)
		newShardCandidate = append(newShardCandidate, tempNewShardCandidate...)
	}
	beaconView.BeaconCommittee = append(beaconView.BeaconCommittee, newBeaconCandidate...)
	beaconView.ConsensusAlgorithm = common.BlsConsensus
	beaconView.ShardConsensusAlgorithm = make(map[byte]string)
	for shardID := 0; shardID < beaconView.ActiveShards; shardID++ {
		beaconView.ShardCommittee[byte(shardID)] = append(beaconView.ShardCommittee[byte(shardID)], newShardCandidate[shardID*beaconView.MinShardCommitteeSize:(shardID+1)*beaconView.MinShardCommitteeSize]...)
		beaconView.ShardConsensusAlgorithm[byte(shardID)] = common.BlsConsensus
	}
	beaconView.Epoch = 1
	beaconView.NumOfBlocksByProducers = make(map[string]uint64)
	return nil
}

/*
	processInstruction, process these instruction:
	- Random Instruction
		+ format
			["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
		+ store random number into FinalView
	- Swap Instruction
		+ format
			["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
			["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
        + Update shard/beacon pending validator and shard/beacon committee in FinalView
	- Stake Instruction
		+ format
			["stake", "pubkey1,pubkey2,..." "shard" "txStake1,txStake2,..." "rewardReceiver1,rewardReceiver2,..." flag]
			["stake", "pubkey1,pubkey2,..." "beacon" "txStake1,txStake2,..." "rewardReceiver1,rewardReceiver2,..." flag]
		+ Get Stake public key and for later storage
	Return param
	#1 error
	#2 random flag
	#3 new beacon candidate
	#4 new shard candidate

*/
func (beaconView *BeaconView) processInstruction(instruction []string) (error, bool, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey) {
	newBeaconCandidates := []incognitokey.CommitteePublicKey{}
	newShardCandidates := []incognitokey.CommitteePublicKey{}
	if len(instruction) < 1 {
		return nil, false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
	}
	// ["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
	if instruction[0] == RandomAction {
		temp, err := strconv.Atoi(instruction[1])
		if err != nil {
			return NewBlockChainError(ProcessRandomInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		}
		beaconView.CurrentRandomNumber = int64(temp)
		Logger.log.Infof("Random number found %+v", beaconView.CurrentRandomNumber)
		return nil, true, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
	}
	if instruction[0] == StopAutoStake {
		committeePublicKeys := strings.Split(instruction[1], ",")
		for _, committeePublicKey := range committeePublicKeys {
			allCommitteeValidatorCandidate := beaconView.getAllCommitteeValidatorCandidateFlattenList()
			// check existence in all committee list
			if common.IndexOfStr(committeePublicKey, allCommitteeValidatorCandidate) == -1 {
				// if not found then delete auto staking data for this public key if present
				if _, ok := beaconView.AutoStaking[committeePublicKey]; ok {
					delete(beaconView.AutoStaking, committeePublicKey)
				}
			} else {
				// if found in committee list then turn off auto staking
				if _, ok := beaconView.AutoStaking[committeePublicKey]; ok {
					beaconView.AutoStaking[committeePublicKey] = false
				}
			}
		}
	}
	if instruction[0] == SwapAction {
		Logger.log.Info("Swap Instruction", instruction)
		inPublickeys := strings.Split(instruction[1], ",")
		Logger.log.Info("Swap Instruction In Public Keys", inPublickeys)
		inPublickeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(inPublickeys)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		}
		outPublickeys := strings.Split(instruction[2], ",")
		Logger.log.Info("Swap Instruction Out Public Keys", outPublickeys)
		outPublickeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(outPublickeys)
		if err != nil {
			if len(outPublickeys) != 0 {
				return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
			}
		}

		if instruction[3] == "shard" {
			temp, err := strconv.Atoi(instruction[4])
			if err != nil {
				return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
			}
			shardID := byte(temp)
			// delete in public key out of sharding pending validator list
			if len(instruction[1]) > 0 {
				shardPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(beaconView.ShardPendingValidator[shardID])
				if err != nil {
					return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				tempShardPendingValidator, err := RemoveValidator(shardPendingValidatorStr, inPublickeys)
				if err != nil {
					return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				// update shard pending validator
				beaconView.ShardPendingValidator[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(tempShardPendingValidator)
				if err != nil {
					return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				// add new public key to committees
				beaconView.ShardCommittee[shardID] = append(beaconView.ShardCommittee[shardID], inPublickeyStructs...)
			}
			// delete out public key out of current committees
			if len(instruction[2]) > 0 {
				//for _, value := range outPublickeyStructs {
				//	delete(beaconView.RewardReceiver, value.GetIncKeyBase58())
				//}
				shardCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconView.ShardCommittee[shardID])
				if err != nil {
					return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				tempShardCommittees, err := RemoveValidator(shardCommitteeStr, outPublickeys)
				if err != nil {
					return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				// remove old public key in shard committee update shard committee
				beaconView.ShardCommittee[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(tempShardCommittees)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				// Check auto stake in out public keys list
				// if auto staking not found or flag auto stake is false then do not re-stake for this out public key
				// if auto staking flag is true then system will automatically add this out public key to current candidate list
				for index, outPublicKey := range outPublickeys {
					if isAutoRestaking, ok := beaconView.AutoStaking[outPublicKey]; !ok {
						if _, ok := beaconView.RewardReceiver[outPublicKey]; ok {
							delete(beaconView.RewardReceiver, outPublickeyStructs[index].GetIncKeyBase58())
						}
						continue
					} else {
						if !isAutoRestaking {
							// delete this flag for next time staking
							delete(beaconView.RewardReceiver, outPublickeyStructs[index].GetIncKeyBase58())
							delete(beaconView.AutoStaking, outPublicKey)
						} else {
							shardCandidate, err := incognitokey.CommitteeBase58KeyListToStruct([]string{outPublicKey})
							if err != nil {
								return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
							}
							newShardCandidates = append(newShardCandidates, shardCandidate...)
						}
					}
				}
			}
		} else if instruction[3] == "beacon" {
			if len(instruction[1]) > 0 {
				beaconPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(beaconView.BeaconPendingValidator)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				tempBeaconPendingValidator, err := RemoveValidator(beaconPendingValidatorStr, inPublickeys)
				if err != nil {
					return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				// update beacon pending validator
				beaconView.BeaconPendingValidator, err = incognitokey.CommitteeBase58KeyListToStruct(tempBeaconPendingValidator)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				// add new public key to beacon committee
				beaconView.BeaconCommittee = append(beaconView.BeaconCommittee, inPublickeyStructs...)
			}
			if len(instruction[2]) > 0 {
				// delete reward receiver
				//for _, value := range outPublickeyStructs {
				//	delete(beaconView.RewardReceiver, value.GetIncKeyBase58())
				//}
				beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconView.BeaconCommittee)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				tempBeaconCommittes, err := RemoveValidator(beaconCommitteeStr, outPublickeys)
				if err != nil {
					return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				// remove old public key in beacon committee and update beacon best state
				beaconView.BeaconCommittee, err = incognitokey.CommitteeBase58KeyListToStruct(tempBeaconCommittes)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				for index, outPublicKey := range outPublickeys {
					if isAutoRestaking, ok := beaconView.AutoStaking[outPublicKey]; !ok {
						if _, ok := beaconView.RewardReceiver[outPublicKey]; ok {
							delete(beaconView.RewardReceiver, outPublickeyStructs[index].GetIncKeyBase58())
						}
						continue
					} else {
						if !isAutoRestaking {
							delete(beaconView.RewardReceiver, outPublickeyStructs[index].GetIncKeyBase58())
							delete(beaconView.AutoStaking, outPublicKey)
						} else {
							beaconCandidate, err := incognitokey.CommitteeBase58KeyListToStruct([]string{outPublicKey})
							if err != nil {
								return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
							}
							newBeaconCandidates = append(newBeaconCandidates, beaconCandidate...)
						}
					}
				}
			}
		}
		return nil, false, newBeaconCandidates, newShardCandidates
	}
	// Update candidate
	// get staking candidate list and store
	// store new staking candidate
	if instruction[0] == StakeAction && instruction[2] == "beacon" {
		beaconCandidates := strings.Split(instruction[1], ",")
		beaconCandidatesStructs, err := incognitokey.CommitteeBase58KeyListToStruct(beaconCandidates)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		}
		beaconRewardReceivers := strings.Split(instruction[4], ",")
		beaconAutoReStaking := strings.Split(instruction[5], ",")
		if len(beaconCandidatesStructs) != len(beaconRewardReceivers) && len(beaconRewardReceivers) != len(beaconAutoReStaking) {
			return NewBlockChainError(StakeInstructionError, fmt.Errorf("Expect Beacon Candidate (length %+v) and Beacon Reward Receiver (length %+v) and Beacon Auto ReStaking (lenght %+v) have equal length", len(beaconCandidates), len(beaconRewardReceivers), len(beaconAutoReStaking))), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		}
		for index, candidate := range beaconCandidatesStructs {
			beaconView.RewardReceiver[candidate.GetIncKeyBase58()] = beaconRewardReceivers[index]
			if beaconAutoReStaking[index] == "true" {
				beaconView.AutoStaking[beaconCandidates[index]] = true
			} else {
				beaconView.AutoStaking[beaconCandidates[index]] = false
			}
		}

		newBeaconCandidates = append(newBeaconCandidates, beaconCandidatesStructs...)
		return nil, false, newBeaconCandidates, newShardCandidates
	}
	if instruction[0] == StakeAction && instruction[2] == "shard" {
		shardCandidates := strings.Split(instruction[1], ",")
		shardCandidatesStructs, err := incognitokey.CommitteeBase58KeyListToStruct(shardCandidates)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		}
		shardRewardReceivers := strings.Split(instruction[4], ",")
		shardAutoReStaking := strings.Split(instruction[5], ",")
		if len(shardCandidates) != len(shardRewardReceivers) && len(shardRewardReceivers) != len(shardAutoReStaking) {
			return NewBlockChainError(StakeInstructionError, fmt.Errorf("Expect Beacon Candidate (length %+v) and Beacon Reward Receiver (length %+v) and Shard Auto ReStaking (length %+v) have equal length", len(shardCandidates), len(shardRewardReceivers), len(shardAutoReStaking))), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		}
		for index, candidate := range shardCandidatesStructs {
			beaconView.RewardReceiver[candidate.GetIncKeyBase58()] = shardRewardReceivers[index]
			if shardAutoReStaking[index] == "true" {
				beaconView.AutoStaking[shardCandidates[index]] = true
			} else {
				beaconView.AutoStaking[shardCandidates[index]] = false
			}
		}
		newShardCandidates = append(newShardCandidates, shardCandidatesStructs...)
		return nil, false, newBeaconCandidates, newShardCandidates
	}
	return nil, false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
}

func (blockchain *BlockChain) processStoreBeaconBlock(
	beaconBlock *BeaconBlock,
	snapshotBeaconCommittees []incognitokey.CommitteePublicKey,
	snapshotAllShardCommittees map[byte][]incognitokey.CommitteePublicKey,
	snapshotRewardReceivers map[string]string,
) error {

	// Logger.log.Debugf("BEACON | Process Store Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, beaconBlock.Header.Hash())
	// blockHash := beaconBlock.Header.Hash()
	// for shardID, shardStates := range beaconBlock.Body.ShardState {
	// 	for _, shardState := range shardStates {
	// 		err := blockchain.config.DataBase.StoreAcceptedShardToBeacon(shardID, beaconBlock.Header.Height, shardState.Hash)
	// 		if err != nil {
	// 			return NewBlockChainError(StoreAcceptedShardToBeaconError, err)
	// 		}
	// 	}
	// }
	// Logger.log.Infof("BEACON | Store Committee in Beacon Block Height %+v ", beaconBlock.Header.Height)
	// if err := blockchain.config.DataBase.StoreShardCommitteeByHeight(beaconBlock.Header.Height, snapshotAllShardCommittees); err != nil {
	// 	return NewBlockChainError(StoreShardCommitteeByHeightError, err)
	// }
	// if err := blockchain.config.DataBase.StoreBeaconCommitteeByHeight(beaconBlock.Header.Height, snapshotBeaconCommittees); err != nil {
	// 	return NewBlockChainError(StoreBeaconCommitteeByHeightError, err)
	// }
	// if err := blockchain.config.DataBase.StoreRewardReceiverByHeight(beaconBlock.Header.Height, snapshotRewardReceivers); err != nil {
	// 	return NewBlockChainError(StoreRewardReceiverByHeightError, err)
	// }
	// if err := blockchain.config.DataBase.StoreAutoStakingByHeight(beaconBlock.Header.Height, blockchain.FinalView.Beacon.AutoStaking); err != nil {
	// 	return NewBlockChainError(StoreAutoStakingByHeightError, err)
	// }
	// //================================Store cross shard state ==================================
	// if beaconBlock.Body.ShardState != nil {
	// 	blockchain.FinalView.Beacon.lock.Lock()
	// 	lastCrossShardState := blockchain.FinalView.Beacon.LastCrossShardState
	// 	for fromShard, shardBlocks := range beaconBlock.Body.ShardState {
	// 		for _, shardBlock := range shardBlocks {
	// 			for _, toShard := range shardBlock.CrossShard {
	// 				if fromShard == toShard {
	// 					continue
	// 				}
	// 				if lastCrossShardState[fromShard] == nil {
	// 					lastCrossShardState[fromShard] = make(map[byte]uint64)
	// 				}
	// 				lastHeight := lastCrossShardState[fromShard][toShard] // get last cross shard height from shardID  to crossShardShardID
	// 				waitHeight := shardBlock.Height
	// 				err := blockchain.config.DataBase.StoreCrossShardNextHeight(fromShard, toShard, lastHeight, waitHeight)
	// 				if err != nil {
	// 					blockchain.FinalView.Beacon.lock.Unlock()
	// 					return NewBlockChainError(StoreCrossShardNextHeightError, err)
	// 				}
	// 				//beacon process shard_to_beacon in order so cross shard next height also will be saved in order
	// 				//dont care overwrite this value
	// 				err = blockchain.config.DataBase.StoreCrossShardNextHeight(fromShard, toShard, waitHeight, 0)
	// 				if err != nil {
	// 					blockchain.FinalView.Beacon.lock.Unlock()
	// 					return NewBlockChainError(StoreCrossShardNextHeightError, err)
	// 				}
	// 				if lastCrossShardState[fromShard] == nil {
	// 					lastCrossShardState[fromShard] = make(map[byte]uint64)
	// 				}
	// 				lastCrossShardState[fromShard][toShard] = waitHeight //update lastHeight to waitHeight
	// 			}
	// 		}
	// 		blockchain.config.CrossShardPool[fromShard].UpdatePool()
	// 	}
	// 	blockchain.FinalView.Beacon.lock.Unlock()
	// }
	// //=============================END Store cross shard state ==================================
	// if err := blockchain.config.DataBase.StoreBeaconBlockIndex(blockHash, beaconBlock.Header.Height); err != nil {
	// 	return NewBlockChainError(StoreBeaconBlockIndexError, err)
	// }

	// batchPutData := []database.BatchData{}

	// Logger.log.Debugf("Store Beacon FinalView Height %+v", beaconBlock.Header.Height)
	// if err := blockchain.StoreBeaconBestState(&batchPutData); err != nil {
	// 	return NewBlockChainError(StoreBeaconBestStateError, err)
	// }
	// Logger.log.Debugf("Store Beacon Block Height %+v with Hash %+v ", beaconBlock.Header.Height, blockHash)
	// if err := blockchain.config.DataBase.StoreBeaconBlock(beaconBlock, blockHash, &batchPutData); err != nil {
	// 	return NewBlockChainError(StoreBeaconBlockError, err)
	// }

	// err := blockchain.updateDatabaseWithBlockRewardInfo(beaconBlock, &batchPutData)
	// if err != nil {
	// 	return NewBlockChainError(UpdateDatabaseWithBlockRewardInfoError, err)
	// }
	// // execute, store
	// err = blockchain.processBridgeInstructions(beaconBlock, &batchPutData)
	// if err != nil {
	// 	return NewBlockChainError(ProcessBridgeInstructionError, err)
	// }

	// // execute, store
	// err = blockchain.processPDEInstructions(beaconBlock, &batchPutData)
	// if err != nil {
	// 	return NewBlockChainError(ProcessPDEInstructionError, err)
	// }

	// return blockchain.config.DataBase.PutBatch(batchPutData)
	return nil
}
