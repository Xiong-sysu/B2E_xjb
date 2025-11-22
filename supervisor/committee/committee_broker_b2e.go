package committee

import (
	"blockEmulator/broker"
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"blockEmulator/supervisor/Broker2Earn"
	"blockEmulator/supervisor/signal"
	"blockEmulator/supervisor/supervisor_log"
	"blockEmulator/utils"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CLPA committee operations
type BrokerCommitteeMod_b2e struct {
	csvPath      string
	dataTotalNum int
	nowDataNum   int
	dataTxNums   int
	batchDataNum int

	//broker related  attributes avatar
	broker               *broker.Broker
	brokerConfirm1Pool   map[string]*message.Mag1Confirm
	brokerConfirm2Pool   map[string]*message.Mag2Confirm
	restBrokerRawMegPool []*message.BrokerRawMeg
	brokerTxPool         []*core.Transaction
	brokerModuleLock     sync.Mutex
	brokerBalanceLock    sync.Mutex

	// logger module
	sl *supervisor_log.SupervisorLog

	// control components
	Ss          *signal.StopSignal // to control the stop message sending
	IpNodeTable map[uint64]map[uint64]string

	// log balance
	Result_lockBalance   map[string][]string
	Result_brokerBalance map[string][]string
	Result_Profit        map[string][]string

	// 新增：B2E算法时间统计相关字段
	b2eExecutionTimes []time.Duration
	//totalB2ETransactions int
	totalB2EIterations        int
	epochB2ETransactions      []int
	txlen                     []int
	rest_BrokerRawMegPoolLen  []int
	notHandleCtxByBroker      []int
	len_alloctedBrokerRawMegs []int
}

func NewBrokerCommitteeMod_b2e(Ip_nodeTable map[uint64]map[uint64]string, Ss *signal.StopSignal, sl *supervisor_log.SupervisorLog, csvFilePath string, dataNum, batchNum int) *BrokerCommitteeMod_b2e {

	broker := new(broker.Broker)
	broker.NewBroker(nil)
	result_lockBalance := make(map[string][]string)
	result_brokerBalance := make(map[string][]string)
	result_Profit := make(map[string][]string)
	block_txs := make(map[uint64][]string)

	for _, brokeraddress := range broker.BrokerAddress {
		result_lockBalance[brokeraddress] = make([]string, 0)
		result_brokerBalance[brokeraddress] = make([]string, 0)
		result_Profit[brokeraddress] = make([]string, 0)

		a := ""
		b := ""
		title := ""
		for i := 0; i < params.ShardNum; i++ {
			title += "shard" + strconv.Itoa(i) + ","
			a += params.Init_broker_Balance.String() + ","
			b += "0,"
		}
		result_lockBalance[brokeraddress] = append(result_lockBalance[brokeraddress], title)
		result_brokerBalance[brokeraddress] = append(result_brokerBalance[brokeraddress], title)
		result_Profit[brokeraddress] = append(result_Profit[brokeraddress], title)

		result_lockBalance[brokeraddress] = append(result_lockBalance[brokeraddress], b)
		result_brokerBalance[brokeraddress] = append(result_brokerBalance[brokeraddress], a)
		result_Profit[brokeraddress] = append(result_Profit[brokeraddress], b)
	}
	for i := 0; i < params.ShardNum; i++ {
		block_txs[uint64(i)] = make([]string, 0)
		block_txs[uint64(i)] = append(block_txs[uint64(i)], "txExcuted, broker1Txs, broker2Txs, allocatedTxs")
	}

	return &BrokerCommitteeMod_b2e{
		csvPath:              csvFilePath,
		dataTotalNum:         dataNum,
		batchDataNum:         batchNum,
		nowDataNum:           0,
		dataTxNums:           0,
		brokerConfirm1Pool:   make(map[string]*message.Mag1Confirm),
		brokerConfirm2Pool:   make(map[string]*message.Mag2Confirm),
		restBrokerRawMegPool: make([]*message.BrokerRawMeg, 0),
		brokerTxPool:         make([]*core.Transaction, 0),
		broker:               broker,
		IpNodeTable:          Ip_nodeTable,
		Ss:                   Ss,
		sl:                   sl,
		Result_lockBalance:   result_lockBalance,
		Result_brokerBalance: result_brokerBalance,
		Result_Profit:        result_Profit,

		b2eExecutionTimes: make([]time.Duration, 0),
		//totalB2ETransactions: 0,
		totalB2EIterations:        0,
		epochB2ETransactions:      make([]int, 0),
		txlen:                     make([]int, 0),
		rest_BrokerRawMegPoolLen:  make([]int, 0),
		notHandleCtxByBroker:      make([]int, 0),
		len_alloctedBrokerRawMegs: make([]int, 0),
	}

}

func (bcm *BrokerCommitteeMod_b2e) HandleOtherMessage([]byte) {}

func (bcm *BrokerCommitteeMod_b2e) fetchModifiedMap(key string) uint64 {
	return uint64(utils.Addr2Shard(key))
}

func (bcm *BrokerCommitteeMod_b2e) txSending(txlist []*core.Transaction) {
	// the txs will be sent
	sendToShard := make(map[uint64][]*core.Transaction)

	for idx := 0; idx <= len(txlist); idx++ {
		if idx > 0 && (idx%params.InjectSpeed == 0 || idx == len(txlist)) {
			// send to shard
			for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
				it := message.InjectTxs{
					Txs:       sendToShard[sid],
					ToShardID: sid,
				}
				itByte, err := json.Marshal(it)
				if err != nil {
					log.Panic(err)
				}
				send_msg := message.MergeMessage(message.CInject, itByte)
				go networks.TcpDial(send_msg, bcm.IpNodeTable[sid][0])
			}
			sendToShard = make(map[uint64][]*core.Transaction)
			time.Sleep(time.Second)
		}
		if idx == len(txlist) {
			break
		}
		tx := txlist[idx]
		sendersid := bcm.fetchModifiedMap(tx.Sender)

		if bcm.broker.IsBroker(tx.Sender) {
			sendersid = bcm.fetchModifiedMap(tx.Recipient)
		}
		sendToShard[sendersid] = append(sendToShard[sendersid], tx)
	}
}

