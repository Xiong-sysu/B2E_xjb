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
	"bufio"
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
	BAT_size                  []float64

	currentBlockHeight uint64            // 全局区块高度
	shardBlockCount    map[uint64]uint64 // 每个分片的区块计数
	receivedShards     map[uint64]bool   // 当前轮次中已收到区块的分片
	blockHeightLock    sync.Mutex        // 保护区块高度的锁

	// 新增：broker 动态管理相关
	brokerAddressPool []string                  // 所有 broker 地址池
	brokerEvents      map[uint64][]*BrokerEvent // epoch -> 事件列表

	// 新增：broker 操作记录
	brokerOperationLogPath string      // broker 操作日志文件路径
	brokerOpLogFile        *os.File    // 日志文件句柄
	brokerOpLogWriter      *csv.Writer // CSV writer
	brokerOpLogLock        sync.Mutex  // 保护日志写入
	// 区块记录相关
	blockLogPath   string      // 区块日志文件路径
	blockLogFile   *os.File    // 日志文件句柄
	blockLogWriter *csv.Writer // CSV writer
	blockLogLock   sync.Mutex  // 保护日志写入

	// 新增：性能监控
	lockWaitTime     []time.Duration // 记录每次获取锁的等待时间
	blockProcessTime []time.Duration // 记录每次处理区块的时间
	monitorLock      sync.Mutex

	// ===== 新增：详细计时字段 =====
	balanceUpdateTime []time.Duration
	eventTime         []time.Duration
	confirmTime       []time.Duration
	checkTime         []time.Duration
	exitCheckTime     []time.Duration
	addResultTime     []time.Duration
	recordBlockTime   []time.Duration
	// ===============================

	// ===== 新增: dealTxByBroker 详细计时 =====
	dealTxTimings struct {
		// 各阶段等锁时间
		getBalanceLockWait      []time.Duration // GetActiveBrokerBalance 等锁
		b2eCallLockWait         []time.Duration // B2E 调用前等锁
		filterLockWait          []time.Duration // 过滤阶段等锁
		lockTokenLockWait       []time.Duration // lockToken 等锁
		handleAllocatedLockWait []time.Duration // handleAllocatedTx 等锁

		// 各阶段执行时间
		getBalanceExecTime      []time.Duration // 获取余额快照耗时
		b2eCallExecTime         []time.Duration // B2E 算法执行耗时(已有,但重新记录)
		filterExecTime          []time.Duration // 过滤操作耗时
		lockTokenExecTime       []time.Duration // lockToken 执行耗时
		generateBATExecTime     []time.Duration // GenerateAllocatedTx 耗时
		handleAllocatedExecTime []time.Duration // handleAllocatedTx 耗时
		handleRawMagExecTime    []time.Duration // handleBrokerRawMag 耗时

		// 过滤统计
		filteredCount     []int // 被过滤掉的交易数
		lockTokenRejected []int // lockToken 中拒绝的交易数

		// 并发情况
		concurrentCalls []int // 同时进入 dealTxByBroker 的次数
	}
	dealTxTimingsLock sync.Mutex // 保护上述统计数据
}

// BrokerEvent broker 事件结构
type BrokerEvent struct {
	Operation     string // "join" or "exit"
	BrokerIndices []int  // broker 索引列表
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

	// 新增：初始化区块计数器
	shardBlockCount := make(map[uint64]uint64)
	receivedShards := make(map[uint64]bool)
	for i := uint64(0); i < uint64(params.ShardNum); i++ {
		shardBlockCount[i] = 0
		receivedShards[i] = false
	}

	bcm := &BrokerCommitteeMod_b2e{
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
		BAT_size:                  make([]float64, 0),
		currentBlockHeight:        0,
		shardBlockCount:           shardBlockCount,
		receivedShards:            receivedShards,
		brokerOperationLogPath:    params.DataWrite_path + "broker_operations.csv",
		brokerEvents:              make(map[uint64][]*BrokerEvent),
		blockLogPath:              params.DataWrite_path + "block_received_log.csv",
		// 在最后添加（blockLogPath 初始化后面）
		lockWaitTime:     make([]time.Duration, 0),
		blockProcessTime: make([]time.Duration, 0),
	}

	// ===== 新增: 初始化 dealTxTimings =====
	bcm.dealTxTimings.getBalanceLockWait = make([]time.Duration, 0)
	bcm.dealTxTimings.b2eCallLockWait = make([]time.Duration, 0)
	bcm.dealTxTimings.filterLockWait = make([]time.Duration, 0)
	bcm.dealTxTimings.lockTokenLockWait = make([]time.Duration, 0)
	bcm.dealTxTimings.handleAllocatedLockWait = make([]time.Duration, 0)

	bcm.dealTxTimings.getBalanceExecTime = make([]time.Duration, 0)
	bcm.dealTxTimings.b2eCallExecTime = make([]time.Duration, 0)
	bcm.dealTxTimings.filterExecTime = make([]time.Duration, 0)
	bcm.dealTxTimings.lockTokenExecTime = make([]time.Duration, 0)
	bcm.dealTxTimings.generateBATExecTime = make([]time.Duration, 0)
	bcm.dealTxTimings.handleAllocatedExecTime = make([]time.Duration, 0)
	bcm.dealTxTimings.handleRawMagExecTime = make([]time.Duration, 0)

	bcm.dealTxTimings.filteredCount = make([]int, 0)
	bcm.dealTxTimings.lockTokenRejected = make([]int, 0)
	bcm.dealTxTimings.concurrentCalls = make([]int, 0)

	// 初始化区块日志
	bcm.initBlockLog()
	// 新增：读取所有 broker 地址到地址池
	bcm.loadBrokerAddressPool()

	// 新增：读取事件控制文件
	bcm.loadBrokerEvents()

	// 新增：初始化 broker 操作日志文件
	bcm.initBrokerOperationLog()
	return bcm

}

func (bcm *BrokerCommitteeMod_b2e) HandleOtherMessage([]byte) {}

func (bcm *BrokerCommitteeMod_b2e) fetchModifiedMap(key string) uint64 {
	return uint64(utils.Addr2Shard(key))
}

