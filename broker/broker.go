package broker

import (
	"blockEmulator/message"
	"blockEmulator/params"
	"bufio"
	"fmt"
	"math/big"
	"os"
)

type Broker struct {
	BrokerRawMegs  map[string]*message.BrokerRawMeg
	ChainConfig    *params.ChainConfig
	BrokerAddress  []string
	BrokerBalance  map[string]map[uint64]*big.Int
	LockBalance    map[string]map[uint64]*big.Int
	ProfitBalance  map[string]map[uint64]*big.Float
	RawTx2BrokerTx map[string][]string
	Unbounding     map[string]bool // 标记准备退出的 broker
	Brokerage      *big.Float
}

func (b *Broker) NewBroker(pcc *params.ChainConfig) {
	b.BrokerRawMegs = make(map[string]*message.BrokerRawMeg)
	b.RawTx2BrokerTx = make(map[string][]string)
	b.ChainConfig = pcc
	b.BrokerAddress = b.initBrokerAddr(params.BrokerNum)
	b.BrokerBalance = b.initBrokerBalance(params.Init_broker_Balance)
	b.LockBalance = b.initBrokerBalance(big.NewInt(0))
	b.ProfitBalance = b.initFloatBalance(big.NewFloat(0))
	b.Brokerage = big.NewFloat(params.Brokerage)
	b.Unbounding = make(map[string]bool)
}

func (b *Broker) IsBroker(address string) bool {
	for _, brokerAddress := range b.BrokerAddress {
		if brokerAddress == address {
			return true
		}
	}
	return false
}

func (b *Broker) initBrokerAddr(num int) []string {
	b.BrokerBalance = make(map[string]map[uint64]*big.Int)
	brokerAddress := make([]string, 0)
	filePath := "./broker/broker1000"
	readFile, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		address := fileScanner.Text()
		brokerAddress = append(brokerAddress, address)
		num--
		if num == 0 {
			break
		}
	}

	readFile.Close()
	return brokerAddress
}

func (b *Broker) initBrokerBalance(balance *big.Int) map[string]map[uint64]*big.Int {
	BrokerBalance := make(map[string]map[uint64]*big.Int)
	for _, address := range b.BrokerAddress {
		BrokerBalance[address] = make(map[uint64]*big.Int)
		shardBalance := new(big.Int).Div(new(big.Int).Set(balance), big.NewInt(int64(params.ShardNum)))
		for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
			BrokerBalance[address][sid] = new(big.Int).Set(shardBalance)
		}
	}
	return BrokerBalance
}

func (b *Broker) initFloatBalance(balance *big.Float) map[string]map[uint64]*big.Float {
	BrokerBalance := make(map[string]map[uint64]*big.Float)
	for _, address := range b.BrokerAddress {
		BrokerBalance[address] = make(map[uint64]*big.Float)
		for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
			BrokerBalance[address][sid] = new(big.Float).Set(balance)
		}
	}
	return BrokerBalance
}

// InitiateUnbounding 标记 broker 准备退出系统
// 进入 unbounding 状态后，该 broker 将不再参与 B2E 分配
func (b *Broker) InitiateUnbounding(address string) error {
	// 检查是否是有效的 broker
	if !b.IsBroker(address) {
		return fmt.Errorf("地址 %s 不是有效的 broker", address)
	}

	// 检查是否已经在 unbounding 状态
	if b.Unbounding[address] {
		return fmt.Errorf("Broker %s 已经处于 unbounding 状态", address)
	}
	// 标记为 unbounding
	b.Unbounding[address] = true
	return nil
}

// IsUnbounding 检查 broker 是否处于 unbounding 状态
func (b *Broker) IsUnbounding(address string) bool {
	return b.Unbounding[address]
}