func (bcm *BrokerCommitteeMod_b2e) MsgSendingControl() {
	txfile, err := os.Open(bcm.csvPath)
	if err != nil {
		log.Panic(err)
	}
	defer txfile.Close()
	reader := csv.NewReader(txfile)
	txlist := make([]*core.Transaction, 0) // save the txs in this epoch (round)

	recoderNum := 0
	oldNum := 0

	for {
		data, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panic(err)
		}
		if tx, ok := data2tx(data, uint64(bcm.nowDataNum)); ok {
			txlist = append(txlist, tx)
			bcm.nowDataNum++
			bcm.dataTxNums++
		} else {
			continue
		}

		// batch sending condition
		if len(txlist) == int(bcm.batchDataNum) || bcm.dataTxNums == bcm.dataTotalNum {

			itx := bcm.dealTxByBroker(txlist)
			bcm.txSending(itx)

			txlist = make([]*core.Transaction, 0)
			bcm.Ss.StopGap_Reset()
		}

		if bcm.dataTxNums == bcm.dataTotalNum {
			for len(bcm.restBrokerRawMegPool) != 0 {
				if len(bcm.restBrokerRawMegPool) == oldNum {
					recoderNum++
				} else {
					recoderNum = 0
				}
				bcm.dealTxByBroker(txlist)
				if len(bcm.restBrokerRawMegPool) > 0 {
					println("brokerTx value is ", bcm.restBrokerRawMegPool[0].Tx.Value.String())
				}

				time.Sleep(time.Second)
				oldNum = len(bcm.restBrokerRawMegPool)
				// if recoderNum >= 10 {
				// 	break
				// }
			}
			break
		}
	}

}

func (bcm *BrokerCommitteeMod_b2e) HandleBlockInfo(b *message.BlockInfoMsg) {
	bcm.sl.Slog.Printf("received from shard %d in epoch %d.\n", b.SenderShardID, b.Epoch)
	if b.BlockBodyLength == 0 {
		return
	}

	// add createConfirm
	txs := make([]*core.Transaction, 0)
	txs = append(txs, b.Broker1Txs...)
	txs = append(txs, b.Broker2Txs...)
	bcm.brokerModuleLock.Lock()
	// when accept ctx1, update all accounts
	bcm.brokerBalanceLock.Lock()
	println("block length is ", len(b.ExcutedTxs))
	for _, tx := range b.Broker1Txs {
		brokeraddress, sSid, rSid := tx.Recipient, bcm.fetchModifiedMap(tx.OriginalSender), bcm.fetchModifiedMap(tx.FinalRecipient)

		bcm.broker.LockBalance[brokeraddress][rSid].Sub(bcm.broker.LockBalance[brokeraddress][rSid], tx.Value)
		bcm.broker.BrokerBalance[brokeraddress][sSid].Add(bcm.broker.BrokerBalance[brokeraddress][sSid], tx.Value)

		fee := new(big.Float).SetInt64(tx.Fee.Int64())

		fee = fee.Mul(fee, bcm.broker.Brokerage)

		bcm.broker.ProfitBalance[brokeraddress][sSid].Add(bcm.broker.ProfitBalance[brokeraddress][sSid], fee)

	}
	bcm.add_result()
	// bcm.SaveB2ETimeStats()
	bcm.brokerBalanceLock.Unlock()
	bcm.brokerModuleLock.Unlock()
	bcm.createConfirm(txs)
}