func (bcm *BrokerCommitteeMod_b2e) txSending(txlist []*core.Transaction) {
	// the txs will be sent

	startTime := time.Now()

	sendToShard := make(map[uint64][]*core.Transaction)

	txNum := 0
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
				txNum += len(sendToShard[sid])
				go networks.TcpDial(send_msg, bcm.IpNodeTable[sid][0])
			}
			sendToShard = make(map[uint64][]*core.Transaction)
			//time.Sleep(time.Second)
		} //发送到源分片
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
	duration := time.Since(startTime)
	if duration > 100*time.Millisecond {
		fmt.Printf("[TxSending] 警告：发送%d笔交易耗时%v\n", len(txlist), duration)
	}
	fmt.Printf("Send %d tx\n", txNum)

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
func max[T uint64](a, b T) T {
	if a > b {
		return a
	}
	return b
}
func (bcm *BrokerCommitteeMod_b2e) saveMonitorStats() {
	dirpath := params.DataWrite_path + "monitor/"
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Printf("警告: 创建监控目录失败: %v", err)
		return
	}

	targetPath := dirpath + "lock_wait_stats_detailed.csv"
	file, err := os.Create(targetPath)
	if err != nil {
		log.Printf("警告: 创建监控文件失败: %v", err)
		return
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	// ===== 修改：包含所有详细字段 =====
	w.Write([]string{
		"Index",
		"WaitTime(ms)",
		"TotalProcessTime(ms)",
		"BalanceUpdate(ms)",
		"EventExec(ms)",
		"CreateConfirm(ms)",
		"CheckBroker(ms)",
		"ExitCheck(ms)",
		"AddResult(ms)",
		"RecordBlock(ms)",
	})

	for i := 0; i < len(bcm.lockWaitTime); i++ {
		w.Write([]string{
			strconv.Itoa(i),
			strconv.FormatFloat(float64(bcm.lockWaitTime[i].Microseconds())/1000.0, 'f', 3, 64),
			strconv.FormatFloat(float64(bcm.blockProcessTime[i].Microseconds())/1000.0, 'f', 3, 64),
			strconv.FormatFloat(float64(bcm.balanceUpdateTime[i].Microseconds())/1000.0, 'f', 3, 64),
			strconv.FormatFloat(float64(bcm.eventTime[i].Microseconds())/1000.0, 'f', 3, 64),
			strconv.FormatFloat(float64(bcm.confirmTime[i].Microseconds())/1000.0, 'f', 3, 64),
			strconv.FormatFloat(float64(bcm.checkTime[i].Microseconds())/1000.0, 'f', 3, 64),
			strconv.FormatFloat(float64(bcm.exitCheckTime[i].Microseconds())/1000.0, 'f', 3, 64),
			strconv.FormatFloat(float64(bcm.addResultTime[i].Microseconds())/1000.0, 'f', 3, 64),
			strconv.FormatFloat(float64(bcm.recordBlockTime[i].Microseconds())/1000.0, 'f', 3, 64),
		})
	}

	fmt.Printf("[监控] 详细性能数据已保存到: %s\n", targetPath)
}
func (bcm *BrokerCommitteeMod_b2e) HandleBlockInfo(b *message.BlockInfoMsg) {
	waitStart := time.Now()

	bcm.sl.Slog.Printf("received from shard %d in height %d.\n", b.SenderShardID,
		bcm.shardBlockCount[b.SenderShardID]+1)

	bcm.shardBlockCount[b.SenderShardID]++

	if b.BlockBodyLength == 0 {
		return
	}

	txs := make([]*core.Transaction, 0)
	txs = append(txs, b.Broker1Txs...)
	txs = append(txs, b.Broker2Txs...)

	bcm.brokerModuleLock.Lock()
	waitTime := time.Since(waitStart)
	bcm.brokerBalanceLock.Lock()
	processStart := time.Now()

	// ===== 计时点1 =====
	t1 := time.Now()
	// ===================

	bcm.currentBlockHeight++
	fmt.Printf("[BrokerHeight] 区块高度为 %d\n", bcm.currentBlockHeight)
	println("block length is ", len(b.ExcutedTxs))
	targetBroker := "32be343b94f860124dc4fee278fdcbd38c102d88"

	for _, tx := range b.Broker1Txs {
		if tx.Recipient == targetBroker {
			rSid := bcm.fetchModifiedMap(tx.FinalRecipient)
			fmt.Printf("[追踪-HandleBlock] 区块 %d, Broker %s 在分片 %d 释放 %s\n",
				bcm.currentBlockHeight, targetBroker[:8], rSid, tx.Value.String())
		}
		brokeraddress, sSid, rSid := tx.Recipient, bcm.fetchModifiedMap(tx.OriginalSender), bcm.fetchModifiedMap(tx.FinalRecipient)
		bcm.broker.LockBalance[brokeraddress][rSid].Sub(bcm.broker.LockBalance[brokeraddress][rSid], tx.Value)
		bcm.broker.BrokerBalance[brokeraddress][sSid].Add(bcm.broker.BrokerBalance[brokeraddress][sSid], tx.Value)
		fee := new(big.Float).SetInt64(tx.Fee.Int64())
		fee = fee.Mul(fee, bcm.broker.Brokerage)
		bcm.broker.ProfitBalance[brokeraddress][sSid].Add(bcm.broker.ProfitBalance[brokeraddress][sSid], fee)
	}

	// ===== 计时点2 =====
	balanceUpdateTime := time.Since(t1)
	t2 := time.Now()
	// ===================

	bcm.executeBrokerEvents(bcm.currentBlockHeight)

	// ===== 计时点3 =====
	eventTime := time.Since(t2)
	//t3 := time.Now()
	// ===================

	brokersBefore := make(map[string]bool)
	for _, addr := range bcm.broker.BrokerAddress {
		brokersBefore[addr] = true
	}

	bcm.brokerBalanceLock.Unlock()
	bcm.brokerModuleLock.Unlock()
	//unlockTime := time.Since(t3) // 记录解锁和准备工作的时间

	// ===== 计时点4 =====
	t4 := time.Now()
	// ===================

	bcm.createConfirm(txs)

	// ===== 计时点5 =====
	confirmTime := time.Since(t4)
	t5 := time.Now()
	// ===================

	bcm.broker.CheckAndProcessUnboundingBrokers()

	// ===== 计时点6 =====
	checkTime := time.Since(t5)
	t6 := time.Now()
	// ===================

	for addr := range brokersBefore {
		if !bcm.broker.IsBroker(addr) {
			brokerIndex := bcm.findBrokerIndex(addr)
			shardNum := uint64(utils.Addr2Shard(addr))
			shardHeight := bcm.shardBlockCount[shardNum]
			bcm.recordBrokerOperation(shardNum, shardHeight, bcm.currentBlockHeight, "exit_complete", brokerIndex, addr)
			fmt.Printf("[BrokerExit] 区块 %d: 地址 %s 完成退出\n", bcm.currentBlockHeight, addr)
		}
	}

	// ===== 计时点7 =====
	exitCheckTime := time.Since(t6)
	t7 := time.Now()
	// ===================

	bcm.add_result()

	// ===== 计时点8 =====
	addResultTime := time.Since(t7)
	t8 := time.Now()
	// ===================

	bcm.recordBlockInfo(b)

	// ===== 计时点9 =====
	recordBlockTime := time.Since(t8)
	// ===================

	processTime := time.Since(processStart)

	// ===== 保存所有计时数据（不打印） =====
	bcm.monitorLock.Lock()
	bcm.lockWaitTime = append(bcm.lockWaitTime, waitTime)
	bcm.blockProcessTime = append(bcm.blockProcessTime, processTime)
	bcm.balanceUpdateTime = append(bcm.balanceUpdateTime, balanceUpdateTime)
	bcm.eventTime = append(bcm.eventTime, eventTime)
	bcm.confirmTime = append(bcm.confirmTime, confirmTime)
	bcm.checkTime = append(bcm.checkTime, checkTime)
	bcm.exitCheckTime = append(bcm.exitCheckTime, exitCheckTime)
	bcm.addResultTime = append(bcm.addResultTime, addResultTime)
	bcm.recordBlockTime = append(bcm.recordBlockTime, recordBlockTime)
	bcm.monitorLock.Unlock()
	// ========================================
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

	relay_txs := make(map[uint64][]*core.Transaction)

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
				// relay tx
				tx.IsRelay = true
				relay_txs[sSid] = append(relay_txs[sSid], tx)
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
	println("len_brokerRecordMegs", len(brokerRecordMegs)) // record new injecting tx nums

	for shardID, txs := range relay_txs {
		if len(txs) == 0 {
			continue
		}
		relayTxs := message.InjectTxs{
			Txs:       txs,
			ToShardID: shardID,
		}
		itByte, err := json.Marshal(relayTxs)
		if err != nil {
			log.Printf("序列化错误: %v", err)
			continue
		}
		send_msg := message.MergeMessage(message.CInject, itByte)
		go networks.TcpDial(send_msg, bcm.IpNodeTable[shardID][0])
	}

	// // 新增：记录交易数量
	// transactionCount := len(brokerRecordMegs)
	// bcm.totalB2ETransactions += transactionCount

	// 新增：记录epoch brokermsg
	// bcm.epochB2ETransactions = append(bcm.epochB2ETransactions, transactionCount)

	// 新增：测量B2E函数执行时间
	// 调用 B2E
	// ===== 阶段1: 获取活跃 broker 余额 =====
	lockWaitStart := time.Now()
	bcm.brokerBalanceLock.Lock()
	getBalanceLockWait := time.Since(lockWaitStart)

	execStart := time.Now()
	activeBrokerBalance := bcm.broker.GetActiveBrokerBalance()
	getBalanceExecTime := time.Since(execStart)

	bcm.brokerBalanceLock.Unlock()

	// 记录统计
	bcm.recordDealTxTiming("getBalance", getBalanceLockWait, getBalanceExecTime, 0, 0)

	// 统计可用余额
	totalAvailable := big.NewInt(0)
	brokerWithBalance := 0
	for brokerAddr, shardBalances := range activeBrokerBalance {
		brokerTotal := big.NewInt(0)
		for _, balance := range shardBalances {
			brokerTotal.Add(brokerTotal, balance)
		}
		totalAvailable.Add(totalAvailable, brokerTotal)
		if brokerTotal.Cmp(big.NewInt(0)) > 0 {
			brokerWithBalance++
		}

		// 打印前10个broker的余额情况
		if brokerWithBalance < 10 {
			fmt.Printf("[BrokerBalance] Broker %s: 总余额=%s\n",
				brokerAddr, brokerTotal.String())
		}
	}

	fmt.Printf("[B2E Stats] Active broker数: %d, 有余额的broker: %d, 总可用余额: %s\n",
		len(activeBrokerBalance), brokerWithBalance, totalAvailable.String())

	b2eStart := time.Now()
	alloctedBrokerRawMegs, restBrokerRawMeg := Broker2Earn.B2E(brokerRawMegs, activeBrokerBalance)
	b2eExecTime := time.Since(b2eStart)

	bcm.recordDealTxTiming("b2e", 0, b2eExecTime, 0, 0)

	fmt.Printf("[B2E Result] 成功分配: %d笔, 失败: %d笔\n",
		len(alloctedBrokerRawMegs), len(restBrokerRawMeg))

	// 在 B2E 调用后
	if len(restBrokerRawMeg) > 0 {
		fmt.Printf("[分析] B2E 未能分配 %d 笔交易，原因分析：\n", len(restBrokerRawMeg))

		// 统计交易金额分布
		smallTx := 0  // < 1000
		mediumTx := 0 // 1000-10000
		largeTx := 0  // > 10000

		for _, msg := range restBrokerRawMeg {
			value := msg.Tx.Value.Int64()
			if value < 1000 {
				smallTx++
			} else if value < 10000 {
				mediumTx++
			} else {
				largeTx++
			}
		}

		fmt.Printf("  交易金额分布: 小额=%d, 中额=%d, 大额=%d\n", smallTx, mediumTx, largeTx)
		fmt.Printf("  Active broker 总余额: %s\n", totalAvailable.String())

		// 检查是否有 broker 余额足够
		hasEnoughBalance := false
		for _, shardBalances := range activeBrokerBalance {
			brokerTotal := big.NewInt(0)
			for _, balance := range shardBalances {
				brokerTotal.Add(brokerTotal, balance)
			}

			if len(restBrokerRawMeg) > 0 && brokerTotal.Cmp(restBrokerRawMeg[0].Tx.Value) > 0 {
				hasEnoughBalance = true
				break
			}
		}

		fmt.Printf("  至少一个 broker 余额足够: %v\n", hasEnoughBalance)
	}
	// 新增：保存执行时间
	bcm.b2eExecutionTimes = append(bcm.b2eExecutionTimes, b2eExecTime)
	bcm.totalB2EIterations++

	//alloctedBrokerRawMegs, restBrokerRawMeg := Broker2Earn.B2E(brokerRawMegs, bcm.broker.BrokerBalance)
	for _, item := range restBrokerRawMeg {
		bcm.restBrokerRawMegPool = append(bcm.restBrokerRawMegPool, item)
	}
	println("len_alloctedBrokerRawMegs", len(alloctedBrokerRawMegs))
	//bcm.brokerBalanceLock.Unlock()

	// ===== 阶段3: 过滤 unbonding broker =====
	filterStart := time.Now()
	validAllocations := make([]*message.BrokerRawMeg, 0)
	filteredCount := 0

	for _, msg := range alloctedBrokerRawMegs {
		if bcm.broker.IsUnbounding(msg.Broker) || !bcm.broker.IsBroker(msg.Broker) {
			restBrokerRawMeg = append(restBrokerRawMeg, msg)
			filteredCount++
		} else {
			validAllocations = append(validAllocations, msg)
		}
	}
	alloctedBrokerRawMegs = validAllocations
	filterExecTime := time.Since(filterStart)

	bcm.recordDealTxTiming("filter", 0, filterExecTime, filteredCount, 0)
	// ================================================

	bcm.len_alloctedBrokerRawMegs = append(bcm.len_alloctedBrokerRawMegs, len(alloctedBrokerRawMegs))
	// ===== 阶段4: GenerateAllocatedTx =====
	genBATStart := time.Now()
	allocatedTxs := bcm.GenerateAllocatedTx(alloctedBrokerRawMegs)
	genBATExecTime := time.Since(genBATStart)

	bcm.recordDealTxTiming("generateBAT", 0, genBATExecTime, 0, 0)
	aByte, err := json.Marshal(allocatedTxs)
	if err != nil {
		log.Panic(err)
	}
	// 新增：记录BAT大小
	batSize := float64(len(aByte)) / 1024.0      // 转换为KB
	bcm.BAT_size = append(bcm.BAT_size, batSize) // 计算得到了 BAT 大小（KB）

	if len(alloctedBrokerRawMegs) != 0 {
		// ===== 阶段5: handleAllocatedTx =====
		handleAllocatedStart := time.Now()
		bcm.handleAllocatedTx(allocatedTxs)
		handleAllocatedExecTime := time.Since(handleAllocatedStart)
		bcm.recordDealTxTiming("handleAllocated", 0, handleAllocatedExecTime, 0, 0)

		// ===== 阶段6: lockToken =====
		lockTokenStart := time.Now()
		rejectedCount := bcm.lockToken(alloctedBrokerRawMegs) // 修改返回值
		lockTokenExecTime := time.Since(lockTokenStart)
		bcm.recordDealTxTiming("lockToken", 0, lockTokenExecTime, 0, rejectedCount)

		// ===== 阶段7: handleBrokerRawMag =====
		handleRawMagStart := time.Now()
		bcm.handleBrokerRawMag(alloctedBrokerRawMegs)
		handleRawMagExecTime := time.Since(handleRawMagStart)
		bcm.recordDealTxTiming("handleRawMag", 0, handleRawMagExecTime, 0, 0)
	}

	bcm.SaveB2ETimeStats(relay_txs, itxs)
	return itxs
}

