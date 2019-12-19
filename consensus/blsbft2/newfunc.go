package blsbftv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/chain"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wire"

	peer2 "github.com/libp2p/go-libp2p-core/peer"
)

func (e BLSBFT) preValidateCheck(block *common.BlockInterface) bool {
	// blockViewHash := block.GetPreviousViewHash()

	return true
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

func (e BLSBFT) processProposeMsg(proposeMsg *BFTPropose) error {
	// proposer only propose once
	// voter can vote on multi-views

	block, err := e.Chain.UnmarshalBlock(proposeMsg.Block)
	if err != nil {
		return err
	}
	currentViewTimeslot := e.currentTimeslotOfViews[block.GetPreviousViewHash().String()]

	if block.GetTimeslot() > currentViewTimeslot {
		// hmm... something wrong with local clock?
		return fmt.Errorf("this propose block has timeslot higher than current timeslot. BLOCK:%v CURRENT:%v", block.GetTimeslot(), currentViewTimeslot)
	}
	blockHash := block.Hash().String()
	if blockHash == e.Chain.GetBestView().GetTipBlock().Hash().String() {
		//send this block
	}
	e.lockOnGoingBlocks.RLock()
	if _, ok := e.onGoingBlocks[blockHash]; ok {
		if e.onGoingBlocks[blockHash].Block != nil {
			e.lockOnGoingBlocks.RUnlock()
			return errors.New("already received this propose block")
		}
	}
	e.lockOnGoingBlocks.RUnlock()

	view, err := e.Chain.GetViewByHash(block.GetPreviousViewHash())
	if err != nil {
		if block.GetHeight() > e.Chain.GetBestView().GetHeight() {
			//request block
			return nil
		}
		return err
	}

	consensusCfg, err := parseConsensusConfig(view.GetConsensusConfig())
	if err != nil {
		return err
	}
	consensusSlottime, err := time.ParseDuration(consensusCfg.Slottime)
	if err != nil {
		return err
	}
	// if view.GetHeight() == e.Chain.GetBestView().GetHeight() {
	if err := e.validateProducer(block, view, int64(consensusSlottime.Seconds()), view.GetCommittee(), e.Logger); err != nil {
		return err
	}
	if len(e.onGoingBlocks) > 0 {
		e.lockOnGoingBlocks.RLock()
		if bestBlockHash, ok := e.bestProposeBlockOfView[block.GetPreviousViewHash().String()]; ok {
			if block.GetTimeslot() < e.onGoingBlocks[bestBlockHash].Timeslot {
				e.lockOnGoingBlocks.RUnlock()
				instance, err := e.createBlockConsensusInstance(view, blockHash)
				if err != nil {
					return err
				}
				e.lockOnGoingBlocks.RLock()
				instance.addBlock(block)
				e.bestProposeBlockOfView[block.GetPreviousViewHash().String()] = blockHash
				e.onGoingBlocks[bestBlockHash].createAndSendVote()
				e.lockOnGoingBlocks.RUnlock()
			}
		} else {
			defer e.lockOnGoingBlocks.RUnlock()
			instance := e.onGoingBlocks[blockHash]
			err := instance.addBlock(block)
			if err != nil {
				return err
			}
		}
	} else {
		err := view.ValidateBlock(block, true)
		if err != nil {
			return err
		}
		instance, err := e.createBlockConsensusInstance(view, blockHash)
		if err != nil {
			return err
		}
		e.lockOnGoingBlocks.RLock()
		instance.addBlock(block)
		e.bestProposeBlockOfView[block.GetPreviousViewHash().String()] = blockHash
		e.onGoingBlocks[blockHash].createAndSendVote()
		e.lockOnGoingBlocks.RUnlock()
	}
	// }
	return nil
}

func (e *BLSBFT) processVoteMsg(vote *BFTVote) error {
	e.lockOnGoingBlocks.RLock()
	viewHash, err := common.Hash{}.NewHashFromStr(vote.ViewHash)
	if err != nil {
		return err
	}
	view, err := e.Chain.GetViewByHash(viewHash)
	if err != nil {
		return err
	}
	var instance *blockConsensusInstance
	if _, ok := e.onGoingBlocks[vote.BlockHash]; !ok {
		e.lockOnGoingBlocks.RUnlock()
		instance, err = e.createBlockConsensusInstance(view, vote.BlockHash)
		if err != nil {
			return err
		}
	}
	if instance == nil {
		instance = e.onGoingBlocks[vote.BlockHash]
	}

	if err := instance.addVote(vote); err != nil {
		return err
	}
	voteMsg, err := MakeBFTVoteMsg(vote, e.ChainKey)
	if err != nil {
		return err
	}
	e.Node.PushMessageToChain(voteMsg, e.Chain)
	return nil
}

func (e *BLSBFT) processRequestBlkMsg(requestMsg *BFTRequestBlock) error {
	e.lockOnGoingBlocks.RLock()
	defer e.lockOnGoingBlocks.RUnlock()
	block, ok := e.onGoingBlocks[requestMsg.BlockHash]
	if ok {
		blockData, err := json.Marshal(block)
		if err != nil {
			return err
		}
		msg, err := MakeBFTProposeMsg(blockData, e.ChainKey, e.UserKeySet)
		if err != nil {
			return err
		}
		peerID, err := peer2.IDB58Decode(requestMsg.PeerID)
		if err != nil {
			return err
		}
		go e.Node.PushMessageToPeer(msg, peerID)
	}
	return nil
}

func (e *BLSBFT) ProcessBFTMsg(msg *wire.MessageBFT) {
	switch msg.Type {
	case MSG_PROPOSE:
		var msgPropose BFTPropose
		err := json.Unmarshal(msg.Content, &msgPropose)
		if err != nil {
			e.Logger.Error(err)
			return
		}
		go e.processProposeMsg(&msgPropose)
	case MSG_VOTE:
		var msgVote BFTVote
		err := json.Unmarshal(msg.Content, &msgVote)
		if err != nil {
			e.Logger.Error(err)
			return
		}
		go e.processVoteMsg(&msgVote)
	case MSG_REQUESTBLK:
		var msgRequest BFTRequestBlock
		err := json.Unmarshal(msg.Content, &msgRequest)
		if err != nil {
			e.Logger.Error(err)
			return
		}
		go e.processRequestBlkMsg(&msgRequest)
	default:
		e.Logger.Critical("Unknown BFT message type")
		return
	}
}

func (blockCI *blockConsensusInstance) addVote(vote *BFTVote) error {
	blockCI.lockVote.Lock()
	defer blockCI.lockVote.Unlock()
	if _, ok := blockCI.Votes[vote.Validator]; !ok {
		return errors.New("already received this vote")
	}
	err := validateVote(vote)
	if err != nil {
		return err
	}
	blockCI.Votes[vote.Validator] = vote
	return nil
}

func (blockCI *blockConsensusInstance) confirmVote(blockHash *common.Hash, vote *BFTVote) error {
	data := blockHash.GetBytes()
	data = append(data, vote.BLS...)
	data = append(data, vote.BRI...)
	data = common.HashB(data)
	var err error
	vote.VoteSig, err = blockCI.Engine.UserKeySet.BriSignData(data)
	return err
}

func (blockCI *blockConsensusInstance) createAndSendVote() error {
	var vote BFTVote

	pubKey := blockCI.Engine.UserKeySet.GetPublicKey()
	selfIdx := common.IndexOfStr(pubKey.GetMiningKeyBase58(consensusName), blockCI.Committee.StringList)

	blsSig, err := blockCI.Engine.UserKeySet.BLSSignData(blockCI.Block.Hash().GetBytes(), selfIdx, blockCI.Committee.ByteList)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	bridgeSig := []byte{}
	if metadata.HasBridgeInstructions(blockCI.Block.GetInstructions()) {
		bridgeSig, err = blockCI.Engine.UserKeySet.BriSignData(blockCI.Block.Hash().GetBytes())
		if err != nil {
			return consensus.NewConsensusError(consensus.UnExpectedError, err)
		}
	}

	vote.BLS = blsSig
	vote.BRI = bridgeSig
	vote.Validator = pubKey.GetMiningKeyBase58(consensusName)

	msg, err := MakeBFTVoteMsg(&vote, blockCI.Engine.ChainKey)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}

	blockCI.Votes[pubKey.GetMiningKeyBase58(consensusName)] = &vote
	blockCI.Engine.Logger.Info("sending vote...")
	go blockCI.Engine.Node.PushMessageToChain(msg, blockCI.Engine.Chain)
	return nil
}

