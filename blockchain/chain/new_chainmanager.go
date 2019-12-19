package chain

import (
	"bytes"
	"encoding/gob"
	"errors"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

/*
	How to load Chain Manager when bootup
	Logic to store chainview
	How to manage Chain View
*/

type ChainManager struct {
	manager *ViewGraph
	name    string
	lock    *sync.RWMutex
}

func (s *ChainManager) ConnectBlockAndAddView(block common.BlockInterface) error {
	panic("implement me")
}

//to get byte array of this chain data, for loading later
func (s *ChainManager) GetChainData() (res []byte) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

//create new chain with root manager
func InitNewChain(name string, rootView consensus.ChainViewInterface) *ChainManager {
	cm := &ChainManager{
		name: name,
		lock: new(sync.RWMutex),
	}
	cm.manager = NewViewGraph(name, rootView, cm.lock)
	return cm
}

func LoadChain(data []byte) (res *ChainManager) {
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(res)
	if err != nil {
		panic(err)
	}
	return res
}

func (s *ChainManager) AddView(view consensus.ChainViewInterface) error {
	s.manager.AddView(view)
	return nil
}

func (ChainManager) GetChainName() string {
	panic("implement me")
}

func (ChainManager) IsReady() bool {
	panic("implement me")
}

func (cManager ChainManager) GetActiveShardNumber() int {
	return 0
}

func (ChainManager) GetPubkeyRole(pubkey string, round int) (string, byte) {
	panic("implement me")
}

func (ChainManager) GetConsensusType() string {
	panic("implement me")
}

func (ChainManager) GetTimeStamp() int64 {
	panic("implement me")
}

func (ChainManager) GetMinBlkInterval() time.Duration {
	panic("implement me")
}

func (ChainManager) GetMaxBlkCreateTime() time.Duration {
	panic("implement me")
}

func (ChainManager) GetHeight() uint64 {
	panic("implement me")
}

func (ChainManager) GetCommitteeSize() int {
	panic("implement me")
}

func (ChainManager) GetCommittee() []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (ChainManager) GetPubKeyCommitteeIndex(string) int {
	panic("implement me")
}

func (ChainManager) GetLastProposerIndex() int {
	panic("implement me")
}

func (ChainManager) UnmarshalBlock(blockString []byte) (common.BlockInterface, error) {
	panic("implement me")
}

func (ChainManager) ValidateBlockSignatures(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	panic("implement me")
}

func (s *ChainManager) GetShardID() int {
	panic("implement me")
}

func (s ChainManager) GetBestView() consensus.ChainViewInterface {
	return s.manager.GetBestView()
}

func (s ChainManager) GetFinalView() consensus.ChainViewInterface {
	return s.manager.GetFinalView()
}

func (ChainManager) GetAllViews() map[string]consensus.ChainViewInterface {
	panic("implement me")
}

func (s *ChainManager) GetViewByHash(h *common.Hash) (consensus.ChainViewInterface, error) {
	viewnode, ok := s.manager.node[*h]
	if !ok {
		return nil, errors.New("view not exist")
	} else {
		return viewnode.view, nil
	}
}

func (ChainManager) GetGenesisTime() int64 {
	panic("implement me")
}

func (s *ChainManager) GetAllTipBlocksHash() []*common.Hash {
	var result []*common.Hash
	for _, node := range s.manager.node {
		result = append(result, node.view.GetTipBlock().Hash())
	}
	return result
}