func (bcm *BrokerCommitteeMod_b2e) createConfirm(txs []*core.Transaction) {
	confirm1s := make([]*message.Mag1Confirm, 0)
	confirm2s := make([]*message.Mag2Confirm, 0)
	bcm.brokerModuleLock.Lock()
	for _, tx := range txs {
		if confirm1, ok := bcm.brokerConfirm1Pool[string(tx.TxHash)]; ok {
			confirm1s = append(confirm1s, confirm1)
		}
		if confirm2, ok := bcm.brokerConfirm2Pool[string(tx.TxHash)]; ok {
			confirm2s = append(confirm2s, confirm2)
		}
	}
	bcm.brokerModuleLock.Unlock()

	if len(confirm1s) != 0 {
		bcm.handleTx1ConfirmMag(confirm1s)
	}

	if len(confirm2s) != 0 {
		bcm.handleTx2ConfirmMag(confirm2s)
	}
}

func (bcm *BrokerCommitteeMod_b2e) dealTxByBroker(txs []*core.Transaction) (itxs []*core.Transaction) {
	itxs = make([]*core.Transaction, 0)
	brokerRawMegs := make([]*message.BrokerRawMeg, 0)
	bcm.txlen = append(bcm.txlen, len(txs))
	bcm.rest_BrokerRawMegPoolLen = append(bcm.rest_BrokerRawMegPoolLen, len(bcm.restBrokerRawMegPool))

	brokerRecordMegs := make([]*message.BrokerRawMeg, 0)
	//copy(brokerRawMegs, bcm.restBrokerRawMegPool)
	for _, item := range bcm.restBrokerRawMegPool {
		brokerRawMegs = append(brokerRawMegs, item)
	}
	bcm.restBrokerRawMegPool = make([]*message.BrokerRawMeg, 0)

	println("len_brokerRawMegs", len(brokerRawMegs))
	count := 0
	for _, tx := range txs {
		rSid := bcm.fetchModifiedMap(tx.Recipient)
		sSid := bcm.fetchModifiedMap(tx.Sender)
		if rSid != sSid && !bcm.broker.IsBroker(tx.Recipient) && !bcm.broker.IsBroker(tx.Sender) {
			brokerBalance := params.Init_broker_Balance
			if brokerBalance.Cmp(tx.Value) < 0 {
				count++
				continue
			}
			brokerRawMeg := &message.BrokerRawMeg{
				Tx:     tx,
				Broker: bcm.broker.BrokerAddress[0],
			}
			brokerRawMegs = append(brokerRawMegs, brokerRawMeg)
			brokerRecordMegs = append(brokerRecordMegs, brokerRawMeg)
		} else {
			if bcm.broker.IsBroker(tx.Recipient) || bcm.broker.IsBroker(tx.Sender) {
				tx.HasBroker = true
				tx.SenderIsBroker = bcm.broker.IsBroker(tx.Sender)
			}
			itxs = append(itxs, tx)
		}
	}
	bcm.notHandleCtxByBroker = append(bcm.notHandleCtxByBroker, count)
	bcm.brokerBalanceLock.Lock()
	println("len_brokerRecordMegs", len(brokerRecordMegs)) // record new injecting tx nums

	// // 新增：记录交易数量
	// transactionCount := len(brokerRecordMegs)
	// bcm.totalB2ETransactions += transactionCount

	// 新增：记录epoch brokermsg
	// bcm.epochB2ETransactions = append(bcm.epochB2ETransactions, transactionCount)

	// 新增：测量B2E函数执行时间

	startTime := time.Now()
	alloctedBrokerRawMegs, restBrokerRawMeg := Broker2Earn.B2E(brokerRawMegs, bcm.broker.BrokerBalance)
	executionTime := time.Since(startTime)

	// 新增：保存执行时间
	bcm.b2eExecutionTimes = append(bcm.b2eExecutionTimes, executionTime)
	bcm.totalB2EIterations++

	fmt.Printf("New executionTime= %v\n", executionTime)
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")

	//alloctedBrokerRawMegs, restBrokerRawMeg := Broker2Earn.B2E(brokerRawMegs, bcm.broker.BrokerBalance)
	for _, item := range restBrokerRawMeg {
		bcm.restBrokerRawMegPool = append(bcm.restBrokerRawMegPool, item)
	}
	println("len_alloctedBrokerRawMegs", len(alloctedBrokerRawMegs))
	bcm.brokerBalanceLock.Unlock()
	bcm.len_alloctedBrokerRawMegs = append(bcm.len_alloctedBrokerRawMegs, len(alloctedBrokerRawMegs))
	allocatedTxs := bcm.GenerateAllocatedTx(alloctedBrokerRawMegs)
	if len(alloctedBrokerRawMegs) != 0 {
		bcm.handleAllocatedTx(allocatedTxs)
		bcm.lockToken(alloctedBrokerRawMegs)
		bcm.handleBrokerRawMag(alloctedBrokerRawMegs)
	}

	bcm.SaveB2ETimeStats()
	return itxs
}

