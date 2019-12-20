package blsbftv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type blockValidation interface {
	common.BlockInterface
	AddValidationField(validationData string) error
}

type ValidationData struct {
	ProducerBLSSig []byte
	ProducerBriSig []byte
	ValidatiorsIdx []int
	AggSig         []byte
	BridgeSig      [][]byte
}

func DecodeValidationData(data string) (*ValidationData, error) {
	var valData ValidationData
	err := json.Unmarshal([]byte(data), &valData)
	if err != nil {
		return nil, consensus.NewConsensusError(consensus.DecodeValidationDataError, err)
	}
	return &valData, nil
}

func EncodeValidationData(validationData ValidationData) (string, error) {
	result, err := json.Marshal(validationData)
	if err != nil {
		return "", consensus.NewConsensusError(consensus.EncodeValidationDataError, err)
	}
	return string(result), nil
}

func (e BLSBFT) CreateValidationData(block common.BlockInterface) ValidationData {
	var valData ValidationData
	valData.ProducerBLSSig, _ = e.UserKeySet.BriSignData(block.Hash().GetBytes())
	return valData
}

func (e BLSBFT) validateProducer(block common.BlockInterface, view consensus.ChainViewInterface, slotTime int64, committee []incognitokey.CommitteePublicKey, log common.Logger) error {
	log.Info("ValidateProducerPosition...")
	if err := validateProducerPosition(block, view.GetGenesisTime(), slotTime, committee); err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	log.Info("ValidateProducerSig...")
	if err := validateProducerSig(block); err != nil {
		return consensus.NewConsensusError(consensus.ProducerSignatureError, err)
	}
	return nil
}

func validateProducerSig(block common.BlockInterface) error {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}

	producerKey := incognitokey.CommitteePublicKey{}
	err = producerKey.FromBase58(block.GetProducer())
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	//start := time.Now()
	if err := validateSingleBriSig(block.Hash(), valData.ProducerBLSSig, producerKey.MiningPubKey[common.BridgeConsensus]); err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	//end := time.Now().Sub(start)
	//fmt.Printf("ConsLog just verify %v\n", end.Seconds())
	return nil
}

func (e BLSBFT) ValidateProducerSig(block common.BlockInterface) error {
	return validateProducerSig(block)
}
func validateCommitteeSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	committeeBLSKeys := []blsmultisig.PublicKey{}
	for _, member := range committee {
		committeeBLSKeys = append(committeeBLSKeys, member.MiningPubKey[consensusName])
	}
	if err := validateBLSSig(block.Hash(), valData.AggSig, valData.ValidatiorsIdx, committeeBLSKeys); err != nil {
		fmt.Println(block.Hash(), block.GetValidationField())
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	return nil
}

func (e BLSBFT) ValidateCommitteeSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	return validateCommitteeSig(block, committee)
}

func (e BLSBFT) ValidateData(data []byte, sig string, publicKey string) error {
	sigByte, _, err := base58.Base58Check{}.Decode(sig)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	publicKeyByte := []byte(publicKey)
	// if err != nil {
	// 	return consensus.NewConsensusError(consensus.UnExpectedError, err)
	// }
	//fmt.Printf("ValidateData data %v, sig %v, publicKey %v\n", data, sig, publicKeyByte)
	dataHash := new(common.Hash)
	dataHash.NewHash(data)
	_, err = bridgesig.Verify(publicKeyByte, dataHash.GetBytes(), sigByte) //blsmultisig.Verify(sigByte, data, []int{0}, []blsmultisig.PublicKey{publicKeyByte})
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	return nil
}

func validateSingleBLSSig(
	dataHash *common.Hash,
	blsSig []byte,
	selfIdx int,
	committee []blsmultisig.PublicKey,
) error {
	//start := time.Now()
	result, err := blsmultisig.Verify(blsSig, dataHash.GetBytes(), []int{selfIdx}, committee)
	//end := time.Now().Sub(start)
	//fmt.Printf("ConsLog single verify %v\n", end.Seconds())
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	if !result {
		return consensus.NewConsensusError(consensus.UnExpectedError, errors.New("invalid BLS Signature"))
	}
	return nil
}

func validateSingleBriSig(
	dataHash *common.Hash,
	briSig []byte,
	validatorPk []byte,
) error {
	result, err := bridgesig.Verify(validatorPk, dataHash.GetBytes(), briSig)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	if !result {
		return consensus.NewConsensusError(consensus.UnExpectedError, errors.New("invalid BRI Signature"))
	}
	return nil
}

func validateBLSSig(
	dataHash *common.Hash,
	aggSig []byte,
	validatorIdx []int,
	committee []blsmultisig.PublicKey,
) error {
	result, err := blsmultisig.Verify(aggSig, dataHash.GetBytes(), validatorIdx, committee)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	if !result {
		return consensus.NewConsensusError(consensus.UnExpectedError, errors.New("Invalid Signature!"))
	}
	return nil
}

func validateVote(Vote *BFTVote) error {
	data := []byte{}
	blkHash, err := common.Hash{}.NewHashFromStr(Vote.BlockHash)
	if err != nil {
		return err
	}
	data = append(data, blkHash.GetBytes()...)
	data = append(data, Vote.BLS...)
	data = append(data, Vote.BRI...)
	dataHash := common.HashH(data)

	publicKeyByte := []byte(Vote.Validator)
	err = validateSingleBriSig(&dataHash, Vote.VoteSig, publicKeyByte)
	return err
}

func (e BLSBFT) ValidateBlockWithConsensus(block common.BlockInterface) error {

	return nil
}

func validateBlockWithView(block common.BlockInterface, view consensus.ChainViewInterface) (BFTVote, error) {
	err := view.ValidateBlock(block, true)
	if err != nil {
		return BFTVote{}, err
	}
	var v BFTVote

	return v, nil
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

func (e *BLSBFT) validatePreSignBlock(block common.BlockInterface) (consensus.ChainViewInterface, error) {
	currentViewTimeslot := e.currentTimeslotOfViews[block.GetPreviousViewHash().String()]

	if block.GetTimeslot() > currentViewTimeslot {
		// hmm... something wrong with local clock?
		return nil, fmt.Errorf("this propose block has timeslot higher than current timeslot. BLOCK:%v CURRENT:%v", block.GetTimeslot(), currentViewTimeslot)
	}
	blockHash := block.Hash().String()
	if blockHash == e.Chain.GetBestView().GetTipBlock().Hash().String() {
		//send this block
	}
	e.lockOnGoingBlocks.RLock()
	if _, ok := e.onGoingBlocks[blockHash]; ok {
		if e.onGoingBlocks[blockHash].Block != nil {
			e.lockOnGoingBlocks.RUnlock()
			return nil, errors.New("already received this propose block")
		}
	}
	e.lockOnGoingBlocks.RUnlock()

	view, err := e.Chain.GetViewByHash(block.GetPreviousViewHash())
	if err != nil {
		if block.GetHeight() > e.Chain.GetBestView().GetHeight() {
			//request block
			return nil, nil
		}
		return nil, err
	}

	consensusCfg, err := parseConsensusConfig(view.GetConsensusConfig())
	if err != nil {
		return nil, err
	}
	consensusSlottime, err := time.ParseDuration(consensusCfg.Slottime)
	if err != nil {
		return nil, err
	}
	// if view.GetHeight() == e.Chain.GetBestView().GetHeight() {
	if err := e.validateProducer(block, view, int64(consensusSlottime.Seconds()), view.GetCommittee(), e.Logger); err != nil {
		return nil, err
	}
	err = view.ValidateBlock(block, true)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