func (bcm *BrokerCommitteeMod_b2e) lockToken(alloctedBrokerRawMegs []*message.BrokerRawMeg) int {
	targetBroker := "32be343b94f860124dc4fee278fdcbd38c102d88"

	// ===== 测量等锁时间 =====
	lockWaitStart := time.Now()
	bcm.brokerBalanceLock.Lock()
	lockWaitDuration := time.Since(lockWaitStart)

	// 记录等锁时间
	bcm.dealTxTimingsLock.Lock()
	bcm.dealTxTimings.lockTokenLockWait = append(bcm.dealTxTimings.lockTokenLockWait, lockWaitDuration)
	bcm.dealTxTimingsLock.Unlock()

	rejectedCount := 0
	rejectedBrokers := make(map[string]int)

	for _, brokerRawMeg := range alloctedBrokerRawMegs {
		tx := brokerRawMeg.Tx
		brokerAddress := brokerRawMeg.Broker

		if bcm.broker.IsUnbounding(brokerAddress) {
			bcm.restBrokerRawMegPool = append(bcm.restBrokerRawMegPool, brokerRawMeg)
			rejectedCount++
			rejectedBrokers[brokerAddress]++
			continue
		}

		rSid := bcm.fetchModifiedMap(tx.Recipient)

		// ===== 在实际锁定前打印 =====
		if brokerAddress == targetBroker {
			fmt.Printf("[追踪-LockToken] Broker %s 在分片 %d 锁定 %s\n",
				targetBroker[:8], rSid, brokerRawMeg.Tx.Value.String())
		}

		bcm.broker.LockBalance[brokerAddress][rSid].Add(bcm.broker.LockBalance[brokerAddress][rSid], tx.Value)
		bcm.broker.BrokerBalance[brokerAddress][rSid].Sub(bcm.broker.BrokerBalance[brokerAddress][rSid], tx.Value)
	}

	if rejectedCount > 0 {
		fmt.Printf("[LockToken] 拒绝了 %d 笔交易，涉及 unbonding broker: %v\n",
			rejectedCount, rejectedBrokers)
	}

	bcm.brokerBalanceLock.Unlock()
	return rejectedCount // ===== 返回拒绝数量 =====
}