// GetTotalLockBalance 获取 broker 在所有分片的总锁定资金
func (b *Broker) GetTotalLockBalance(address string) *big.Int {
	total := big.NewInt(0)

	if lockBalances, exists := b.LockBalance[address]; exists {
		for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
			if balance, ok := lockBalances[sid]; ok {
				total.Add(total, balance)
			}
		}
	}

	return total
}

// GetTotalBalance 获取 broker 在所有分片的资金
func (b *Broker) GetTotalBalance(address string) *big.Int {
	total := big.NewInt(0)

	if balance, exists := b.BrokerBalance[address]; exists {
		for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
			if balance, ok := balance[sid]; ok {
				total.Add(total, balance)
			}
		}
	}

	return total
}

// GetActiveBrokers 获取所有活跃的 broker 列表（排除 unbounding 状态的）
// 这个函数应该在 B2E 算法中使用，以避免给即将退出的 broker 分配交易
func (b *Broker) GetActiveBrokers() []string {
	activeBrokers := make([]string, 0)
	for _, address := range b.BrokerAddress {
		if !b.Unbounding[address] {
			activeBrokers = append(activeBrokers, address)
		}
	}
	return activeBrokers
}

// GetActiveBrokerCount 获取活跃 broker 的数量
func (b *Broker) GetActiveBrokerCount() int {
	return len(b.GetActiveBrokers())
}

// 检查 broker 是否可以退出系统
// 条件：所有分片的 LockBalance 都为 0
func (b *Broker) CanExitSystem(address string) bool {
	// 必须先进入 unbounding 状态
	if !b.Unbounding[address] {
		return false
	}

	// 检查所有分片的 LockBalance
	lockBalances, exists := b.LockBalance[address]
	if !exists {
		return true // 如果没有记录，认为可以退出
	}

	// 遍历所有分片，检查 LockBalance 是否都为 0
	for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
		if balance, ok := lockBalances[sid]; ok {
			if balance.Cmp(big.NewInt(0)) > 0 {
				fmt.Printf("[Unbounding] Broker %s 在分片 %d 仍有锁定资金: %s\n",
					address, sid, balance.String())
				return false
			}
		}
	}

	return true
}

// RemoveBroker 从系统中彻底移除 broker
// 注意：只有 返回 true 时才应该调用此函数
func (b *Broker) RemoveBroker(address string) error {
	// 检查是否可以退出
	if !b.CanExitSystem(address) {
		totalLock := b.GetTotalLockBalance(address)
		return fmt.Errorf("Broker %s 还不能退出，仍有锁定资金: %s",
			address, totalLock.String())
	}

	// 从 BrokerAddress 列表中移除
	newBrokerAddress := make([]string, 0)
	for _, addr := range b.BrokerAddress {
		if addr != address {
			newBrokerAddress = append(newBrokerAddress, addr)
		}
	}
	b.BrokerAddress = newBrokerAddress

	// 清理所有相关的 map
	delete(b.BrokerBalance, address)
	delete(b.LockBalance, address)
	delete(b.ProfitBalance, address)
	delete(b.Unbounding, address)
	delete(b.BrokerRawMegs, address)

	//fmt.Printf("[Unbounding] Broker %s 已成功从系统中移除，现在是普通用户\n", address)
	return nil
}
func (b *Broker) GetActiveBrokerBalanceSnapshot() map[string]map[uint64]*big.Int {
	snapshot := make(map[string]map[uint64]*big.Int)

	for _, addr := range b.BrokerAddress {
		// 跳过 unbonding 的 broker
		if b.IsUnbounding(addr) {
			continue
		}

		snapshot[addr] = make(map[uint64]*big.Int)
		for shardID, balance := range b.BrokerBalance[addr] {
			// ✅ 深拷贝余额（重要！）
			snapshot[addr][shardID] = new(big.Int).Set(balance)
		}
	}

	return snapshot
}