func (bcm *BrokerCommitteeMod_b2e) lockToken(alloctedBrokerRawMegs []*message.BrokerRawMeg) {
	bcm.brokerBalanceLock.Lock()

	for _, brokerRawMeg := range alloctedBrokerRawMegs {
		tx := brokerRawMeg.Tx
		brokerAddress := brokerRawMeg.Broker
		rSid := bcm.fetchModifiedMap(tx.Recipient)
		bcm.broker.LockBalance[brokerAddress][rSid].Add(bcm.broker.LockBalance[brokerAddress][rSid], tx.Value)
		bcm.broker.BrokerBalance[brokerAddress][rSid].Sub(bcm.broker.BrokerBalance[brokerAddress][rSid], tx.Value)
	}

	bcm.brokerBalanceLock.Unlock()
}
func (bcm *BrokerCommitteeMod_b2e) handleAllocatedTx(alloctedTx map[uint64][]*core.Transaction) {

	bcm.brokerBalanceLock.Lock()
	for shardId, txs := range alloctedTx {
		for _, tx := range txs {
			if tx.IsAllocatedSender {
				bcm.broker.BrokerBalance[tx.Sender][shardId].Sub(bcm.broker.BrokerBalance[tx.Sender][shardId], tx.Value)
			}
			if tx.IsAllocatedRecipent {
				bcm.broker.BrokerBalance[tx.Recipient][shardId].Add(bcm.broker.BrokerBalance[tx.Recipient][shardId], tx.Value)
			}
		}

		it := message.InjectTxs{
			Txs:       txs,
			ToShardID: shardId,
		}
		itByte, err := json.Marshal(it)
		if err != nil {
			log.Panic(err)
		}
		send_msg := message.MergeMessage(message.CInjectHead, itByte)
		go networks.TcpDial(send_msg, bcm.IpNodeTable[shardId][0])
		time.Sleep(time.Second)
	}
	bcm.brokerBalanceLock.Unlock()
}