//	func (bcm *BrokerCommitteeMod_b2e) lockToken(alloctedBrokerRawMegs []*message.BrokerRawMeg) {
//		targetBroker := "32be343b94f860124dc4fee278fdcbd38c102d88"
//
//		bcm.brokerBalanceLock.Lock()
//
//		rejectedCount := 0
//		rejectedBrokers := make(map[string]int)
//		for _, brokerRawMeg := range alloctedBrokerRawMegs {
//
//			tx := brokerRawMeg.Tx
//			brokerAddress := brokerRawMeg.Broker
//			// if broker is unbonding, cancel BAT
//			if bcm.broker.IsUnbounding(brokerAddress) {
//				bcm.restBrokerRawMegPool = append(bcm.restBrokerRawMegPool, brokerRawMeg)
//				rejectedCount++
//				rejectedBrokers[brokerAddress]++
//				continue
//			}
//			rSid := bcm.fetchModifiedMap(tx.Recipient)
//
//			fmt.Printf("[追踪-LockToken] Broker %s 在分片 %d 锁定 %s\n",
//				targetBroker[:8], rSid, brokerRawMeg.Tx.Value.String())
//			bcm.broker.LockBalance[brokerAddress][rSid].Add(bcm.broker.LockBalance[brokerAddress][rSid], tx.Value)
//			bcm.broker.BrokerBalance[brokerAddress][rSid].Sub(bcm.broker.BrokerBalance[brokerAddress][rSid], tx.Value)
//		}
//		if rejectedCount > 0 {
//			fmt.Printf("[LockToken] 拒绝了 %d 笔交易，涉及 unbonding broker: %v\n",
//				rejectedCount, rejectedBrokers)
//		}
//		bcm.brokerBalanceLock.Unlock()
//	}
func (bcm *BrokerCommitteeMod_b2e) handleAllocatedTx(alloctedTx map[uint64][]*core.Transaction) {
	// ===== 测量等锁时间 =====
	lockWaitStart := time.Now()
	bcm.brokerBalanceLock.Lock()
	lockWaitDuration := time.Since(lockWaitStart)

	// 记录等锁时间
	bcm.dealTxTimingsLock.Lock()
	bcm.dealTxTimings.handleAllocatedLockWait = append(bcm.dealTxTimings.handleAllocatedLockWait, lockWaitDuration)
	bcm.dealTxTimingsLock.Unlock()

	for shardId, txs := range alloctedTx {
		for _, tx := range txs {

			// judge whether BATs
			if tx.Sender != tx.Recipient && !tx.IsAllocatedSender && !tx.IsAllocatedRecipent {
				continue
			}
			// if broker is unbonding, cancel BAT
			if bcm.broker.IsUnbounding(tx.Sender) {
				continue
			}
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
		//time.Sleep(time.Second)
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
	tx1s := make(map[uint64][]*core.Transaction, 0)
	for _, brokerType1Meg := range brokerType1Megs {
		ctx := brokerType1Meg.RawMeg.Tx
		tx1 := core.NewTransaction(ctx.Sender, brokerType1Meg.Broker, ctx.Value, ctx.Nonce, ctx.Fee)
		tx1.OriginalSender = ctx.Sender
		tx1.FinalRecipient = ctx.Recipient
		tx1.RawTxHash = make([]byte, len(ctx.TxHash))
		copy(tx1.RawTxHash, ctx.TxHash)
		ssid := uint64(utils.Addr2Shard(ctx.Sender))
		tx1s[ssid] = append(tx1s[ssid], tx1)
		confirm1 := &message.Mag1Confirm{
			RawMeg:  brokerType1Meg.RawMeg,
			Tx1Hash: tx1.TxHash,
		}
		bcm.brokerModuleLock.Lock()
		bcm.brokerConfirm1Pool[string(tx1.TxHash)] = confirm1
		bcm.brokerModuleLock.Unlock()
	}

	for shardId, txs := range tx1s {

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
	}
	fmt.Println("BrokerType1Mes received by shard,  add brokerTx1 len ", len(tx1s))
}

func (bcm *BrokerCommitteeMod_b2e) handleBrokerType2Mes(brokerType2Megs []*message.BrokerType2Meg) {
	tx2s := make(map[uint64][]*core.Transaction, 0)
	for _, mes := range brokerType2Megs {
		ctx := mes.RawMeg.Tx
		tx2 := core.NewTransaction(mes.Broker, ctx.Recipient, ctx.Value, ctx.Nonce, ctx.Fee)
		tx2.OriginalSender = ctx.Sender
		tx2.FinalRecipient = ctx.Recipient
		tx2.RawTxHash = make([]byte, len(ctx.TxHash))
		copy(tx2.RawTxHash, ctx.TxHash)

		rsid := uint64(utils.Addr2Shard(ctx.Recipient))
		tx2s[rsid] = append(tx2s[rsid], tx2)

		confirm2 := &message.Mag2Confirm{
			RawMeg:  mes.RawMeg,
			Tx2Hash: tx2.TxHash,
		}
		//bcm.brokerModuleLock.Lock()
		bcm.brokerConfirm2Pool[string(tx2.TxHash)] = confirm2
		//bcm.brokerModuleLock.Unlock()
	}

	for shardId, txs := range tx2s {

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
	}
	//go bcm.txSending(tx2s)
	fmt.Println("broker tx2 add to pool len ", len(tx2s))
}

// to get the digest of rawMeg
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

		// if broker is unbonding, cancel broker1Txs
		brokerAddress := meg.Broker
		if bcm.broker.IsUnbounding(brokerAddress) {
			continue
		}
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
		//for shardId, balance := range shardMap {
		for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
			a += shardMap[sid].String() + ","
			b += bcm.broker.LockBalance[brokerAddress][sid].String() + ","
			c += bcm.broker.ProfitBalance[brokerAddress][sid].String() + ","
		}
		a += "\n"
		b += "\n"
		c += "\n"
		bcm.Result_lockBalance[brokerAddress] = append(bcm.Result_lockBalance[brokerAddress], b)
		bcm.Result_brokerBalance[brokerAddress] = append(bcm.Result_brokerBalance[brokerAddress], a)
		bcm.Result_Profit[brokerAddress] = append(bcm.Result_Profit[brokerAddress], c)

		// 实时写入到文件 - 只写入最新的一行数据
		//targetPath0 := dirpath + brokerAddress + "_lockBalance.csv"
		//targetPath1 := dirpath + brokerAddress + "_brokerBalance.csv"
		//targetPath2 := dirpath + brokerAddress + "_Profit.csv"

		// 直接写入最新的行，而不是整个数组
		//bcm.writeLatestRow(targetPath0, []string{b})
		//bcm.writeLatestRow(targetPath1, []string{a})
		//bcm.writeLatestRow(targetPath2, []string{c})
	}

	// 新增：关闭 broker 操作日志文件
	//if bcm.brokerOpLogWriter != nil {
	//	bcm.brokerOpLogWriter.Flush()
	//}
	//if bcm.brokerOpLogFile != nil {
	//	bcm.brokerOpLogFile.Close()
	//	fmt.Printf("[BrokerLog] 操作日志文件已关闭\n")
	//}

	// ===== 新增：保存监控数据 =====
	bcm.saveMonitorStats()
	// =============================
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
func (bcm *BrokerCommitteeMod_b2e) SaveB2ETimeStats(relay_txs map[uint64][]*core.Transaction, itxs []*core.Transaction) {
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
		w.Write([]string{"Iteration", "ExecutionTime(ms)", "epoch_tx_injected", "rest_BrokerRawMegPoolLen", "notHandleCtxByBroker", "len_alloctedBrokerRawMegs", "relayTxnum", "itxNum"})
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

	legthOfRelayTx := 0
	for _, s := range relay_txs {
		legthOfRelayTx += len(s)
	}

	row := []string{
		strconv.Itoa(iteration),
		strconv.FormatFloat(float64(latestTime.Microseconds())/1000.0, 'f', 6, 64),
		strconv.Itoa(bcm.txlen[iteration-1]),
		strconv.Itoa(bcm.rest_BrokerRawMegPoolLen[iteration-1]),
		strconv.Itoa(bcm.notHandleCtxByBroker[iteration-1]),
		strconv.Itoa(bcm.len_alloctedBrokerRawMegs[iteration-1]),
		strconv.Itoa(legthOfRelayTx),
		strconv.Itoa(len(itxs)),
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
	// 新增：关闭 broker 操作日志文件
	//if bcm.brokerOpLogWriter != nil {
	//	bcm.brokerOpLogWriter.Flush()
	//}
	//if bcm.brokerOpLogFile != nil {
	//	bcm.brokerOpLogFile.Close()
	//	fmt.Printf("[BrokerLog] 操作日志文件已关闭\n")
	//}
	bcm.saveMonitorStats()
	// ===== 新增: 保存 dealTxByBroker 详细计时 =====
	bcm.SaveDealTxTimings()
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

// checkAndRemoveBrokers 检查并移除可以退出的 broker
func (bcm *BrokerCommitteeMod_b2e) checkAndRemoveBrokers() {
	// 调用 broker 的检查函数（传入当前区块高度）
	bcm.broker.CheckAndProcessUnboundingBrokers()
}

// TriggerBrokerExit 触发 broker 退出流程
func (bcm *BrokerCommitteeMod_b2e) TriggerBrokerExit(brokerAddress string) error {
	bcm.blockHeightLock.Lock()
	currentHeight := bcm.currentBlockHeight
	bcm.blockHeightLock.Unlock()

	// 触发 broker 退出
	err := bcm.broker.InitiateUnbounding(brokerAddress)
	if err != nil {
		return fmt.Errorf("触发退出失败: %v", err)
	}

	fmt.Printf("[Supervisor] Broker %s 在区块 %d 开始退出流程\n", brokerAddress, currentHeight)
	return nil
}

// GetCurrentBlockHeight 获取当前全局区块高度
func (bcm *BrokerCommitteeMod_b2e) GetCurrentBlockHeight() uint64 {
	bcm.blockHeightLock.Lock()
	defer bcm.blockHeightLock.Unlock()
	return bcm.currentBlockHeight
}

// loadBrokerAddressPool 读取所有 broker 地址到地址池
func (bcm *BrokerCommitteeMod_b2e) loadBrokerAddressPool() {
	filePath := `./broker/broker`
	readFile, err := os.Open(filePath)
	if err != nil {
		log.Printf("警告: 无法打开 broker 地址文件: %v", err)
		return
	}
	defer readFile.Close()

	bcm.brokerAddressPool = make([]string, 0)
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		address := strings.TrimSpace(fileScanner.Text())
		if address != "" {
			bcm.brokerAddressPool = append(bcm.brokerAddressPool, address)
		}
	}

	fmt.Printf("[BrokerPool] 成功加载 %d 个 broker 地址\n", len(bcm.brokerAddressPool))
}

// loadBrokerEvents 读取 broker 事件控制文件
func (bcm *BrokerCommitteeMod_b2e) loadBrokerEvents() {
	filePath := `./broker/broker_events.csv`
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("警告: 无法打开事件文件: %v，将不会有动态 broker 加入/退出", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 跳过表头
	_, err = reader.Read()
	if err != nil {
		log.Printf("警告: 读取事件文件表头失败: %v", err)
		return
	}

	eventCount := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("警告: 读取事件文件行失败: %v", err)
			continue
		}

		// 解析：block_height, operation, broker_indices
		if len(record) < 3 {
			log.Printf("警告: 事件文件格式错误，跳过该行: %v", record)
			continue
		}

		blockHeight, err := strconv.ParseUint(record[0], 10, 64)
		if err != nil {
			log.Printf("警告: 无效的区块高度: %s", record[0])
			continue
		}

		operation := strings.TrimSpace(record[1])
		if operation != "join" && operation != "exit" {
			log.Printf("警告: 无效的操作类型: %s", operation)
			continue
		}

		// 解析索引列表（逗号分隔）
		indicesStr := strings.TrimSpace(record[2])
		indices := make([]int, 0)
		for _, indexStr := range strings.Split(indicesStr, ",") {
			indexStr = strings.TrimSpace(indexStr)
			if indexStr == "" {
				continue
			}
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				log.Printf("警告: 无效的索引: %s", indexStr)
				continue
			}
			indices = append(indices, index)
		}

		if len(indices) == 0 {
			log.Printf("警告: 事件没有有效的 broker 索引")
			continue
		}

		// 创建事件
		event := &BrokerEvent{
			Operation:     operation,
			BrokerIndices: indices,
		}

		log.Printf("第%d区块,操作为 %s\n", blockHeight, operation)
		// 添加到对应 epoch 的事件列表
		bcm.brokerEvents[blockHeight] = append(bcm.brokerEvents[blockHeight], event)
		eventCount++
	}
	// 不加载
	//bcm.brokerEvents = make(map[uint64][]*BrokerEvent)
	fmt.Printf("[BrokerEvents] 成功加载 %d 个事件，涉及 %d 个区块\n", eventCount, len(bcm.brokerEvents))
}

