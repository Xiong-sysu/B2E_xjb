// addtional module for new consensus
package pbft_all

import (
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

// simple implementation of pbftHandleModule interface ...
// only for block request
type RawBrokerPbftExtraHandleMod_forB2E struct {
	pbftNode *PbftConsensusNode
}

// propose request with different types
func (rbhm *RawBrokerPbftExtraHandleMod_forB2E) HandleinPropose() (bool, *message.Request) {
	// new blocks
	block := rbhm.pbftNode.CurChain.GenerateBlock()
	r := &message.Request{
		RequestType: message.BlockRequest,
		ReqTime:     time.Now(),
	}
	r.Msg.Content = block.Encode()

	return true, r
}

// the diy operation in preprepare
func (rbhm *RawBrokerPbftExtraHandleMod_forB2E) HandleinPrePrepare(ppmsg *message.PrePrepare) bool {
	if rbhm.pbftNode.CurChain.IsValidBlock(core.DecodeB(ppmsg.RequestMsg.Msg.Content)) != nil {
		rbhm.pbftNode.pl.Plog.Printf("S%dN%d : not a valid block\n", rbhm.pbftNode.ShardID, rbhm.pbftNode.NodeID)
		return false
	}
	rbhm.pbftNode.pl.Plog.Printf("S%dN%d : the pre-prepare message is correct, putting it into the RequestPool. \n", rbhm.pbftNode.ShardID, rbhm.pbftNode.NodeID)
	rbhm.pbftNode.requestPool[string(ppmsg.Digest)] = ppmsg.RequestMsg
	// merge to be a prepare message
	return true
}

// the operation in prepare, and in pbft + tx relaying, this function does not need to do any.
func (rbhm *RawBrokerPbftExtraHandleMod_forB2E) HandleinPrepare(pmsg *message.Prepare) bool {
	fmt.Println("No operations are performed in Extra handle mod")
	return true
}

// the operation in commit.
func (rbhm *RawBrokerPbftExtraHandleMod_forB2E) HandleinCommit(cmsg *message.Commit) bool {
	r := rbhm.pbftNode.requestPool[string(cmsg.Digest)]
	// requestType ...
	block := core.DecodeB(r.Msg.Content)
	rbhm.pbftNode.pl.Plog.Printf("S%dN%d : adding the block %d...now height = %d \n", rbhm.pbftNode.ShardID, rbhm.pbftNode.NodeID, block.Header.Number, rbhm.pbftNode.CurChain.CurrentBlock.Header.Number)
	rbhm.pbftNode.CurChain.AddBlock(block)
	rbhm.pbftNode.pl.Plog.Printf("S%dN%d : added the block %d... \n", rbhm.pbftNode.ShardID, rbhm.pbftNode.NodeID, block.Header.Number)
	rbhm.pbftNode.CurChain.PrintBlockChain()

	// now try to relay txs to other shards (for main nodes)
	if rbhm.pbftNode.NodeID == rbhm.pbftNode.view {
		// do normal operations for block
		rbhm.pbftNode.pl.Plog.Printf("S%dN%d : main node is trying to send relay txs at height = %d \n", rbhm.pbftNode.ShardID, rbhm.pbftNode.NodeID, block.Header.Number)
		// generate brokertxs and collect txs excuted
		txExcuted := make([]*core.Transaction, 0)
		broker1Txs := make([]*core.Transaction, 0)
		broker2Txs := make([]*core.Transaction, 0)
		allocatedTxs := make([]*core.Transaction, 0)
		relayer_1_2_txs := make([]*core.Transaction, 0)
		relay2Txs_num := 0

		h := 0
		// relay tx for B2E
		rbhm.pbftNode.CurChain.Txpool.RelayPool = make(map[uint64][]*core.Transaction)
		relay1Txs := make([]*core.Transaction, 0)
		innertxs := make([]*core.Transaction, 0)

		// Collect per-transaction CSV rows for this block, then write them in batch to reduce IO.
		txRows := make([][]string, 0, len(block.Body))

		// generate block infos
		for _, tx := range block.Body {
			tx.Commit_time = time.Now()
			rsid := rbhm.pbftNode.CurChain.Get_PartitionMap(tx.Recipient)
			ssid := rbhm.pbftNode.CurChain.Get_PartitionMap(tx.Sender)
			if tx.IsAllocatedRecipent || tx.IsAllocatedSender {
				allocatedTxs = append(allocatedTxs, tx)
				continue
			}
			isInnerShardTx := tx.RawTxHash == nil // 这里也有可能是需要relay的tx，只有broker tx才会有RawTxHash
			tx.IsBroker1Tx = !isInnerShardTx && tx.Sender == tx.OriginalSender && !tx.IsRelay
			tx.IsBroker2Tx = !isInnerShardTx && tx.Recipient == tx.FinalRecipient && !tx.IsRelay
			if tx.IsBroker2Tx {
				broker2Txs = append(broker2Txs, tx)
			} else if tx.IsBroker1Tx {
				broker1Txs = append(broker1Txs, tx)
			} else {
				txExcuted = append(txExcuted, tx) // txExcuted 包含 itx 和 relay1+relayer2 tx（ctx）， 不包含 broker tx 和 allocated tx
				if rsid == ssid {
					innertxs = append(innertxs, tx)
					tx.IsNormalItx = true
				} else {
					relayer_1_2_txs = append(relayer_1_2_txs, tx) // 这里剩下的是不是bat、broker、普通itx，那就是relay 和 has broker
					if tx.HasBroker {
						h++
					}
				}
			}

			// if !tx.IsAllocatedRecipent && !tx.IsAllocatedSender && !isBroker1Tx && !isBroker2Tx {
			// add for relay //原本的判断条件会多很多笔 relay 交易，很奇怪. //解答了，多的是 hasbroker 的
			if tx.IsRelay {
				if rsid != rbhm.pbftNode.ShardID {
					ntx := tx
					ntx.Relayed = true
					rbhm.pbftNode.CurChain.Txpool.AddRelayTx(ntx, rsid)
					relay1Txs = append(relay1Txs, tx)
					tx.IsRelay1 = true
				}

				if rsid == rbhm.pbftNode.ShardID && ssid != rbhm.pbftNode.ShardID && tx.Relayed {
					relay2Txs_num++
					tx.IsRelay2 = true
				}
			}

			tx_makespan := tx.Commit_time.Sub(tx.Time).Milliseconds()

			row := []string{
				strconv.Itoa(int(block.Header.Number)),
				strconv.FormatInt(tx.Time.UnixMilli(), 10),        // 毫秒时间戳
				strconv.FormatInt(tx.Commit_time.UnixMilli(), 10), // 毫秒时间戳
				strconv.FormatFloat(float64(tx_makespan), 'f', 6, 64),
				strconv.FormatBool(tx.IsAllocatedRecipent || tx.IsAllocatedSender),
				strconv.FormatBool(tx.IsBroker1Tx),
				strconv.FormatBool(tx.IsBroker2Tx),
				strconv.FormatBool(tx.HasBroker),
				strconv.FormatBool(tx.IsNormalItx),
				strconv.FormatBool(tx.IsRelay1),
				strconv.FormatBool(tx.IsRelay2),
			}
			txRows = append(txRows, row)
		}

		// send relay txs
		for sid := uint64(0); sid < rbhm.pbftNode.pbftChainConfig.ShardNums; sid++ {
			if sid == rbhm.pbftNode.ShardID {
				continue
			}
			relay := message.Relay{
				Txs:           rbhm.pbftNode.CurChain.Txpool.RelayPool[sid],
				SenderShardID: rbhm.pbftNode.ShardID,
				SenderSeq:     rbhm.pbftNode.sequenceID,
			}
			rByte, err := json.Marshal(relay)
			if err != nil {
				log.Panic()
			}
			msg_send := message.MergeMessage(message.CRelay, rByte)
			go networks.TcpDial(msg_send, rbhm.pbftNode.ip_nodeTable[sid][0])
			rbhm.pbftNode.pl.Plog.Printf("S%dN%d : sended relay txs to %d\n", rbhm.pbftNode.ShardID, rbhm.pbftNode.NodeID, sid)
		}
		rbhm.pbftNode.CurChain.Txpool.ClearRelayPool()

		batByte, _ := json.Marshal(allocatedTxs)
		Bat_byte_Size := len(batByte)

		block_byte, _ := json.Marshal(block)
		Block_byte_Size := len(block_byte)

		Bat_byte_ratio := float64(Bat_byte_Size) / float64(Block_byte_Size) // B

		/// 原B2E，注释掉了发送seqID的代码
		// // send seqID
		// for sid := uint64(0); sid < rbhm.pbftNode.pbftChainConfig.ShardNums; sid++ {
		// 	if sid == rbhm.pbftNode.ShardID {
		// 		continue
		// 	}
		// 	sii := message.SeqIDinfo{
		// 		SenderShardID: rbhm.pbftNode.ShardID,
		// 		SenderSeq:     rbhm.pbftNode.sequenceID,
		// 	}
		// 	sByte, err := json.Marshal(sii)
		// 	if err != nil {
		// 		log.Panic()
		// 	}
		// 	msg_send := message.MergeMessage(message.CSeqIDinfo, sByte)
		// 	go networks.TcpDial(msg_send, rbhm.pbftNode.ip_nodeTable[sid][0])
		// 	rbhm.pbftNode.pl.Plog.Printf("S%dN%d : sended sequence ids to %d\n", rbhm.pbftNode.ShardID, rbhm.pbftNode.NodeID, sid)
		// }

		// send txs excuted in this block to the listener
		// add more message to measure more metrics
		bim := message.BlockInfoMsg{
			BlockBodyLength: len(block.Body),
			ExcutedTxs:      txExcuted,
			Relay1Txs:       relay1Txs,
			Relay1TxNum:     uint64(len(relay1Txs)),
			Broker1TxNum:    uint64(len(broker1Txs)),
			Broker1Txs:      broker1Txs,
			Broker2TxNum:    uint64(len(broker2Txs)),
			Broker2Txs:      broker2Txs,
			AllocatedTxs:    allocatedTxs,
			Epoch:           0,
			SenderShardID:   rbhm.pbftNode.ShardID,
			ProposeTime:     r.ReqTime,
			CommitTime:      time.Now(),
			Bat_byte_Size:   Bat_byte_Size,
			Block_byte_Size: Block_byte_Size,
			Bat_byte_ratio:  Bat_byte_ratio,
			Relay2TxNum:     uint64(relay2Txs_num),
		}
		bByte, err := json.Marshal(bim)
		if err != nil {
			log.Panic()
		}
		msg_send := message.MergeMessage(message.CBlockInfo, bByte)
		go networks.TcpDial(msg_send, rbhm.pbftNode.ip_nodeTable[params.DeciderShard][0])
		rbhm.pbftNode.pl.Plog.Printf("S%dN%d : sended excuted txs\n", rbhm.pbftNode.ShardID, rbhm.pbftNode.NodeID)
		rbhm.pbftNode.CurChain.Txpool.GetLocked()
		rbhm.pbftNode.writeCSVline([]string{strconv.Itoa(int(block.Header.Number)), strconv.Itoa(len(rbhm.pbftNode.CurChain.Txpool.TxQueue)), strconv.Itoa(bim.BlockBodyLength), strconv.Itoa(len(txExcuted)),
			strconv.Itoa(len(bim.Broker1Txs)), strconv.Itoa(len(bim.Broker2Txs)), strconv.Itoa(len(bim.AllocatedTxs)), strconv.Itoa(bim.Bat_byte_Size), strconv.Itoa(bim.Block_byte_Size), strconv.FormatFloat(bim.Bat_byte_ratio, 'f', 6, 64), strconv.Itoa(int(bim.Relay1TxNum)), strconv.Itoa(int(bim.Relay2TxNum)), strconv.Itoa(len(innertxs)), strconv.Itoa(len(relayer_1_2_txs)), strconv.Itoa(h)})

		// write all tx rows for this block in one append operation
		rbhm.pbftNode.writeCSV_txTime(txRows)
		rbhm.pbftNode.CurChain.Txpool.GetUnlocked()
	}
	return true
}

func (rbhm *RawBrokerPbftExtraHandleMod_forB2E) HandleReqestforOldSeq(*message.RequestOldMessage) bool {
	fmt.Println("No operations are performed in Extra handle mod")
	return true
}

// the operation for sequential requests
func (rbhm *RawBrokerPbftExtraHandleMod_forB2E) HandleforSequentialRequest(som *message.SendOldMessage) bool {
	if int(som.SeqStartHeight-som.SeqEndHeight) != len(som.OldRequest) {
		rbhm.pbftNode.pl.Plog.Printf("S%dN%d : the SendOldMessage message is not enough\n", rbhm.pbftNode.ShardID, rbhm.pbftNode.NodeID)
	} else { // add the block into the node pbft blockchain
		for height := som.SeqStartHeight; height <= som.SeqEndHeight; height++ {
			r := som.OldRequest[height-som.SeqStartHeight]
			if r.RequestType == message.BlockRequest {
				b := core.DecodeB(r.Msg.Content)
				rbhm.pbftNode.CurChain.AddBlock(b)
			}
		}
		rbhm.pbftNode.sequenceID = som.SeqEndHeight + 1
		rbhm.pbftNode.CurChain.PrintBlockChain()
	}
	return true
}