func (bcm *BrokerCommitteeMod_b2e) GenerateAllocatedTx(alloctedBrokerRawMegs []*message.BrokerRawMeg) map[uint64][]*core.Transaction {
	//bcm.broker.BrokerBalance
	brokerNewBalance := make(map[string]map[uint64]*big.Int)
	brokerChange := make(map[string]map[uint64]*big.Int)
	brokerPeekChange := make(map[string]map[uint64]*big.Int)

	// 1. init
	alloctedTxs := make(map[uint64][]*core.Transaction)
	for i := 0; i < params.ShardNum; i++ {
		alloctedTxs[uint64(i)] = make([]*core.Transaction, 0)
	}
	for brokerAddress, shardMap := range bcm.broker.BrokerBalance {
		brokerNewBalance[brokerAddress] = make(map[uint64]*big.Int)
		brokerChange[brokerAddress] = make(map[uint64]*big.Int)
		brokerPeekChange[brokerAddress] = make(map[uint64]*big.Int)
		for shardId, balance := range shardMap {
			brokerNewBalance[brokerAddress][shardId] = new(big.Int).Set(balance)
			brokerChange[brokerAddress][shardId] = big.NewInt(0)
			brokerPeekChange[brokerAddress][shardId] = new(big.Int).Set(balance)
		}

	}
	for _, brokerRawMeg := range alloctedBrokerRawMegs {
		sSid := bcm.fetchModifiedMap(brokerRawMeg.Tx.Sender)
		rSid := bcm.fetchModifiedMap(brokerRawMeg.Tx.Recipient)
		brokerAddress := brokerRawMeg.Broker

		brokerNewBalance[brokerAddress][sSid].Add(brokerNewBalance[brokerAddress][sSid], brokerRawMeg.Tx.Value)
		brokerNewBalance[brokerAddress][rSid].Sub(brokerNewBalance[brokerAddress][rSid], brokerRawMeg.Tx.Value)

		brokerPeekChange[brokerAddress][rSid].Sub(brokerPeekChange[brokerAddress][rSid], brokerRawMeg.Tx.Value)
	}

	// generate tx
	bcm.brokerBalanceLock.Lock()

	for brokerAddress, shardMap := range brokerPeekChange {
		for shardId, _ := range shardMap {

			peekBalance := brokerPeekChange[brokerAddress][shardId]

			if peekBalance.Cmp(big.NewInt(0)) < 0 {
				// If FromShard does not have enough balance, find another shard to cover the deficit

				deficit := new(big.Int).Set(peekBalance)
				deficit.Abs(deficit)
				for id, balance := range brokerPeekChange[brokerAddress] {
					if deficit.Cmp(big.NewInt(0)) == 0 {
						break
					}
					if id != shardId && balance.Cmp(big.NewInt(0)) > 0 {
						tmpValue := new(big.Int).Set(deficit)
						if balance.Cmp(deficit) < 0 {
							tmpValue.Set(balance)
							deficit.Sub(deficit, balance)
						} else {
							deficit.SetInt64(0)
						}
						//brokerChange[brokerAddress][id].Sub(brokerChange[brokerAddress][id], tmpValue)
						//brokerChange[brokerAddress][shardId].Add(brokerChange[brokerAddress][shardId], tmpValue)
						brokerNewBalance[brokerAddress][id].Sub(brokerNewBalance[brokerAddress][id], tmpValue)
						brokerNewBalance[brokerAddress][shardId].Add(brokerNewBalance[brokerAddress][shardId], tmpValue)

						brokerPeekChange[brokerAddress][id].Sub(brokerPeekChange[brokerAddress][id], tmpValue)
						brokerPeekChange[brokerAddress][shardId].Add(brokerPeekChange[brokerAddress][shardId], tmpValue)

						brokerChange[brokerAddress][id].Sub(brokerChange[brokerAddress][id], tmpValue)
						brokerChange[brokerAddress][shardId].Add(brokerChange[brokerAddress][shardId], tmpValue)
					}
				}
			}
		}

	}
	// generate allocated tx

	for brokerAddress, shardMap := range brokerChange {
		for shardId, _ := range shardMap {

			diff := brokerChange[brokerAddress][shardId]

			if diff.Cmp(big.NewInt(0)) == 0 {
				continue
			}
			tx := core.NewTransaction(brokerAddress, brokerAddress, new(big.Int).Abs(diff), uint64(bcm.nowDataNum), big.NewInt(0))

			bcm.nowDataNum++
			if diff.Cmp(big.NewInt(0)) < 0 {
				tx.IsAllocatedSender = true
			} else {
				tx.IsAllocatedRecipent = true
			}
			alloctedTxs[shardId] = append(alloctedTxs[shardId], tx)
		}

	}

	bcm.brokerBalanceLock.Unlock()
	return alloctedTxs
}

func (bcm *BrokerCommitteeMod_b2e) handleBrokerType1Mes(brokerType1Megs []*message.BrokerType1Meg) {
	tx1s := make([]*core.Transaction, 0)
	for _, brokerType1Meg := range brokerType1Megs {
		ctx := brokerType1Meg.RawMeg.Tx
		tx1 := core.NewTransaction(ctx.Sender, brokerType1Meg.Broker, ctx.Value, ctx.Nonce, ctx.Fee)
		tx1.OriginalSender = ctx.Sender
		tx1.FinalRecipient = ctx.Recipient
		tx1.RawTxHash = make([]byte, len(ctx.TxHash))
		copy(tx1.RawTxHash, ctx.TxHash)
		tx1s = append(tx1s, tx1)
		confirm1 := &message.Mag1Confirm{
			RawMeg:  brokerType1Meg.RawMeg,
			Tx1Hash: tx1.TxHash,
		}
		bcm.brokerModuleLock.Lock()
		bcm.brokerConfirm1Pool[string(tx1.TxHash)] = confirm1
		bcm.brokerModuleLock.Unlock()
	}
	bcm.txSending(tx1s)
	fmt.Println("BrokerType1Mes received by shard,  add brokerTx1 len ", len(tx1s))
}

