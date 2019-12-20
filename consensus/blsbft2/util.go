package blsbftv2

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
)

// func (e *BLSBFT) getTimeSinceLastBlock() time.Duration {
// 	return time.Since(time.Unix(int64(e.Chain.GetTimeStamp()), 0))
// }

// func (e *BLSBFT) waitForNextTimeslot() bool {
// 	timeSinceLastBlk := e.getTimeSinceLastBlock()
// 	if timeSinceLastBlk >= e.Chain.GetMinBlkInterval() {
// 		return false
// 	} else {
// 		//fmt.Println("\n\nWait for", e.Chain.GetMinBlkInterval()-timeSinceLastBlk, "\n\n")
// 		return true
// 	}
// }

func (e *BLSBFT) ExtractBridgeValidationData(block common.BlockInterface) ([][]byte, []int, error) {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return nil, nil, consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	return valData.BridgeSig, valData.ValidatiorsIdx, nil
}

func (e *BLSBFT) proposeBlock(timeslot uint64) error {
	bestView := e.Chain.GetBestView()
	bestViewHash := bestView.Hash().String()
	e.lockOnGoingBlocks.RLock()
	bestProposedBlockHash, ok := e.bestProposeBlockOfView[bestViewHash]
	e.lockOnGoingBlocks.RUnlock()

	if ok {
		//re-broadcast best proposed block
		e.lockOnGoingBlocks.RLock()
		blockData, _ := json.Marshal(e.onGoingBlocks[bestProposedBlockHash].Block)
		e.lockOnGoingBlocks.RUnlock()
		msg, _ := MakeBFTProposeMsg(blockData, e.ChainKey, e.UserKeySet)
		go e.Node.PushMessageToChain(msg, e.Chain)
		e.onGoingBlocks[bestProposedBlockHash].createAndSendVote()
	} else {
		//create block and boardcast block
		if isProducer(timeslot, bestView.GetCommittee(), e.UserKeySet.GetPublicKeyBase58()) != nil {
			return errors.New("I'm not the block producer")
		}
		block, err := bestView.CreateNewBlock(timeslot)
		if err != nil {
			return err
		}
		validationData := e.CreateValidationData(block)
		validationDataString, _ := EncodeValidationData(validationData)
		block.(blockValidation).AddValidationField(validationDataString)

		blockHash := block.Hash().String()

		instance, err := e.createBlockConsensusInstance(bestView, blockHash)
		if err != nil {
			return err
		}

		err = instance.addBlock(block)
		if err != nil {
			return err
		}

		e.bestProposeBlockOfView[bestViewHash] = blockHash
		e.proposedBlockOnView.BlockHash = blockHash
		e.proposedBlockOnView.ViewHash = bestViewHash
		blockData, _ := json.Marshal(block)
		msg, _ := MakeBFTProposeMsg(blockData, e.ChainKey, e.UserKeySet)
		go e.Node.PushMessageToChain(msg, e.Chain)
	}

	return nil
}

func (vote *BFTVote) signVote(signFunc func(data []byte) ([]byte, error)) error {
	data := []byte(vote.BlockHash)
	data = append(data, vote.BLS...)
	data = append(data, vote.BRI...)
	data = common.HashB(data)
	var err error
	vote.VoteSig, err = signFunc(data)
	return err
}

func getTimeSlot(genesisTime int64, pointInTime int64, slotTime int64) uint64 {
	slotTimeDur := time.Duration(slotTime)
	blockTime := time.Unix(pointInTime, 0)
	timePassed := blockTime.Sub(time.Unix(genesisTime, 0)).Round(slotTimeDur)
	timeSlot := uint64(int64(timePassed.Seconds()) / slotTime)
	return timeSlot
}

func parseConsensusConfig(config string) (*consensusConfig, error) {
	var result consensusConfig
	err := json.Unmarshal([]byte(config), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