// 执行指定区块高度的所有事件
func (bcm *BrokerCommitteeMod_b2e) executeBrokerEvents(blockHeight uint64) {
	events, exists := bcm.brokerEvents[blockHeight]
	if !exists || len(events) == 0 {
		return // 该区块没有事件
	}

	fmt.Printf("[BrokerEvents] 区块 %d: 开始执行 %d 个事件\n", blockHeight, len(events))

	for _, event := range events {
		switch event.Operation {
		case "join":
			bcm.executeBrokerJoin(event, blockHeight)
		case "exit":
			bcm.executeBrokerExit(event, blockHeight)
		default:
			fmt.Printf("[BrokerEvents] 未知操作类型: %s\n", event.Operation)
		}
	}
}

// 执行 broker 加入操作
func (bcm *BrokerCommitteeMod_b2e) executeBrokerJoin(event *BrokerEvent, blockHeight uint64) {
	successCount := 0

	// 使用默认初始余额
	initialBalance := params.Init_broker_Balance

	for _, index := range event.BrokerIndices {
		// 检查索引是否有效
		if index < 0 || index >= len(bcm.brokerAddressPool) {
			fmt.Printf("[BrokerJoin] 警告: 索引 %d 超出地址池范围 (0-%d)\n",
				index, len(bcm.brokerAddressPool)-1)
			continue
		}

		address := bcm.brokerAddressPool[index]

		// 调用 AddBroker
		err := bcm.broker.AddBroker(address, initialBalance)
		if err != nil {
			fmt.Printf("[BrokerJoin] 警告: 索引 %d 地址 %s 加入失败: %v\n",
				index, address, err)
			continue
		}

		// 新增：记录加入操作
		shardNum := uint64(utils.Addr2Shard(address))
		shardHeight := bcm.shardBlockCount[shardNum]

		bcm.recordBrokerOperation(shardNum, shardHeight, blockHeight, "join", index, address)

		successCount++
		fmt.Printf("[BrokerJoin] 区块 %d: 索引 %d 地址 %s 成功加入\n",
			blockHeight, index, address)
	}

	fmt.Printf("[BrokerJoin] 区块 %d: 成功加入 %d/%d 个 broker，当前活跃: %d\n",
		blockHeight, successCount, len(event.BrokerIndices),
		bcm.broker.GetActiveBrokerCount())
}