func (bcm *BrokerCommitteeMod_b2e) handleBrokerType2Mes(brokerType2Megs []*message.BrokerType2Meg) {
	tx2s := make([]*core.Transaction, 0)
	for _, mes := range brokerType2Megs {
		ctx := mes.RawMeg.Tx
		tx2 := core.NewTransaction(mes.Broker, ctx.Recipient, ctx.Value, ctx.Nonce, ctx.Fee)
		tx2.OriginalSender = ctx.Sender
		tx2.FinalRecipient = ctx.Recipient
		tx2.RawTxHash = make([]byte, len(ctx.TxHash))
		copy(tx2.RawTxHash, ctx.TxHash)
		tx2s = append(tx2s, tx2)

		confirm2 := &message.Mag2Confirm{
			RawMeg:  mes.RawMeg,
			Tx2Hash: tx2.TxHash,
		}
		bcm.brokerModuleLock.Lock()
		bcm.brokerConfirm2Pool[string(tx2.TxHash)] = confirm2
		bcm.brokerModuleLock.Unlock()
	}
	bcm.txSending(tx2s)
	fmt.Println("broker tx2 add to pool len ", len(tx2s))
}

// get the digest of rawMeg
func (bcm *BrokerCommitteeMod_b2e) getBrokerRawMagDigest(r *message.BrokerRawMeg) []byte {
	b, err := json.Marshal(r)
	if err != nil {
		log.Panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}

func (bcm *BrokerCommitteeMod_b2e) handleBrokerRawMag(brokerRawMags []*message.BrokerRawMeg) {
	b := bcm.broker
	brokerType1Mags := make([]*message.BrokerType1Meg, 0)
	fmt.Println("broker receive ctx ", len(brokerRawMags))
	bcm.brokerModuleLock.Lock()
	for _, meg := range brokerRawMags {
		b.BrokerRawMegs[string(bcm.getBrokerRawMagDigest(meg))] = meg
		brokerType1Mag := &message.BrokerType1Meg{
			RawMeg:   meg,
			Hcurrent: 0,
			Broker:   meg.Broker,
		}
		brokerType1Mags = append(brokerType1Mags, brokerType1Mag)
	}
	bcm.brokerModuleLock.Unlock()
	bcm.handleBrokerType1Mes(brokerType1Mags)
}

func (bcm *BrokerCommitteeMod_b2e) handleTx1ConfirmMag(mag1confirms []*message.Mag1Confirm) {
	brokerType2Mags := make([]*message.BrokerType2Meg, 0)
	b := bcm.broker

	fmt.Println("receive confirm  brokerTx1 len ", len(mag1confirms))
	bcm.brokerModuleLock.Lock()
	for _, mag1confirm := range mag1confirms {
		RawMeg := mag1confirm.RawMeg
		_, ok := b.BrokerRawMegs[string(bcm.getBrokerRawMagDigest(RawMeg))]
		if !ok {
			fmt.Println("raw message is not exited,tx1 confirms failure !")
			continue
		}
		b.RawTx2BrokerTx[string(RawMeg.Tx.TxHash)] = append(b.RawTx2BrokerTx[string(RawMeg.Tx.TxHash)], string(mag1confirm.Tx1Hash))
		brokerType2Mag := &message.BrokerType2Meg{
			Broker: RawMeg.Broker,
			RawMeg: RawMeg,
		}
		brokerType2Mags = append(brokerType2Mags, brokerType2Mag)
	}
	bcm.brokerModuleLock.Unlock()
	bcm.handleBrokerType2Mes(brokerType2Mags)
}

func (bcm *BrokerCommitteeMod_b2e) handleTx2ConfirmMag(mag2confirms []*message.Mag2Confirm) {
	b := bcm.broker
	fmt.Println("receive confirm  brokerTx2 len ", len(mag2confirms))
	num := 0
	bcm.brokerModuleLock.Lock()
	for _, mag2confirm := range mag2confirms {
		RawMeg := mag2confirm.RawMeg
		b.RawTx2BrokerTx[string(RawMeg.Tx.TxHash)] = append(b.RawTx2BrokerTx[string(RawMeg.Tx.TxHash)], string(mag2confirm.Tx2Hash))
		if len(b.RawTx2BrokerTx[string(RawMeg.Tx.TxHash)]) == 2 {
			num++
		} else {
			fmt.Println(len(b.RawTx2BrokerTx[string(RawMeg.Tx.TxHash)]))
		}
	}
	bcm.brokerModuleLock.Unlock()
	fmt.Println("finish ctx with adding tx1 and tx2 to txpool,len", num)
}

func (bcm *BrokerCommitteeMod_b2e) add_result() {

	// 确保目录存在
	dirpath := params.DataWrite_path + "brokerRsult/"
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}

	for brokerAddress, shardMap := range bcm.broker.BrokerBalance {
		a := ""
		b := ""
		c := ""
		for shardId, balance := range shardMap {
			a += balance.String() + ","
			b += bcm.broker.LockBalance[brokerAddress][shardId].String() + ","
			c += bcm.broker.ProfitBalance[brokerAddress][shardId].String() + ","
		}
		a += "\n"
		b += "\n"
		c += "\n"
		bcm.Result_lockBalance[brokerAddress] = append(bcm.Result_lockBalance[brokerAddress], b)
		bcm.Result_brokerBalance[brokerAddress] = append(bcm.Result_brokerBalance[brokerAddress], a)
		bcm.Result_Profit[brokerAddress] = append(bcm.Result_Profit[brokerAddress], c)

		// 实时写入到文件 - 只写入最新的一行数据
		targetPath0 := dirpath + brokerAddress + "_lockBalance.csv"
		targetPath1 := dirpath + brokerAddress + "_brokerBalance.csv"
		targetPath2 := dirpath + brokerAddress + "_Profit.csv"

		// 直接写入最新的行，而不是整个数组
		bcm.writeLatestRow(targetPath0, []string{b})
		bcm.writeLatestRow(targetPath1, []string{a})
		bcm.writeLatestRow(targetPath2, []string{c})
	}
}

