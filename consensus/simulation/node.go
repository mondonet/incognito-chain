package main

import (
	"fmt"
	"os"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/chain"
	"github.com/incognitochain/incognito-chain/consensus"

	"github.com/incognitochain/incognito-chain/common"
	blsbft "github.com/incognitochain/incognito-chain/consensus/blsbft2"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

type Node struct {
	id              string
	consensusEngine *blsbft.BLSBFT
	chain           consensus.ChainManagerInterface
	nodeList        []*Node
}

var committeePkStruct int

type logWriter struct {
	NodeID string
	fd     *os.File
}

func (s logWriter) Write(p []byte) (n int, err error) {
	s.fd.Write(p)
	return len(p), nil
}

func NewNode(committeePkStruct []incognitokey.CommitteePublicKey, committee []string, index int) *Node {
	name := fmt.Sprintf("node_%d", index)
	node := Node{id: fmt.Sprintf("%d", index)}

	node.chain = chain.InitNewChain("shard0", &blockchain.ShardView{})
	//node.chain.UserPubKey = committeePkStruct[index]

	fd, err := os.OpenFile(fmt.Sprintf("%s.log", name), os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	fd.Truncate(0)
	backendLog := common.NewBackend(logWriter{
		NodeID: name,
		fd:     fd,
	})
	logger := backendLog.Logger("Consensus", false)

	node.consensusEngine = &blsbft.BLSBFT{
		Chain:    node.chain,
		Node:     &node,
		ChainKey: "shard",
		PeerID:   name,
		Logger:   logger,
	}

	prvSeed, err := blsbft.LoadUserKeyFromIncPrivateKey(committee[index])
	failOnError(err)
	failOnError(node.consensusEngine.LoadUserKey(prvSeed))
	return &node
}

func (s *Node) Start() {
	s.consensusEngine.Start()
}

func (s *Node) PushMessageToChain(msg interface{}, chain consensus.ChainManagerInterface) error {
	if msg.(*wire.MessageBFT).Type == "propose" {
		timeSlot := GetTimeSlot(msg.(*wire.MessageBFT).Timestamp)
		pComm := GetSimulation().scenario.proposeComm
		if comm, ok := pComm[timeSlot]; ok {
			for i, c := range s.nodeList {
				if senderComm, ok := comm[s.id]; ok {
					if senderComm[i] == 1 {
						c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
					}
				} else {
					c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
				}

			}
		} else {
			for _, c := range s.nodeList {
				c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
			}
		}
		return nil
	}

	if msg.(*wire.MessageBFT).Type == "vote" {
		vComm := GetSimulation().scenario.voteComm
		timeSlot := GetTimeSlot(msg.(*wire.MessageBFT).Timestamp)
		if comm, ok := vComm[timeSlot]; ok {
			for i, c := range s.nodeList {
				if senderComm, ok := comm[s.id]; ok {
					if senderComm[i] == 1 {
						c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
					}
				} else {
					c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
				}

			}
		} else {
			for _, c := range s.nodeList {
				c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
			}
		}
		return nil
	}
	panic("implement me")
}

func (Node) UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string) {
	//not use in bft
	return
}

func (Node) IsEnableMining() bool {
	//not use in bft
	return true
}

func (Node) GetMiningKeys() string {
	//not use in bft
	panic("implement me")
}

func (Node) GetPrivateKey() string {
	//not use in bft
	panic("implement me")
}

func (Node) DropAllConnections() {
	//not use in bft
	return
}

func (Node) PushMessageToPeer(msg interface{}, peerId libp2p.ID) error {
	return nil
}

func main() {

}
