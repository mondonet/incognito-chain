package blsbftv2

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/wire"
)

const (
	MSG_PROPOSE    = "propose"
	MSG_VOTE       = "vote"
	MSG_REQUESTBLK = "getblk"
)

type BFTPropose struct {
	Block  json.RawMessage
	PeerID string
}

type BFTVote struct {
	ViewHash  string
	BlockHash string
	Validator string
	BLS       []byte
	BRI       []byte
	VoteSig   []byte
}

type BFTRequestBlock struct {
	BlockHash string
	PeerID    string
}

func MakeBFTProposeMsg(block []byte, chainKey string, userKeySet *MiningKey) (wire.Message, error) {
	var proposeCtn BFTPropose
	proposeCtn.Block = block
	proposeCtnBytes, err := json.Marshal(proposeCtn)
	if err != nil {
		return nil, consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = proposeCtnBytes
	msg.(*wire.MessageBFT).Type = MSG_PROPOSE
	return msg, nil
}

func MakeBFTVoteMsg(vote *BFTVote, chainKey string) (wire.Message, error) {
	voteCtnBytes, err := json.Marshal(vote)
	if err != nil {
		return nil, consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = voteCtnBytes
	msg.(*wire.MessageBFT).Type = MSG_VOTE
	return msg, nil
}

func MakeBFTRequestBlk(request BFTRequestBlock, peerID string, chainKey string) (wire.Message, error) {
	requestCtnBytes, err := json.Marshal(request)
	if err != nil {
		return nil, consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = requestCtnBytes
	msg.(*wire.MessageBFT).Type = MSG_REQUESTBLK
	return msg, nil
}

func (e *BLSBFT) processProposeMsg(proposeMsg *BFTPropose) error {
	// proposer only propose once
	// voter can vote on multi-views

	block, err := e.Chain.UnmarshalBlock(proposeMsg.Block)
	if err != nil {
		return err
	}
	blockHash := block.Hash().String()
	view, err := e.validatePreSignBlock(block)
	if err != nil {
		return err
	}

	switch err {
	case UnexpectedError:
		return err
	case BlockAlreadyFinalizedError:
		return e.Node.PushBlockToPeer(view.GetTipBlock(), proposeMsg.PeerID)
	case AlreadyReceivedProposedBlockError:
		e.lockOnGoingBlocks.RLock()
		defer e.lockOnGoingBlocks.RUnlock()
		instance := e.onGoingBlocks[blockHash]
		err := instance.sendOwnVote()
		if err != nil {
			return err
		}
	case BlockIsMissingFromBlockConsensusInstanceError:
		e.lockOnGoingBlocks.RLock()
		defer e.lockOnGoingBlocks.RUnlock()
		instance := e.onGoingBlocks[blockHash]
		err := instance.addBlock(block)
		if err != nil {
			return err
		}
	case nil:
		instance, err := e.createBlockConsensusInstance(view, blockHash)
		if err != nil {
			return err
		}
		e.lockOnGoingBlocks.RLock()
		instance.addBlock(block)
		if bestBlockHash, ok := e.bestProposeBlockOfView[block.GetPreviousViewHash().String()]; ok {
			if block.GetTimeslot() < e.onGoingBlocks[bestBlockHash].Timeslot {
				e.bestProposeBlockOfView[block.GetPreviousViewHash().String()] = blockHash
			}
		}
		e.onGoingBlocks[blockHash].createAndSendVote()
		e.lockOnGoingBlocks.RUnlock()
	}

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
		go e.Node.PushMessageToPeer(msg, requestMsg.PeerID)
	}
	return nil
}

func (e *BLSBFT) ProcessBFTMsg(msg consensus.ConsensusMsgInterface) {
	msgBFT := msg.(*wire.MessageBFT)
	switch msgBFT.Type {
	case MSG_PROPOSE:
		var msgPropose BFTPropose
		err := json.Unmarshal(msgBFT.Content, &msgPropose)
		if err != nil {
			e.Logger.Error(err)
			return
		}
		go e.processProposeMsg(&msgPropose)
	case MSG_VOTE:
		var msgVote BFTVote
		err := json.Unmarshal(msgBFT.Content, &msgVote)
		if err != nil {
			e.Logger.Error(err)
			return
		}
		go e.processVoteMsg(&msgVote)
	case MSG_REQUESTBLK:
		var msgRequest BFTRequestBlock
		err := json.Unmarshal(msgBFT.Content, &msgRequest)
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