// 新增方法：只写入最新的一行数据
func (bcm *BrokerCommitteeMod_b2e) writeLatestRow(targetPath string, latestRows []string) {
	var file *os.File
	var err error

	// 检查文件是否存在，不存在则创建并添加标题
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		file, err = os.Create(targetPath)
		if err != nil {
			log.Panic(err)
		}

		// 写入标题行（假设标题格式与原始数据一致）
		w := csv.NewWriter(file)
		// 标题行示例，实际应根据您的数据结构调整
		title := []string{"Shard0", "Shard1", "Shard2", "Shard3"} // 假设有4个分片
		w.Write(title)
		w.Flush()
	} else {
		// 文件存在则以追加模式打开
		file, err = os.OpenFile(targetPath, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Panic(err)
		}
	}

	defer file.Close()
	w := csv.NewWriter(file)

	for _, str := range latestRows {
		str_arry := strings.Split(str, ",")
		if len(str_arry) > 0 && str_arry[len(str_arry)-1] == "" {
			err = w.Write(str_arry[0 : len(str_arry)-1])
		} else {
			err = w.Write(str_arry)
		}
		if err != nil {
			log.Panic(err)
		}
		w.Flush()
		// 立即将数据刷新到磁盘
		if err := file.Sync(); err != nil {
			log.Printf("Warning: Failed to sync file %s: %v", targetPath, err)
		}
	}
}

