package pbft_all

import (
	"blockEmulator/message"
	"encoding/json"
	"log"
)

// This module used in the blockChain using transaction relaying mechanism.
// "Raw" means that the pbft only make block consensus.
type RawBrokerOutsideModule_forB2E struct {
	pbftNode *PbftConsensusNode
}

// msgType canbe defined in message
func (rrom *RawBrokerOutsideModule_forB2E) HandleMessageOutsidePBFT(msgType message.MessageType, content []byte) bool {
	switch msgType {
	case message.CSeqIDinfo:
		rrom.handleSeqIDinfos(content)
	case message.CInject:
		rrom.handleInjectTx(content)
	case message.CInjectHead:
		rrom.handleInjectTxHead(content)
	case message.CRelay:
		rrom.handleRelay(content)
	default:
	}
	return true
}

func (rrom *RawBrokerOutsideModule_forB2E) handleRelay(content []byte) {
	relay := new(message.Relay)
	err := json.Unmarshal(content, relay)
	if err != nil {
		log.Panic(err)
	}
	rrom.pbftNode.pl.Plog.Printf("S%dN%d : has received relay txs from shard %d, the senderSeq is %d\n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID, relay.SenderShardID, relay.SenderSeq)
	rrom.pbftNode.CurChain.Txpool.AddTxs2Pool(relay.Txs)
	rrom.pbftNode.seqMapLock.Lock()
	rrom.pbftNode.seqIDMap[relay.SenderShardID] = relay.SenderSeq
	rrom.pbftNode.seqMapLock.Unlock()
	rrom.pbftNode.pl.Plog.Printf("S%dN%d : has handled relay txs msg\n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID)
}

// receive SeqIDinfo
func (rrom *RawBrokerOutsideModule_forB2E) handleSeqIDinfos(content []byte) {
	sii := new(message.SeqIDinfo)
	err := json.Unmarshal(content, sii)
	if err != nil {
		log.Panic(err)
	}
	rrom.pbftNode.pl.Plog.Printf("S%dN%d : has received SeqIDinfo from shard %d, the senderSeq is %d\n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID, sii.SenderShardID, sii.SenderSeq)
	rrom.pbftNode.seqMapLock.Lock()
	rrom.pbftNode.seqIDMap[sii.SenderShardID] = sii.SenderSeq
	rrom.pbftNode.seqMapLock.Unlock()
	rrom.pbftNode.pl.Plog.Printf("S%dN%d : has handled SeqIDinfo msg\n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID)
}

func (rrom *RawBrokerOutsideModule_forB2E) handleInjectTx(content []byte) {
	it := new(message.InjectTxs)
	err := json.Unmarshal(content, it)
	if err != nil {
		log.Panic(err)
	}
	rrom.pbftNode.CurChain.Txpool.AddTxs2Pool(it.Txs)
	rrom.pbftNode.pl.Plog.Printf("S%dN%d : has handled injected txs msg, txs: %d \n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID, len(it.Txs))
}

func (rrom *RawBrokerOutsideModule_forB2E) handleInjectTxHead(content []byte) {
	it := new(message.InjectTxs)
	err := json.Unmarshal(content, it)
	if err != nil {
		log.Panic(err)
	}
	rrom.pbftNode.CurChain.Txpool.AddTxs2Pool_Head(it.Txs)
	// if len(it.Txs) <= params.MaxBlockSize_global {
	// 	BAT_byte, _ := json.Marshal(it.Txs)
	// 	BAT_size := len(BAT_byte)
	// }
	// else{

	// }
	rrom.pbftNode.pl.Plog.Printf("S%dN%d : has handled injected txs msg, txs: %d \n", rrom.pbftNode.ShardID, rrom.pbftNode.NodeID, len(it.Txs))
}