// executeBrokerExit 执行 broker 退出操作
func (bcm *BrokerCommitteeMod_b2e) executeBrokerExit(event *BrokerEvent, blockHeight uint64) {
	successCount := 0

	for _, index := range event.BrokerIndices {
		// 检查索引是否有效
		if index < 0 || index >= len(bcm.brokerAddressPool) {
			fmt.Printf("[BrokerExit] 警告: 索引 %d 超出地址池范围 (0-%d)\n",
				index, len(bcm.brokerAddressPool)-1)
			continue
		}

		address := bcm.brokerAddressPool[index]

		// 检查该 broker 是否在系统中
		if !bcm.broker.IsBroker(address) {
			fmt.Printf("[BrokerExit] 警告: 索引 %d 地址 %s 不在系统中\n",
				index, address)
			continue
		}

		// 调用 InitiateUnbounding
		err := bcm.broker.InitiateUnbounding(address)
		if err != nil {
			fmt.Printf("[BrokerExit] 警告: 索引 %d 地址 %s 退出失败: %v\n",
				index, address, err)
			continue
		}
		shardNum := uint64(utils.Addr2Shard(address))
		shardHeight := bcm.shardBlockCount[shardNum]

		// 新增：记录发起退出操作
		bcm.recordBrokerOperation(shardNum, shardHeight, blockHeight, "exit_start", index, address)

		successCount++
		fmt.Printf("[BrokerExit] 区块 %d: 索引 %d 地址 %s 开始退出流程\n",
			blockHeight, index, address)
	}

	fmt.Printf("[BrokerExit] 区块 %d: 成功触发 %d/%d 个 broker 退出，当前活跃: %d\n",
		blockHeight, successCount, len(event.BrokerIndices),
		bcm.broker.GetActiveBrokerCount())
}

