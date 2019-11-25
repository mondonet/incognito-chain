package blsbftv2

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wire"
)

func (e BLSBFT) preValidateCheck(block *common.BlockInterface) bool {

	return true
}

func (e BLSBFT) getProposeBlock() (common.BlockInterface, error) {
	if e.bestProposeBlock == "" {
		block, err := e.Chain.GetBestView().CreateNewBlock(e.currentTimeslot)
		if err != nil {
			return nil, err
		}
		return block, nil
	}
	_ = e.Chain.GetBestView().GetTimeslot()

	return nil, nil

}

func (e *BLSBFT) processProposeMsg(proposeMsg *BFTPropose) error {
	return nil
}

func (e *BLSBFT) processVoteMsg(voteMsg *BFTVote) error {
	return nil
}

func (e *BLSBFT) ProcessBFTMsg(msg *wire.MessageBFT) {
	switch msg.Type {
	case MSG_PROPOSE:
		var msgPropose BFTPropose
		err := json.Unmarshal(msg.Content, &msgPropose)
		if err != nil {
			fmt.Println(err)
			return
		}
		go e.processProposeMsg(&msgPropose)
	case MSG_VOTE:
		var msgVote BFTVote
		err := json.Unmarshal(msg.Content, &msgVote)
		if err != nil {
			fmt.Println(err)
			return
		}
		e.processVoteMsg(&msgVote)
	default:
		e.logger.Critical("???")
		return
	}
}

func (e *BLSBFT) isInTimeslot() bool {
	return false
}

func (e *BLSBFT) enterNewTimeslot() error {

	return nil
}

func (blockCss *blockConsensusInstance) addVote(vote *BFTVote) error {
	blockCss.lockVote.Lock()
	defer blockCss.lockVote.Unlock()
	return nil
}

func (blockCss *blockConsensusInstance) confirmVote(vote *BFTVote) error {
	return nil
}

func (blockCss *blockConsensusInstance) createAndSendVote() (BFTVote, error) {
	var vote BFTVote

	pubKey := blockCss.Engine.UserKeySet.GetPublicKey()
	selfIdx := common.IndexOfStr(pubKey.GetMiningKeyBase58(consensusName), blockCss.CommitteeBLS.StringList)

	blsSig, err := blockCss.Engine.UserKeySet.BLSSignData(blockCss.Block.Hash().GetBytes(), selfIdx, blockCss.CommitteeBLS.ByteList)
	if err != nil {
		return vote, consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	bridgeSig := []byte{}
	if metadata.HasBridgeInstructions(blockCss.Block.GetInstructions()) {
		bridgeSig, err = blockCss.Engine.UserKeySet.BriSignData(blockCss.Block.Hash().GetBytes())
		if err != nil {
			return vote, consensus.NewConsensusError(consensus.UnExpectedError, err)
		}
	}

	vote.BLS = blsSig
	vote.BRI = bridgeSig
	vote.Validator = pubKey.GetMiningKeyBase58(consensusName)

	msg, err := MakeBFTVoteMsg(vote, blockCss.Engine.ChainKey)
	if err != nil {
		return vote, consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	blockCss.Votes[pubKey.GetMiningKeyBase58(consensusName)] = vote
	blockCss.Engine.logger.Info("sending vote...")
	go blockCss.Engine.Node.PushMessageToChain(msg, blockCss.Engine.Chain)
	// e.RoundData.NotYetSendVote = false

	return vote, nil
}

func (blockCss *blockConsensusInstance) processProposeBlock() (BFTVote, error) {
	var v BFTVote
	return v, nil
}

func (blockCss *blockConsensusInstance) initInstance(view blockchain.ChainViewInterface) error {
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