// 新增：保存B2E时间统计数据到CSV文件
func (bcm *BrokerCommitteeMod_b2e) SaveB2ETimeStats() {
	if len(bcm.b2eExecutionTimes) == 0 {
		fmt.Println("no data here #####")
		return // 没有统计数据，直接返回
	}

	// 准备数据
	var totalTime time.Duration
	for _, duration := range bcm.b2eExecutionTimes {
		totalTime += duration
	}
	//averageTime := totalTime / time.Duration(bcm.totalB2EIterations)

	// 创建保存目录
	dirpath := params.DataWrite_path + "b2e_execution_time/"
	// dirpath := params.DataWrite_path + "brokerRsult/"
	fmt.Println("dirpath ", dirpath)

	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}

	// 生成带时间戳的文件名
	// timestamp := time.Now().Format("20060102_150405")
	targetPath := dirpath + "b2e_execution_time" + ".csv"

	// 打开文件（创建或追加）
	var file *os.File
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		file, err = os.Create(targetPath)
		if err != nil {
			log.Panic(err)
		}
		w := csv.NewWriter(file)
		w.Write([]string{"Iteration", "ExecutionTime(ms)", "epoch_tx_injected", "rest_BrokerRawMegPoolLen", "notHandleCtxByBroker", "len_alloctedBrokerRawMegs"})
		w.Flush()
	} else {
		file, err = os.OpenFile(targetPath, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Panic(err)
		}
	}
	defer file.Close()

	w := csv.NewWriter(file)
	// 写入最新的执行时间数据
	latestTime := bcm.b2eExecutionTimes[len(bcm.b2eExecutionTimes)-1]
	iteration := bcm.totalB2EIterations

	row := []string{
		strconv.Itoa(iteration),
		strconv.FormatFloat(float64(latestTime.Microseconds())/1000.0, 'f', 6, 64),
		strconv.Itoa(bcm.txlen[iteration-1]),
		strconv.Itoa(bcm.rest_BrokerRawMegPoolLen[iteration-1]),
		strconv.Itoa(bcm.notHandleCtxByBroker[iteration-1]),
		strconv.Itoa(bcm.len_alloctedBrokerRawMegs[iteration-1]),
	}

	w.Write(row)
	w.Flush()
	// 立即刷新到磁盘
	file.Sync()

	// // 计算并保存汇总数据
	// targetSummaryPath := dirpath + "b2e_summary.csv"
	// // 计算统计数据（这里可以添加您需要的汇总计算）
	// var totalTime time.Duration
	// for _, duration := range bcm.b2eExecutionTimes {
	//     totalTime += duration
	// }

	/*
		// 准备CSV数据
		var resultStr []string

		fmt.Println("totalTime ", totalTime.Milliseconds())
		fmt.Println("########################################################################")
		// 添加总统计信息
		totalInfo := fmt.Sprintf("总迭代次数,%d,总执行时间,%d", bcm.totalB2EIterations, totalTime.Milliseconds())
		resultStr = append(resultStr, totalInfo)

		// 添加表头
		header := "迭代序号,执行时间(微秒),epoch注入交易数, oldtxlen"
		resultStr = append(resultStr, header)

		// 添加每次迭代的详细信息
		for i, duration := range bcm.b2eExecutionTimes {
			row := fmt.Sprintf("%d,%d, %d, %d", i+1, duration.Microseconds(), bcm.txlen[i], bcm.oldtxlen[i])
			resultStr = append(resultStr, row)
		}
	*/

	// 使用现有的Wirte_result方法写入CSV
	// bcm.Wirte_result(targetPath, resultStr)
}

func (bcm *BrokerCommitteeMod_b2e) Result_save() {

	// write to .csv file
	dirpath := params.DataWrite_path + "brokerRsult/"
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}
	for brokerAddress, _ := range bcm.broker.BrokerBalance {
		targetPath0 := dirpath + brokerAddress + "_lockBalance.csv"
		targetPath1 := dirpath + brokerAddress + "_brokerBalance.csv"
		targetPath2 := dirpath + brokerAddress + "_Profit.csv"
		bcm.Wirte_result(targetPath0, bcm.Result_lockBalance[brokerAddress])
		bcm.Wirte_result(targetPath1, bcm.Result_brokerBalance[brokerAddress])
		bcm.Wirte_result(targetPath2, bcm.Result_Profit[brokerAddress])
	}
	// bcm.SaveB2ETimeStats()
}
func (bcm *BrokerCommitteeMod_b2e) Wirte_result(targetPath string, resultStr []string) {

	f, err := os.Open(targetPath)
	if err != nil && os.IsNotExist(err) {
		file, er := os.Create(targetPath)
		if er != nil {
			panic(er)
		}
		defer file.Close()

		w := csv.NewWriter(file)
		w.Flush()
		for _, str := range resultStr {
			str_arry := strings.Split(str, ",")
			if len(str_arry) > 0 && str_arry[len(str_arry)-1] == "" {
				w.Write(str_arry[0 : len(str_arry)-1])
			} else {
				// 否则保留所有元素
				w.Write(str_arry)
			}
			// w.Write(str_arry[0 : len(str_arry)-1])
			w.Flush()
		}
	} else {
		file, err := os.OpenFile(targetPath, os.O_APPEND|os.O_RDWR, 0666)

		if err != nil {
			log.Panic(err)
		}
		defer file.Close()
		writer := csv.NewWriter(file)

		for _, str := range resultStr {
			str_arry := strings.Split(str, ",")
			if len(str_arry) > 0 && str_arry[len(str_arry)-1] == "" {
				err = writer.Write(str_arry[0 : len(str_arry)-1])
			} else {
				err = writer.Write(str_arry)
			}
			if err != nil {
				log.Panic(err)
			}
			writer.Flush()
		}
		// err = writer.Write(resultStr)
		// if err != nil {
		// 	log.Panic()
		// }
		// writer.Flush()
	}
	f.Close()
}