// initBrokerOperationLog 初始化 broker 操作日志文件
func (bcm *BrokerCommitteeMod_b2e) initBrokerOperationLog() {
	// 创建目录
	dirpath := params.DataWrite_path
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Printf("警告: 创建日志目录失败: %v", err)
		return
	}

	// 创建或打开文件
	file, err := os.Create(bcm.brokerOperationLogPath)
	if err != nil {
		log.Printf("警告: 创建 broker 操作日志文件失败: %v", err)
		return
	}

	bcm.brokerOpLogFile = file
	bcm.brokerOpLogWriter = csv.NewWriter(file)

	// 写入表头
	header := []string{"block_height", "ShardID", "ShardHeight", "operation", "broker_index", "broker_address", "timestamp"}
	bcm.brokerOpLogWriter.Write(header)
	bcm.brokerOpLogWriter.Flush()
	file.Sync()

	fmt.Printf("[BrokerLog] 操作日志文件创建: %s\n", bcm.brokerOperationLogPath)
}

// 记录 broker 操作到 CSV（实时写入）
func (bcm *BrokerCommitteeMod_b2e) recordBrokerOperation(shardNum uint64, shardHeight uint64, blockHeight uint64, operation string, brokerIndex int, brokerAddress string) {
	// ===== 新增：调试日志 =====
	//fmt.Printf("[调试-记录函数] 被调用: 区块=%d, 操作=%s, 地址=%s\n",
	//	blockHeight, operation, brokerAddress)
	// ==========================

	if bcm.brokerOpLogWriter == nil {
		// ===== 新增：警告日志 =====
		//fmt.Printf("[错误-记录函数] brokerOpLogWriter 为 nil，无法记录！\n")
		// ==========================
		return
	}

	bcm.brokerOpLogLock.Lock()
	defer bcm.brokerOpLogLock.Unlock()

	timestamp := time.Now().UnixMilli()

	row := []string{
		strconv.FormatUint(blockHeight, 10),
		strconv.FormatUint(shardNum, 10),
		strconv.FormatUint(shardHeight, 10),
		operation,
		strconv.Itoa(brokerIndex),
		brokerAddress,
		strconv.FormatInt(timestamp, 10),
	}

	// ===== 修改：检查错误 =====
	err := bcm.brokerOpLogWriter.Write(row)
	if err != nil {
		fmt.Printf("[错误-记录函数] 写入失败: %v\n", err)
		return
	}

	bcm.brokerOpLogWriter.Flush()

	err = bcm.brokerOpLogFile.Sync()
	if err != nil {
		fmt.Printf("[错误-记录函数] Sync失败: %v\n", err)
	} else {
		fmt.Printf("[成功-记录函数] 已记录: 区块=%d, 操作=%s\n", blockHeight, operation)
	}
	// ==========================
}

// findBrokerIndex 在地址池中查找 broker 的索引
func (bcm *BrokerCommitteeMod_b2e) findBrokerIndex(address string) int {
	for i, addr := range bcm.brokerAddressPool {
		if addr == address {
			return i
		}
	}
	return -1 // 未找到
}

