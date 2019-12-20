package main

import (
	"errors"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

func (serverObj *Server) RequestSyncBlockByHash(blockHash *common.Hash, isUnknownView bool, tipBlocksHash []common.Hash, peerID libp2p.ID) error {
	return nil
}

func (serverObj *Server) PushMessageToChain(msg interface{}, chain consensus.ChainManagerInterface) error {
	chainID := chain.GetShardID()
	if chainID == -1 {
		serverObj.PushMessageToBeacon(msg.(wire.Message), map[libp2p.ID]bool{})
	} else {
		serverObj.PushMessageToShard(msg.(wire.Message), byte(chainID), map[libp2p.ID]bool{})
	}
	return nil
}

func (serverObj *Server) PushBlockToAll(block common.BlockInterface, isBeacon bool) error {
	if isBeacon {
		msg, err := wire.MakeEmptyMessage(wire.CmdBlockBeacon)
		if err != nil {
			Logger.log.Error(err)
			return err
		}
		msg.(*wire.MessageBlockBeacon).Block = block.(*blockchain.BeaconBlock)
		serverObj.PushMessageToAll(msg)
		return nil
	} else {
		shardBlock := block.(*blockchain.ShardBlock)
		msgShard, err := wire.MakeEmptyMessage(wire.CmdBlockShard)
		if err != nil {
			Logger.log.Error(err)
			return err
		}
		msgShard.(*wire.MessageBlockShard).Block = shardBlock
		serverObj.PushMessageToShard(msgShard, shardBlock.Header.ShardID, map[libp2p.ID]bool{})

		shardToBeaconBlk := shardBlock.CreateShardToBeaconBlock(serverObj.blockChain)
		msgShardToBeacon, err := wire.MakeEmptyMessage(wire.CmdBlkShardToBeacon)
		if err != nil {
			Logger.log.Error(err)
			return err
		}
		msgShardToBeacon.(*wire.MessageShardToBeacon).Block = shardToBeaconBlk
		serverObj.PushMessageToBeacon(msgShardToBeacon, map[libp2p.ID]bool{})

		crossShardBlks := shardBlock.CreateAllCrossShardBlock(serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().GetActiveShardNumber())
		for shardID, crossShardBlk := range crossShardBlks {
			msgCrossShardShard, err := wire.MakeEmptyMessage(wire.CmdCrossShard)
			if err != nil {
				Logger.log.Error(err)
				return err
			}
			msgCrossShardShard.(*wire.MessageCrossShard).Block = crossShardBlk
			serverObj.PushMessageToShard(msgCrossShardShard, shardID, map[libp2p.ID]bool{})
		}
	}
	return nil
}

func (serverObj *Server) PublishNodeState(userLayer string, shardID int) error {
	Logger.log.Infof("[peerstate] Start Publish SelfPeerState")
	listener := serverObj.connManager.GetConfig().ListenerPeer

	// if (userRole != common.CommitteeRole) && (userRole != common.ValidatorRole) && (userRole != common.ProposerRole) {
	// 	return errors.New("Not in committee, don't need to publish node state!")
	// }

	// userKey, _ := serverObj.consensusEngine.GetCurrentMiningPublicKey()
	// metrics.SetGlobalParam("MINING_PUBKEY", userKey)
	msg, err := wire.MakeEmptyMessage(wire.CmdPeerState)
	if err != nil {
		return err
	}
	msg.(*wire.MessagePeerState).Beacon = blockchain.ChainState{
		serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().(*blockchain.BeaconView).BestBlock.Header.Timestamp,
		serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().(*blockchain.BeaconView).BeaconHeight,
		serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().(*blockchain.BeaconView).BestBlockHash,
		serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().(*blockchain.BeaconView).Hash(),
	}

	if userLayer != common.BeaconRole {
		msg.(*wire.MessagePeerState).Shards[byte(shardID)] = blockchain.ChainState{
			serverObj.blockChain.Chains[common.GetShardChainKey(byte(shardID))].GetBestView().(*blockchain.ShardView).TipBlock.Header.Timestamp,
			serverObj.blockChain.Chains[common.GetShardChainKey(byte(shardID))].GetBestView().(*blockchain.ShardView).GetHeight(),
			*serverObj.blockChain.Chains[common.GetShardChainKey(byte(shardID))].GetBestView().(*blockchain.ShardView).TipBlock.Hash(),
			serverObj.blockChain.Chains[common.GetShardChainKey(byte(shardID))].GetBestView().(*blockchain.ShardView).Hash(),
		}
	} else {
		msg.(*wire.MessagePeerState).ShardToBeaconPool = serverObj.shardToBeaconPool.GetValidBlockHeight()
		Logger.log.Infof("[peerstate] %v", msg.(*wire.MessagePeerState).ShardToBeaconPool)
	}

	if (cfg.NodeMode == common.NodeModeAuto || cfg.NodeMode == common.NodeModeShard) && shardID >= 0 {
		msg.(*wire.MessagePeerState).CrossShardPool[byte(shardID)] = serverObj.crossShardPool[byte(shardID)].GetValidBlockHeight()
	}

	//
	currentMiningKey := serverObj.consensusEngine.GetMiningPublicKeys()[serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().GetConsensusType()]
	msg.(*wire.MessagePeerState).SenderMiningPublicKey, err = currentMiningKey.ToBase58()
	if err != nil {
		return err
	}
	msg.SetSenderID(listener.GetPeerID())
	Logger.log.Infof("[peerstate] PeerID send to Proxy when publish node state %v \n", listener.GetPeerID())
	if err != nil {
		return err
	}
	Logger.log.Debugf("Publish peerstate")
	serverObj.PushMessageToAll(msg)
	return nil
}

func (serverObj *Server) BoardcastNodeState() error {
	listener := serverObj.connManager.GetConfig().ListenerPeer
	msg, err := wire.MakeEmptyMessage(wire.CmdPeerState)
	if err != nil {
		return err
	}
	msg.(*wire.MessagePeerState).Beacon = blockchain.ChainState{
		serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().(*blockchain.BeaconView).BestBlock.Header.Timestamp,
		serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().(*blockchain.BeaconView).BeaconHeight,
		serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().(*blockchain.BeaconView).BestBlockHash,
		serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().(*blockchain.BeaconView).Hash(),
	}
	for _, shardID := range serverObj.blockChain.Synker.GetCurrentSyncShards() {
		msg.(*wire.MessagePeerState).Shards[shardID] = blockchain.ChainState{
			serverObj.blockChain.Chains[common.GetShardChainKey(shardID)].GetBestView().(*blockchain.ShardView).TipBlock.Header.Timestamp,
			serverObj.blockChain.Chains[common.GetShardChainKey(shardID)].GetBestView().(*blockchain.ShardView).GetHeight(),
			*serverObj.blockChain.Chains[common.GetShardChainKey(shardID)].GetBestView().(*blockchain.ShardView).TipBlock.Hash(),
			serverObj.blockChain.Chains[common.GetShardChainKey(shardID)].GetBestView().(*blockchain.ShardView).Hash(),
		}
	}
	msg.(*wire.MessagePeerState).ShardToBeaconPool = serverObj.shardToBeaconPool.GetValidBlockHeight()

	publicKeyInBase58CheckEncode, _ := serverObj.consensusEngine.GetCurrentMiningPublicKey()
	// signDataInBase58CheckEncode := common.EmptyString
	if publicKeyInBase58CheckEncode != "" {
		_, shardID := serverObj.consensusEngine.GetUserLayer()
		if (cfg.NodeMode == common.NodeModeAuto || cfg.NodeMode == common.NodeModeShard) && shardID >= 0 {
			msg.(*wire.MessagePeerState).CrossShardPool[byte(shardID)] = serverObj.crossShardPool[byte(shardID)].GetValidBlockHeight()
		}
	}
	userKey, _ := serverObj.consensusEngine.GetCurrentMiningPublicKey()
	if userKey != "" {
		// metrics.SetGlobalParam("MINING_PUBKEY", userKey)
		userRole, shardID := serverObj.blockChain.Chains[common.BeaconChainKey].GetBestView().GetPubkeyRole(userKey, 0)
		if (cfg.NodeMode == common.NodeModeAuto || cfg.NodeMode == common.NodeModeShard) && userRole == common.ShardRole {
			userRole, _ = serverObj.blockChain.Chains[common.GetShardChainKey(shardID)].GetBestView().GetPubkeyRole(userKey, 0)
			if userRole == common.ProposerRole || userRole == common.ValidatorRole {
				msg.(*wire.MessagePeerState).CrossShardPool[shardID] = serverObj.crossShardPool[shardID].GetValidBlockHeight()
			}
		}
	}
	msg.SetSenderID(listener.GetPeerID())
	Logger.log.Debugf("Broadcast peerstate from %s", listener.GetRawAddress())
	serverObj.PushMessageToAll(msg)
	return nil
}

func (serverObj *Server) PushMessageGetBlockBeaconByHeight(from uint64, to uint64) error {
	msgs, err := serverObj.highway.Requester.GetBlockBeaconByHeight(
		false, // bySpecific
		from,  // from
		nil,   // heights (this params just != nil if bySpecific == true)
		to,    // to
	)
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	// TODO(@0xbunyip): instead of putting response to queue, use it immediately in synker
	serverObj.putResponseMsgs(msgs)
	return nil
}

func (serverObj *Server) PushMessageGetBlockBeaconBySpecificHeight(heights []uint64, getFromPool bool) error {
	Logger.log.Infof("[byspecific] Get blk beacon by Specific heights %v", heights)
	msgs, err := serverObj.highway.Requester.GetBlockBeaconByHeight(
		true,    // bySpecific
		0,       // from
		heights, // heights (this params just != nil if bySpecific == true)
		0,       // to
	)
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	// TODO(@0xbunyip): instead of putting response to queue, use it immediately in synker
	serverObj.putResponseMsgs(msgs)
	return nil
}

func (serverObj *Server) PushMessageGetBlockBeaconByHash(blkHashes []common.Hash, getFromPool bool, peerID libp2p.ID) error {
	msgs, err := serverObj.highway.Requester.GetBlockBeaconByHash(
		blkHashes, // by blockHashes
	)
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	serverObj.putResponseMsgs(msgs)
	return nil
}

func (serverObj *Server) PushMessageGetBlockShardByHeight(shardID byte, from uint64, to uint64) error {
	msgs, err := serverObj.highway.Requester.GetBlockShardByHeight(
		int32(shardID), // shardID
		false,          // bySpecific
		from,           // from
		nil,            // heights
		to,             // to
	)
	if err != nil {
		Logger.log.Error(err)
		return err
	}

	serverObj.putResponseMsgs(msgs)
	return nil
}

func (serverObj *Server) PushMessageGetBlockShardBySpecificHeight(shardID byte, heights []uint64, getFromPool bool) error {
	Logger.log.Infof("[byspecific] Get blk shard %v by Specific heights %v", shardID, heights)
	msgs, err := serverObj.highway.Requester.GetBlockShardByHeight(
		int32(shardID), // shardID
		true,           // bySpecific
		0,              // from
		heights,        // heights
		0,              // to
	)
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	serverObj.putResponseMsgs(msgs)
	return nil
}

func (serverObj *Server) PushMessageGetBlockShardByHash(shardID byte, blkHashes []common.Hash, getFromPool bool, peerID libp2p.ID) error {
	Logger.log.Infof("[blkbyhash] Get blk shard by hash %v", blkHashes)
	msgs, err := serverObj.highway.Requester.GetBlockShardByHash(
		int32(shardID),
		blkHashes, // by blockHashes
	)
	if err != nil {
		Logger.log.Infof("[blkbyhash] Get blk shard by hash error %v ", err)
		Logger.log.Error(err)
		return err
	}
	Logger.log.Infof("[blkbyhash] Get blk shard by hash get %v ", msgs)

	serverObj.putResponseMsgs(msgs)
	return nil
}

func (serverObj *Server) PushMessageGetBlockShardToBeaconByHeight(shardID byte, from uint64, to uint64) error {
	msgs, err := serverObj.highway.Requester.GetBlockShardToBeaconByHeight(
		int32(shardID),
		false, // by Specific
		from,  // sfrom
		nil,   // nil because request via [from:to]
		to,    // to
	)
	if err != nil {
		Logger.log.Error(err)
		return err
	}

	serverObj.putResponseMsgs(msgs)
	return nil
}

func (serverObj *Server) PushMessageGetBlockShardToBeaconByHash(shardID byte, blkHashes []common.Hash, getFromPool bool, peerID libp2p.ID) error {
	Logger.log.Debugf("Send a GetShardToBeacon")
	listener := serverObj.connManager.GetConfig().ListenerPeer
	msg, err := wire.MakeEmptyMessage(wire.CmdGetShardToBeacon)
	if err != nil {
		return err
	}
	msg.(*wire.MessageGetShardToBeacon).ByHash = true
	msg.(*wire.MessageGetShardToBeacon).FromPool = getFromPool
	msg.(*wire.MessageGetShardToBeacon).ShardID = shardID
	msg.(*wire.MessageGetShardToBeacon).BlkHashes = blkHashes
	msg.(*wire.MessageGetShardToBeacon).Timestamp = time.Now().Unix()
	msg.SetSenderID(listener.GetPeerID())
	Logger.log.Debugf("Send a GetCrossShard from %s", listener.GetRawAddress())
	if peerID == "" {
		return serverObj.PushMessageToShard(msg, shardID, map[libp2p.ID]bool{})
	}
	return serverObj.PushMessageToPeer(msg, peerID.Pretty())
}

func (serverObj *Server) PushMessageGetBlockShardToBeaconBySpecificHeight(
	shardID byte,
	blkHeights []uint64,
	getFromPool bool,
	peerID libp2p.ID,
) error {
	msgs, err := serverObj.highway.Requester.GetBlockShardToBeaconByHeight(
		int32(shardID),
		true,       //by Specific
		0,          //from 0 to 0 because request via blkheights
		blkHeights, //
		0,          // to 0
	)
	if err != nil {
		Logger.log.Error(err)
		return err
	}

	serverObj.putResponseMsgs(msgs)
	return nil

}

func (serverObj *Server) PushMessageGetBlockCrossShardByHash(fromShard byte, toShard byte, blkHashes []common.Hash, getFromPool bool, peerID libp2p.ID) error {
	Logger.log.Debugf("Send a GetCrossShard")
	listener := serverObj.connManager.GetConfig().ListenerPeer
	msg, err := wire.MakeEmptyMessage(wire.CmdGetCrossShard)
	if err != nil {
		return err
	}
	msg.(*wire.MessageGetCrossShard).ByHash = true
	msg.(*wire.MessageGetCrossShard).FromPool = getFromPool
	msg.(*wire.MessageGetCrossShard).FromShardID = fromShard
	msg.(*wire.MessageGetCrossShard).ToShardID = toShard
	msg.(*wire.MessageGetCrossShard).BlkHashes = blkHashes
	msg.(*wire.MessageGetCrossShard).Timestamp = time.Now().Unix()
	msg.SetSenderID(listener.GetPeerID())
	Logger.log.Debugf("Send a GetCrossShard from %s", listener.GetRawAddress())
	if peerID == "" {
		return serverObj.PushMessageToShard(msg, fromShard, map[libp2p.ID]bool{})
	}
	return serverObj.PushMessageToPeer(msg, peerID.Pretty())

}

func (serverObj *Server) PushMessageGetBlockCrossShardBySpecificHeight(fromShard byte, toShard byte, blkHeights []uint64, getFromPool bool, peerID libp2p.ID) error {
	msgs, err := serverObj.highway.Requester.GetBlockCrossShardByHeight(
		int32(fromShard),
		int32(toShard),
		blkHeights,
		getFromPool,
	)
	if err != nil {
		Logger.log.Error(err)
		return err
	}

	serverObj.putResponseMsgs(msgs)
	return nil
}

/*
PushMessageToAll broadcast msg
*/
func (serverObj *Server) PushMessageToAll(msg wire.Message) error {
	Logger.log.Debug("Push msg to all peers")

	// Publish message to highway
	if err := serverObj.highway.PublishMessage(msg); err != nil {
		return err
	}

	return nil
}

/*
PushMessageToPeer push msg to peer
*/
func (serverObj *Server) PushMessageToPeer(msg interface{}, peerIDString string) error {
	Logger.log.Debugf("Push msg to peer %s", peerIDString)
	var dc chan<- struct{}
	peerConn := serverObj.connManager.GetConfig().ListenerPeer.GetPeerConnByPeerID(peerIDString)
	if peerConn != nil {
		msg.(wire.Message).SetSenderID(serverObj.connManager.GetConfig().ListenerPeer.GetPeerID())
		peerConn.QueueMessageWithEncoding(msg.(wire.Message), dc, peer.MessageToPeer, nil)
		Logger.log.Debugf("Pushed peer %s", peerIDString)
		return nil
	} else {
		Logger.log.Error("RemotePeer not exist!")
	}
	return errors.New("RemotePeer not found")
}

/*
PushMessageToPeer push msg to pbk
*/
func (serverObj *Server) PushMessageToPbk(msg wire.Message, pbk string) error {
	Logger.log.Debugf("Push msg to pbk %s", pbk)
	peerConns := serverObj.connManager.GetPeerConnOfPublicKey(pbk)
	if len(peerConns) > 0 {
		for _, peerConn := range peerConns {
			msg.SetSenderID(peerConn.GetListenerPeer().GetPeerID())
			peerConn.QueueMessageWithEncoding(msg, nil, peer.MessageToPeer, nil)
		}
		Logger.log.Debugf("Pushed pbk %s", pbk)
		return nil
	} else {
		Logger.log.Error("RemotePeer not exist!")
	}
	return errors.New("RemotePeer not found")
}

/*
PushMessageToPeer push msg to pbk
*/
func (serverObj *Server) PushMessageToShard(msg wire.Message, shard byte, exclusivePeerIDs map[libp2p.ID]bool) error {
	Logger.log.Debugf("Push msg to shard %d", shard)

	// Publish message to highway
	if err := serverObj.highway.PublishMessageToShard(msg, shard); err != nil {
		return err
	}

	return nil
}

func (serverObj *Server) PushRawBytesToShard(p *peer.PeerConn, msgBytes *[]byte, shard byte) error {
	Logger.log.Debugf("Push raw bytes to shard %d", shard)
	peerConns := serverObj.connManager.GetPeerConnOfShard(shard)
	if len(peerConns) > 0 {
		for _, peerConn := range peerConns {
			if p == nil || peerConn != p {
				peerConn.QueueMessageWithBytes(msgBytes, nil)
			}
		}
		Logger.log.Debugf("Pushed shard %d", shard)
	} else {
		Logger.log.Error("RemotePeer of shard not exist!")
		peerConns := serverObj.connManager.GetPeerConnOfAll()
		for _, peerConn := range peerConns {
			if p == nil || peerConn != p {
				peerConn.QueueMessageWithBytes(msgBytes, nil)
			}
		}
	}
	return nil
}

/*
PushMessageToBeacon push msg to beacon node
*/
func (serverObj *Server) PushMessageToBeacon(msg wire.Message, exclusivePeerIDs map[libp2p.ID]bool) error {
	// Publish message to highway
	if err := serverObj.highway.PublishMessage(msg); err != nil {
		return err
	}

	return nil
}

func (serverObj *Server) PushRawBytesToBeacon(p *peer.PeerConn, msgBytes *[]byte) error {
	Logger.log.Debugf("Push raw bytes to beacon")
	peerConns := serverObj.connManager.GetPeerConnOfBeacon()
	if len(peerConns) > 0 {
		for _, peerConn := range peerConns {
			if p == nil || peerConn != p {
				peerConn.QueueMessageWithBytes(msgBytes, nil)
			}
		}
		Logger.log.Debugf("Pushed raw bytes beacon done")
	} else {
		Logger.log.Error("RemotePeer of beacon raw bytes not exist!")
		peerConns := serverObj.connManager.GetPeerConnOfAll()
		for _, peerConn := range peerConns {
			if p == nil || peerConn != p {
				peerConn.QueueMessageWithBytes(msgBytes, nil)
			}
		}
	}
	return nil
}

// handleAddPeerMsg deals with adding new peers.  It is invoked from the
// peerHandler goroutine.
func (serverObj *Server) handleAddPeerMsg(peer *peer.Peer) bool {
	if peer == nil {
		return false
	}
	Logger.log.Debug("Zero peer have just sent a message version")
	//Logger.log.Debug(peer)
	return true
}

func (serverObj *Server) PushVersionMessage(peerConn *peer.PeerConn) error {
	// push message version
	msg, err := wire.MakeEmptyMessage(wire.CmdVersion)
	msg.(*wire.MessageVersion).Timestamp = time.Now().UnixNano()
	msg.(*wire.MessageVersion).LocalAddress = peerConn.GetListenerPeer().GetListeningAddress()
	msg.(*wire.MessageVersion).RawLocalAddress = peerConn.GetListenerPeer().GetRawAddress()
	msg.(*wire.MessageVersion).LocalPeerId = peerConn.GetListenerPeer().GetPeerID()
	msg.(*wire.MessageVersion).RemoteAddress = peerConn.GetListenerPeer().GetListeningAddress()
	msg.(*wire.MessageVersion).RawRemoteAddress = peerConn.GetListenerPeer().GetRawAddress()
	msg.(*wire.MessageVersion).RemotePeerId = peerConn.GetListenerPeer().GetPeerID()
	msg.(*wire.MessageVersion).ProtocolVersion = serverObj.protocolVersion

	// ValidateTransaction Public Key from ProducerPrvKey
	// publicKeyInBase58CheckEncode, publicKeyType := peerConn.GetListenerPeer().GetConfig().ConsensusEngine.GetCurrentMiningPublicKey()
	signDataInBase58CheckEncode := common.EmptyString
	// if publicKeyInBase58CheckEncode != "" {
	// msg.(*wire.MessageVersion).PublicKey = publicKeyInBase58CheckEncode
	// msg.(*wire.MessageVersion).PublicKeyType = publicKeyType
	// Logger.log.Info("Start Process Discover Peers", publicKeyInBase58CheckEncode)
	// sign data
	msg.(*wire.MessageVersion).PublicKey, msg.(*wire.MessageVersion).PublicKeyType, signDataInBase58CheckEncode, err = peerConn.GetListenerPeer().GetConfig().ConsensusEngine.SignDataWithCurrentMiningKey([]byte(peerConn.GetRemotePeer().GetPeerID().Pretty()))
	if err == nil {
		msg.(*wire.MessageVersion).SignDataB58 = signDataInBase58CheckEncode
	}
	// }
	// if peerConn.GetListenerPeer().GetConfig().UserKeySet != nil {
	// 	msg.(*wire.MessageVersion).PublicKey = peerConn.GetListenerPeer().GetConfig().UserKeySet.GetPublicKeyInBase58CheckEncode()
	// 	signDataB58, err := peerConn.GetListenerPeer().GetConfig().UserKeySet.SignDataInBase58CheckEncode()
	// 	if err == nil {
	// 		msg.(*wire.MessageVersion).SignDataB58 = signDataB58
	// 	}
	// }
	if err != nil {
		return err
	}
	peerConn.QueueMessageWithEncoding(msg, nil, peer.MessageToPeer, nil)
	return nil
}

func (serverObj *Server) PushBlockToPeer(block common.BlockInterface, peerID string) error {
	return nil
}
