package main

import (
	"errors"

	//"github.com/incognitochain/incognito-chain/metrics"

	"time"

	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"

	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

// OnBlock is invoked when a peer receives a block message.  It
// blocks until the coin block has been fully processed.
func (serverObj *Server) OnBlockShard(p *peer.PeerConn,
	msg *wire.MessageBlockShard) {
	Logger.log.Debug("Receive a new blockshard START")

	var txProcessed chan struct{}
	serverObj.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new blockshard END")
}

func (serverObj *Server) OnBlockBeacon(p *peer.PeerConn,
	msg *wire.MessageBlockBeacon) {
	Logger.log.Debug("Receive a new blockbeacon START")

	var txProcessed chan struct{}
	serverObj.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new blockbeacon END")
}

func (serverObj *Server) OnCrossShard(p *peer.PeerConn,
	msg *wire.MessageCrossShard) {
	Logger.log.Debug("Receive a new crossshard START")

	var txProcessed chan struct{}
	serverObj.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new crossshard END")
}

func (serverObj *Server) OnShardToBeacon(p *peer.PeerConn,
	msg *wire.MessageShardToBeacon) {
	Logger.log.Debug("Receive a new shardToBeacon START")

	var txProcessed chan struct{}
	serverObj.netSync.QueueBlock(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new shardToBeacon END")
}

func (serverObj *Server) OnGetBlockBeacon(_ *peer.PeerConn, msg *wire.MessageGetBlockBeacon) {
	Logger.log.Debug("Receive a " + msg.MessageType() + " message START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueGetBlockBeacon(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a " + msg.MessageType() + " message END")
}
func (serverObj *Server) OnGetBlockShard(_ *peer.PeerConn, msg *wire.MessageGetBlockShard) {
	Logger.log.Debug("Receive a " + msg.MessageType() + " message START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueGetBlockShard(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a " + msg.MessageType() + " message END")
}

func (serverObj *Server) OnGetCrossShard(_ *peer.PeerConn, msg *wire.MessageGetCrossShard) {
	Logger.log.Debug("Receive a getcrossshard START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Debug("Receive a getcrossshard END")
}

func (serverObj *Server) OnGetShardToBeacon(_ *peer.PeerConn, msg *wire.MessageGetShardToBeacon) {
	Logger.log.Debug("Receive a getshardtobeacon START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Debug("Receive a getshardtobeacon END")
}

// OnTx is invoked when a peer receives a tx message.  It blocks
// until the transaction has been fully processed.  Unlock the block
// handler this does not serialize all transactions through a single thread
// transactions don't rely on the previous one in a linear fashion like blocks.
func (serverObj *Server) OnTx(peer *peer.PeerConn, msg *wire.MessageTx) {
	Logger.log.Debug("Receive a new transaction START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueTx(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new transaction END")
}

func (serverObj *Server) OnTxToken(peer *peer.PeerConn, msg *wire.MessageTxToken) {
	Logger.log.Debug("Receive a new transaction(normal token) START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueTxToken(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new transaction(normal token) END")
}

func (serverObj *Server) OnTxPrivacyToken(peer *peer.PeerConn, msg *wire.MessageTxPrivacyToken) {
	Logger.log.Debug("Receive a new transaction(privacy token) START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueTxPrivacyToken(nil, msg, txProcessed)
	//<-txProcessed

	Logger.log.Debug("Receive a new transaction(privacy token) END")
}

/*
// OnVersion is invoked when a peer receives a version message
// and is used to negotiate the protocol version details as well as kick start
// the communications.
*/
func (serverObj *Server) OnVersion(peerConn *peer.PeerConn, msg *wire.MessageVersion) {
	Logger.log.Debug("Receive version message START")

	pbk := ""
	pbkType := ""
	if msg.PublicKey != "" {
		//TODO hy set publickey here
		//fmt.Printf("Message %v %v %v\n", msg.SignDataB58, msg.PublicKey, msg.PublicKeyType)
		err := serverObj.consensusEngine.VerifyData([]byte(peerConn.GetListenerPeer().GetPeerID().Pretty()), msg.SignDataB58, msg.PublicKey, msg.PublicKeyType)
		//fmt.Println(err)

		if err == nil {
			pbk = msg.PublicKey
			pbkType = msg.PublicKeyType
		} else {
			peerConn.ForceClose()
			return
		}
	}

	peerConn.GetRemotePeer().SetPublicKey(pbk, pbkType)

	remotePeer := &peer.Peer{}
	remotePeer.SetListeningAddress(msg.LocalAddress)
	remotePeer.SetPeerID(msg.LocalPeerId)
	remotePeer.SetRawAddress(msg.RawLocalAddress)
	remotePeer.SetPublicKey(pbk, pbkType)
	serverObj.cNewPeers <- remotePeer

	if msg.ProtocolVersion != serverObj.protocolVersion {
		Logger.log.Error(errors.New("Not correct version "))
		peerConn.ForceClose()
		return
	}

	// check for accept connection
	if accepted, e := serverObj.connManager.CheckForAcceptConn(peerConn); !accepted {
		// not accept connection -> force close
		Logger.log.Error(e)
		peerConn.ForceClose()
		return
	}

	msgV, err := wire.MakeEmptyMessage(wire.CmdVerack)
	if err != nil {
		return
	}

	msgV.(*wire.MessageVerAck).Valid = true
	msgV.(*wire.MessageVerAck).Timestamp = time.Now()

	peerConn.QueueMessageWithEncoding(msgV, nil, peer.MessageToPeer, nil)

	//	push version message again
	if !peerConn.VerAckReceived() {
		err := serverObj.PushVersionMessage(peerConn)
		if err != nil {
			Logger.log.Error(err)
		}
	}

	Logger.log.Debug("Receive version message END")
}

/*
OnVerAck is invoked when a peer receives a version acknowlege message
*/
func (serverObj *Server) OnVerAck(peerConn *peer.PeerConn, msg *wire.MessageVerAck) {
	Logger.log.Debug("Receive verack message START")

	if msg.Valid {
		peerConn.SetVerValid(true)

		if peerConn.GetIsOutbound() {
			serverObj.addrManager.Good(peerConn.GetRemotePeer())
		}

		// send message for get addr
		//msgSG, err := wire.MakeEmptyMessage(wire.CmdGetAddr)
		//if err != nil {
		//	return
		//}
		//var dc chan<- struct{}
		//peerConn.QueueMessageWithEncoding(msgSG, dc, peer.MessageToPeer, nil)

		//	broadcast addr to all peer
		//listen := serverObj.connManager.GetListeningPeer()
		//msgSA, err := wire.MakeEmptyMessage(wire.CmdAddr)
		//if err != nil {
		//	return
		//}
		//
		//rawPeers := []wire.RawPeer{}
		//peers := serverObj.addrManager.AddressCache()
		//for _, peer := range peers {
		//	getPeerId, _ := serverObj.connManager.GetPeerId(peer.GetRawAddress())
		//	if peerConn.GetRemotePeerID().Pretty() != getPeerId {
		//		pk, pkT := peer.GetPublicKey()
		//		rawPeers = append(rawPeers, wire.RawPeer{peer.GetRawAddress(), pkT, pk})
		//	}
		//}
		//msgSA.(*wire.MessageAddr).RawPeers = rawPeers
		//var doneChan chan<- struct{}
		//listen.GetPeerConnsMtx().Lock()
		//for _, peerConn := range listen.GetPeerConns() {
		//	Logger.log.Debug("QueueMessageWithEncoding", peerConn)
		//	peerConn.QueueMessageWithEncoding(msgSA, doneChan, peer.MessageToPeer, nil)
		//}
		//listen.GetPeerConnsMtx().Unlock()
	} else {
		peerConn.SetVerValid(true)
	}

	Logger.log.Debug("Receive verack message END")
}

func (serverObj *Server) OnGetAddr(peerConn *peer.PeerConn, msg *wire.MessageGetAddr) {
	Logger.log.Debug("Receive getaddr message START")

	// send message for addr
	msgS, err := wire.MakeEmptyMessage(wire.CmdAddr)
	if err != nil {
		return
	}

	peers := serverObj.addrManager.AddressCache()
	rawPeers := []wire.RawPeer{}
	for _, peer := range peers {
		getPeerId, _ := serverObj.connManager.GetPeerId(peer.GetRawAddress())
		if peerConn.GetRemotePeerID().Pretty() != getPeerId {
			pk, pkT := peer.GetPublicKey()
			rawPeers = append(rawPeers, wire.RawPeer{peer.GetRawAddress(), pkT, pk})
		}
	}

	msgS.(*wire.MessageAddr).RawPeers = rawPeers
	var dc chan<- struct{}
	peerConn.QueueMessageWithEncoding(msgS, dc, peer.MessageToPeer, nil)

	Logger.log.Debug("Receive getaddr message END")
}

func (serverObj *Server) OnAddr(peerConn *peer.PeerConn, msg *wire.MessageAddr) {
	Logger.log.Debugf("Receive addr message %v", msg.RawPeers)
}

func (serverObj *Server) OnBFTMsg(p *peer.PeerConn, msg wire.Message) {
	Logger.log.Debug("Receive a BFTMsg START")
	var txProcessed chan struct{}
	isRelayNodeForConsensus := cfg.Accelerator
	if isRelayNodeForConsensus {
		senderPublicKey, _ := p.GetRemotePeer().GetPublicKey()
		// panic(senderPublicKey)
		// fmt.Println("eiiiiiiiiiiiii")
		// os.Exit(0)
		//TODO hy check here

		bestState := serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().(*blockchain.BeaconView)
		beaconCommitteeList, err := incognitokey.CommitteeKeyListToString(bestState.BeaconCommittee)
		if err != nil {
			panic(err)
		}
		isInBeaconCommittee := common.IndexOfStr(senderPublicKey, beaconCommitteeList) != -1
		if isInBeaconCommittee {
			serverObj.PushMessageToBeacon(msg, map[libp2p.ID]bool{p.GetRemotePeerID(): true})
		}
		shardCommitteeList := make(map[byte][]string)
		for shardID, committee := range bestState.GetShardCommittee() {
			shardCommitteeList[shardID], err = incognitokey.CommitteeKeyListToString(committee)
			if err != nil {
				panic(err)
			}
		}
		for shardID, committees := range shardCommitteeList {
			isInShardCommitee := common.IndexOfStr(senderPublicKey, committees) != -1
			if isInShardCommitee {
				serverObj.PushMessageToShard(msg, shardID, map[libp2p.ID]bool{p.GetRemotePeerID(): true})
				break
			}
		}
	}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Debug("Receive a BFTMsg END")
}

func (serverObj *Server) OnPeerState(_ *peer.PeerConn, msg *wire.MessagePeerState) {
	Logger.log.Debug("Receive a peerstate START")
	var txProcessed chan struct{}
	serverObj.netSync.QueueMessage(nil, msg, txProcessed)
	Logger.log.Debug("Receive a peerstate END")
}