// 定期检查并处理可以退出的 broker
// 这个函数应该在每个 epoch 或区块处理完成后调用
func (b *Broker) CheckAndProcessUnboundingBrokers() {
	for address := range b.Unbounding {
		if b.Unbounding[address] && b.CanExitSystem(address) {
			fmt.Printf("[Unbounding] 检测到 Broker %s 满足退出条件，准备移除\n", address)
			err := b.RemoveBroker(address)
			if err != nil {
				fmt.Printf("[Unbounding] 移除 Broker %s 失败: %v\n", address, err)
			} else {
				fmt.Printf("[Unbounding] Broker %s 已成功退出系统\n", address)
			}
		}
	}
}

// AddBroker 添加新的 broker 到系统
// address: 新 broker 的地址
// initialBalance: 初始质押金额（每个分片的余额）
func (b *Broker) AddBroker(address string, initialBalance *big.Int) error {
	// 检查是否已经是 broker
	if b.IsBroker(address) {
		return fmt.Errorf("地址 %s 已经是 broker", address)
	}

	// 验证地址格式（可选，根据你的地址格式要求）
	if address == "" {
		return fmt.Errorf("broker 地址不能为空")
	}

	// 验证初始余额（可选，设置最小质押要求）

	// 添加到 BrokerAddress 列表
	b.BrokerAddress = append(b.BrokerAddress, address)

	// 初始化各个分片的余额
	b.BrokerBalance[address] = make(map[uint64]*big.Int)
	b.LockBalance[address] = make(map[uint64]*big.Int)
	b.ProfitBalance[address] = make(map[uint64]*big.Float)

	shardBalance := new(big.Int).Div(new(big.Int).Set(initialBalance), big.NewInt(int64(params.ShardNum)))
	for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
		b.BrokerBalance[address][sid] = new(big.Int).Set(shardBalance)
		b.LockBalance[address][sid] = big.NewInt(0)
		b.ProfitBalance[address][sid] = big.NewFloat(0)
	}

	fmt.Printf("[Join] Broker %s 已成功加入系统，初始余额: %s\n",
		address, initialBalance.String())

	return nil
}

// ReactivateBroker 重新激活一个已退出的 broker
// 这个函数用于让之前退出的 broker 重新加入系统
func (b *Broker) ReactivateBroker(address string, initialBalance *big.Int) error {
	// 检查是否已经是活跃 broker
	if b.IsBroker(address) {
		return fmt.Errorf("地址 %s 已经是活跃的 broker", address)
	}

	// 检查是否在 unbounding 状态（不允许重新激活正在退出的）
	if b.Unbounding[address] {
		return fmt.Errorf("地址 %s 正在退出过程中，不能重新激活", address)
	}

	// 添加回系统（与 AddBroker 类似）
	b.BrokerAddress = append(b.BrokerAddress, address)

	// 重新初始化余额
	b.BrokerBalance[address] = make(map[uint64]*big.Int)
	b.LockBalance[address] = make(map[uint64]*big.Int)
	b.ProfitBalance[address] = make(map[uint64]*big.Float)

	shardBalance := new(big.Int).Div(new(big.Int).Set(params.Init_broker_Balance), big.NewInt(int64(params.ShardNum)))
	for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
		b.BrokerBalance[address][sid] = new(big.Int).Set(shardBalance)
		b.LockBalance[address][sid] = big.NewInt(0)
		b.ProfitBalance[address][sid] = big.NewFloat(0)
	}

	fmt.Printf("[Join] Broker %s 已重新激活\n", address)

	return nil
}

// B2E 函数不需要修改，保持原样

// 在 broker.go 中添加辅助函数
func (b *Broker) GetActiveBrokerBalance() map[string]map[uint64]*big.Int {
	activeBrokerBalance := make(map[string]map[uint64]*big.Int)

	for _, activeBroker := range b.GetActiveBrokers() {
		activeBrokerBalance[activeBroker] = b.BrokerBalance[activeBroker]
	}

	return activeBrokerBalance
}