func validateProposeBlock(block common.BlockInterface, view chain.ChainViewInterface) (BFTVote, error) {
	err := view.ValidateBlock(block, true)
	if err != nil {
		return BFTVote{}, err
	}
	var v BFTVote

	return v, nil
}

func (blockCI *blockConsensusInstance) initInstance(view chain.ChainViewInterface) error {
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

func validateProducerPosition(block common.BlockInterface, genesisTime int64, slotTime int64, committee []incognitokey.CommitteePublicKey) error {
	timeSlot := getTimeSlot(genesisTime, block.GetBlockTimestamp(), slotTime)
	if block.GetTimeslot() != timeSlot {
		return consensus.NewConsensusError(consensus.InvalidTimeslotError, fmt.Errorf("Timeslot should be %v instead of %v", timeSlot, block.GetTimeslot()))
	}
	return isProducer(timeSlot, committee, block.GetProducer())
}

func getProducerPosition(timeslot uint64, committeeLen uint64) uint64 {
	return timeslot % committeeLen
}

func isProducer(timeslot uint64, committee []incognitokey.CommitteePublicKey, producerPbk string) error {
	producerPosition := getProducerPosition(timeslot, uint64(len(committee)))
	tempProducer, err := committee[producerPosition].ToBase58()
	if err != nil {
		return err
	}
	if tempProducer != producerPbk {
		return consensus.NewConsensusError(consensus.UnExpectedError, fmt.Errorf("Producer should be should be %v instead of %v", tempProducer, producerPbk))
	}
	return nil
}

func (blockCI *blockConsensusInstance) addBlock(block common.BlockInterface) error {
	blockCI.Block = block
	blockCI.Timeslot = block.GetTimeslot()
	blockCI.Phase = votePhase
	return nil
}

func (e *BLSBFT) createBlockConsensusInstance(view chain.ChainViewInterface, blockHash string) (*blockConsensusInstance, error) {
	e.lockOnGoingBlocks.Lock()
	defer e.lockOnGoingBlocks.Unlock()
	var blockCI blockConsensusInstance
	blockCI.View = view
	blockCI.Phase = listenPhase

	var cfg consensusConfig
	err := json.Unmarshal([]byte(view.GetConsensusConfig()), &cfg)
	if err != nil {
		return nil, err
	}
	blockCI.ConsensusCfg = cfg
	cmHash := view.GetCommitteeHash()
	cmCache, ok := e.viewCommitteesCache.Get(cmHash.String())
	if !ok {
		committee := view.GetCommittee()
		var cmDecode committeeDecode
		cmDecode.Committee = committee
		cmDecode.ByteList = []blsmultisig.PublicKey{}
		cmDecode.StringList = []string{}
		for _, member := range cmDecode.Committee {
			cmDecode.ByteList = append(cmDecode.ByteList, member.MiningPubKey[consensusName])
		}
		committeeBLSString, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(cmDecode.Committee, consensusName)
		if err != nil {
			return nil, err
		}
		cmDecode.StringList = committeeBLSString
		e.viewCommitteesCache.Add(view.GetCommitteeHash().String(), cmDecode, committeeCacheCleanupTime)
		blockCI.Committee = cmDecode
	} else {
		blockCI.Committee = cmCache.(committeeDecode)
	}

	e.onGoingBlocks[blockHash] = &blockCI
	return &blockCI, nil
}

func (blockCI *blockConsensusInstance) FinalizeBlock() error {
	aggSig, brigSigs, validatorIdx, err := combineVotes(blockCI.Votes, blockCI.Committee.StringList)
	if err != nil {
		return err
	}

	blockCI.ValidationData.AggSig = aggSig
	blockCI.ValidationData.BridgeSig = brigSigs
	blockCI.ValidationData.ValidatiorsIdx = validatorIdx

	validationDataString, _ := EncodeValidationData(blockCI.ValidationData)
	blockCI.Block.(blockValidation).AddValidationField(validationDataString)

	//TODO 0xakk0r0kamui trace who is malicious node if ValidateCommitteeSig return false
	err = validateCommitteeSig(blockCI.Block, blockCI.Committee.Committee)
	if err != nil {
		return err
	}
	view, err := blockCI.View.ConnectBlockAndCreateView(blockCI.Block)
	if err != nil {
		//if blockchainError, ok := err.(*blockchain.BlockChainError); ok {
		//	if blockchainError.Code == blockchain.ErrCodeMessage[blockchain.DuplicateShardBlockError].Code {
		//		return nil
		//	}
		//}
		return err
	}
	err = blockCI.Engine.Chain.AddView(view)
	if err != nil {
		return err
	}

	return nil
}