// 初始化区块日志文件
func (bcm *BrokerCommitteeMod_b2e) initBlockLog() {
	// 创建目录
	dirpath := params.DataWrite_path
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Printf("警告: 创建区块日志目录失败: %v", err)
		return
	}

	// 创建或打开文件
	file, err := os.Create(bcm.blockLogPath)
	if err != nil {
		log.Printf("警告: 创建区块日志文件失败: %v", err)
		return
	}

	bcm.blockLogFile = file
	bcm.blockLogWriter = csv.NewWriter(file)

	// 写入表头
	header := []string{
		"global_block_height", // 全局区块高度
		"shard_id",            // 分片ID
		"shard_block_height",  // 该分片的区块高度
		"ExcutedTxs",          // 总交易数
		"Broker1Txs",          // 执行的交易数
		"Broker2Txs",          // broker1交易数
		"Relay1Txs",           // broker2交易数
		"Relay2TxNum",         // broker2交易数
		"AllocatedTxs",        // 跨片交易数（如果有）
		"timestamp",           // 时间戳
	}

	bcm.blockLogWriter.Write(header)
	bcm.blockLogWriter.Flush()
	file.Sync()

	fmt.Printf("[BlockLog] 区块日志文件创建: %s\n", bcm.blockLogPath)
}
func (bcm *BrokerCommitteeMod_b2e) recordBlockInfo(blockMsg *message.BlockInfoMsg) {
	if bcm.blockLogWriter == nil {
		return // 日志文件未初始化
	}

	bcm.blockLogLock.Lock()
	defer bcm.blockLogLock.Unlock()
	// 获取时间戳
	timestamp := time.Now().UnixMilli()

	// 构建CSV行
	row := []string{
		strconv.FormatUint(bcm.currentBlockHeight, 10),                      // 全局高度
		strconv.FormatUint(uint64(blockMsg.SenderShardID), 10),              // 分片ID
		strconv.FormatUint(bcm.shardBlockCount[blockMsg.SenderShardID], 10), // 分片高度
		strconv.Itoa(len(blockMsg.ExcutedTxs)),                              // 总交易数
		strconv.Itoa(len(blockMsg.Broker1Txs)),                              // 执行的交易数
		strconv.Itoa(len(blockMsg.Broker2Txs)),                              // 执行的交易数
		strconv.Itoa(len(blockMsg.Relay1Txs)),                               // 执行的交易数
		strconv.Itoa(int(blockMsg.Relay2TxNum)),                             // 执行的交易数
		strconv.Itoa(len(blockMsg.AllocatedTxs)),                            // 执行的交易数
		strconv.FormatInt(timestamp, 10),                                    // 时间戳
	}

	// 写入CSV
	bcm.blockLogWriter.Write(row)
	bcm.blockLogWriter.Flush()
	bcm.blockLogFile.Sync() // 立即刷新到磁盘
}

// recordDealTxTiming 记录 dealTxByBroker 各阶段的时间统计
func (bcm *BrokerCommitteeMod_b2e) recordDealTxTiming(stage string, lockWait, execTime time.Duration, filtered, rejected int) {
	bcm.dealTxTimingsLock.Lock()
	defer bcm.dealTxTimingsLock.Unlock()

	switch stage {
	case "getBalance":
		bcm.dealTxTimings.getBalanceLockWait = append(bcm.dealTxTimings.getBalanceLockWait, lockWait)
		bcm.dealTxTimings.getBalanceExecTime = append(bcm.dealTxTimings.getBalanceExecTime, execTime)
	case "b2e":
		bcm.dealTxTimings.b2eCallExecTime = append(bcm.dealTxTimings.b2eCallExecTime, execTime)
	case "filter":
		bcm.dealTxTimings.filterExecTime = append(bcm.dealTxTimings.filterExecTime, execTime)
		bcm.dealTxTimings.filteredCount = append(bcm.dealTxTimings.filteredCount, filtered)
	case "generateBAT":
		bcm.dealTxTimings.generateBATExecTime = append(bcm.dealTxTimings.generateBATExecTime, execTime)
	case "handleAllocated":
		bcm.dealTxTimings.handleAllocatedExecTime = append(bcm.dealTxTimings.handleAllocatedExecTime, execTime)
	case "lockToken":
		bcm.dealTxTimings.lockTokenExecTime = append(bcm.dealTxTimings.lockTokenExecTime, execTime)
		bcm.dealTxTimings.lockTokenRejected = append(bcm.dealTxTimings.lockTokenRejected, rejected)
	case "handleRawMag":
		bcm.dealTxTimings.handleRawMagExecTime = append(bcm.dealTxTimings.handleRawMagExecTime, execTime)
	}
}
func (bcm *BrokerCommitteeMod_b2e) SaveDealTxTimings() {
	dirpath := params.DataWrite_path + "b2e_execution_time/"
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Printf("警告: 创建目录失败: %v", err)
		return
	}

	targetPath := dirpath + "dealTx_detailed_timings.csv"
	file, err := os.Create(targetPath)
	if err != nil {
		log.Printf("警告: 创建文件失败: %v", err)
		log.Printf("警告: 创建文件失败: %v", err)
		return
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	// 写入表头
	w.Write([]string{
		"Iteration",
		"GetBalance_LockWait(ms)",
		"GetBalance_Exec(ms)",
		"B2E_Exec(ms)",
		"Filter_Exec(ms)",
		"Filtered_Count",
		"GenBAT_Exec(ms)",
		"HandleAllocated_LockWait(ms)",
		"HandleAllocated_Exec(ms)",
		"LockToken_LockWait(ms)",
		"LockToken_Exec(ms)",
		"LockToken_Rejected",
		"HandleRawMag_Exec(ms)",
	})

	bcm.dealTxTimingsLock.Lock()
	defer bcm.dealTxTimingsLock.Unlock()

	maxLen := len(bcm.dealTxTimings.b2eCallExecTime)

	for i := 0; i < maxLen; i++ {
		row := []string{
			strconv.Itoa(i + 1),
			formatDuration(bcm.dealTxTimings.getBalanceLockWait, i),
			formatDuration(bcm.dealTxTimings.getBalanceExecTime, i),
			formatDuration(bcm.dealTxTimings.b2eCallExecTime, i),
			formatDuration(bcm.dealTxTimings.filterExecTime, i),
			formatInt(bcm.dealTxTimings.filteredCount, i),
			formatDuration(bcm.dealTxTimings.generateBATExecTime, i),
			formatDuration(bcm.dealTxTimings.handleAllocatedLockWait, i),
			formatDuration(bcm.dealTxTimings.handleAllocatedExecTime, i),
			formatDuration(bcm.dealTxTimings.lockTokenLockWait, i),
			formatDuration(bcm.dealTxTimings.lockTokenExecTime, i),
			formatInt(bcm.dealTxTimings.lockTokenRejected, i),
			formatDuration(bcm.dealTxTimings.handleRawMagExecTime, i),
		}
		w.Write(row)
	}

	fmt.Printf("[DealTxTimings] 详细计时数据已保存到: %s\n", targetPath)
}

// 辅助函数: 格式化 time.Duration 为毫秒字符串
func formatDuration(slice []time.Duration, index int) string {
	if index >= len(slice) {
		return "0.000"
	}
	return strconv.FormatFloat(float64(slice[index].Microseconds())/1000.0, 'f', 3, 64)
}

// 辅助函数: 格式化 int 数组
func formatInt(slice []int, index int) string {
	if index >= len(slice) {
		return "0"
	}
	return strconv.Itoa(slice[index])
}
