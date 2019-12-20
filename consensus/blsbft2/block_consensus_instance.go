package blsbftv2

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
)

type blockConsensusInstance struct {
	Engine         *BLSBFT
	View           consensus.ChainViewInterface
	ConsensusCfg   consensusConfig
	Block          common.BlockInterface
	ValidationData ValidationData
	Votes          map[string]*BFTVote
	lockVote       sync.RWMutex
	Timeslot       uint64
	Phase          string
	Committee      committeeDecode
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

func (blockCI *blockConsensusInstance) addBlock(block common.BlockInterface) error {
	blockCI.Block = block
	blockCI.Timeslot = block.GetTimeslot()
	blockCI.Phase = votePhase
	return nil
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
		//if blockchainError, ok := err.(*blockconsensus.BlockChainError); ok {
		//	if blockchainError.Code == blockconsensus.ErrCodeMessage[blockconsensus.DuplicateShardBlockError].Code {
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

func (blockCI *blockConsensusInstance) sendOwnVote() error {
	pubKey := blockCI.Engine.UserKeySet.GetPublicKey()
	vote := blockCI.Votes[pubKey.GetMiningKeyBase58(consensusName)]

	msg, err := MakeBFTVoteMsg(vote, blockCI.Engine.ChainKey)
	if err != nil {
		return err
	}

	blockCI.Engine.Logger.Info("sending vote...")
	go blockCI.Engine.Node.PushMessageToChain(msg, blockCI.Engine.Chain)
	return nil
}

func (e *BLSBFT) createBlockConsensusInstance(view consensus.ChainViewInterface, blockHash string) (*blockConsensusInstance, error) {
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
