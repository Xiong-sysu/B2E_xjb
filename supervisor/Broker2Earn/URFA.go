package Broker2Earn

import (
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/utils"
	"math/big"
	"sort"
)

// Relax functin
// Input tx list and brokerTable (brokers balance)
// Output a map that ctx -> broker

func URFA_Linear(brokerRawMegs []*message.BrokerRawMeg, BrokerBalance map[string]*big.Int, alpha int, sigma float64) []*RatioBrokerRawMeg {

	// get all broker address
	brokerAddresses := make([]string, 0, len(BrokerBalance))

	for brokerAddress := range BrokerBalance {
		brokerAddresses = append(brokerAddresses, brokerAddress)
	}
	// build tmp broker balance
	nowBrokerBalance := make(map[string]*big.Int)
	for brokerID, _ := range BrokerBalance {
		nowBrokerBalance[brokerID] = new(big.Int).SetInt64(0)
	}

	// deal txs
	txs := make([]*core.Transaction, 0)
	for _, brokerRawMeg := range brokerRawMegs {
		txs = append(txs, brokerRawMeg.Tx)
	}

	// sort tx by Fee/Value
	sort.Slice(txs, func(i, j int) bool {
		// tranform to big.float
		div1 := new(big.Float).SetInt(txs[i].Value)
		div1.Add(div1, big.NewFloat(0.000001))
		div1.Quo(new(big.Float).SetInt(txs[i].Fee), div1)

		div2 := new(big.Float).SetInt(txs[j].Value)
		div2.Add(div2, big.NewFloat(0.000001))
		div2.Quo(new(big.Float).SetInt(txs[j].Fee), div2)

		return div1.Cmp(div2) > 0
	})

	// allocate tx to broker 1 by 1

	result := make([]*RatioBrokerRawMeg, 0)
	brokerIndex := 0
	for _, tx := range txs {
		if brokerIndex >= len(brokerAddresses) {
			break
		}
		brokerB := nowBrokerBalance[brokerAddresses[brokerIndex]]
		brokerC := BrokerBalance[brokerAddresses[brokerIndex]]
		// if tx value bigger than broker, jump
		if brokerC.Cmp(tx.Value) < 0 {
			continue
		}

		tmp_brokerRawMeg := &RatioBrokerRawMeg{
			Tx:          tx,
			BrokerRatio: make(map[utils.Address]float64),
		}

		tmpValue := new(big.Int).SetInt64(0)
		// if tx value + allocated broker > broker balance, split tx
		if brokerB.Cmp(brokerC) < 0 && tmpValue.Add(brokerB, tx.Value).Cmp(brokerC) >= 0 {

			//println("tx address A ", tx.TxHash, " ", tx.Value.String(), " ", brokerB.String(), " ", brokerC.String())
			// full the broker
			brokerB.Set(brokerC)

			// search the rest part of tx belong to
			sub := new(big.Int).SetInt64(0)
			sub.Sub(tmpValue, brokerC)

			// calulate ratio
			div1 := new(big.Float).SetInt(sub)
			div2 := new(big.Float).SetInt(tx.Value)

			ratio, _ := div1.Quo(div1, div2).Float64()

			//println("2brokerIndex is ", brokerIndex, " / ", len(brokerAddresses))
			tmp_brokerRawMeg.BrokerRatio[brokerAddresses[brokerIndex]] = 1.0 - ratio

			brokerIndex += 1
			if brokerIndex >= len(brokerAddresses) {
				break
			}
			for j := brokerIndex; j < len(brokerAddresses); j++ {
				brokerB1 := nowBrokerBalance[brokerAddresses[brokerIndex]]
				brokerC1 := BrokerBalance[brokerAddresses[brokerIndex]]
				tmpValue1 := new(big.Int)
				tmpValue1.Add(brokerB1, sub)

				if brokerC1.Cmp(tmpValue1) < 0 {
					continue
				}
				tmp_brokerRawMeg.BrokerRatio[brokerAddresses[j]] = ratio
				break

			}

		} else {
			//println("tx address B ", tx.TxHash, " ", tx.Value.String(), " ", brokerB.String(), " ", brokerC.String())
			nowBrokerBalance[brokerAddresses[brokerIndex]].Add(brokerB, tx.Value)
			tmp_brokerRawMeg.BrokerRatio[brokerAddresses[brokerIndex]] = 1.0
		}
		result = append(result, tmp_brokerRawMeg)

	}
	return result
}

// URFA_Integer - Unit Revenue First Algorithm (Integer version)
// Pure integer allocation algorithm, standalone version
// Input: brokerRawMegs and BrokerBalance (broker balance in each shard)
// Output: allocated transactions and rest transactions
func URFA_Integer(brokerRawMegs []*message.BrokerRawMeg, BrokerBalance map[string]map[uint64]*big.Int) ([]*message.BrokerRawMeg, []*message.BrokerRawMeg) {

	// 1. Filter cross-shard transactions
	crossShardTxs := make([]*message.BrokerRawMeg, 0)
	for _, brokerRawMeg := range brokerRawMegs {
		if utils.Addr2Shard(brokerRawMeg.Tx.Recipient) != utils.Addr2Shard(brokerRawMeg.Tx.Sender) {
			crossShardTxs = append(crossShardTxs, brokerRawMeg)
		}
	}

	brokerBalance := make(map[string]*big.Int)
	brokerAddresses := make([]string, 0)

	for brokerID, shardAmounts := range BrokerBalance {
		totalAmount := big.NewInt(0)
		for _, amount := range shardAmounts {
			totalAmount.Add(totalAmount, amount)
		}
		brokerBalance[brokerID] = totalAmount
		brokerAddresses = append(brokerAddresses, brokerID)
	}

	weightCapacities := make(map[string]*big.Int)
	for brokerID := range brokerBalance {
		weightCapacities[brokerID] = new(big.Int).SetInt64(0)
	}

	type TxWithDensity struct {
		BrokerRawMeg *message.BrokerRawMeg
		Density      *big.Float
	}

	txsWithDensity := make([]*TxWithDensity, 0, len(crossShardTxs))
	for _, brokerRawMeg := range crossShardTxs {
		tx := brokerRawMeg.Tx

		revenue := new(big.Float).SetInt(tx.Fee)
		revenue.Add(revenue, big.NewFloat(1.0))

		denominator := new(big.Float).SetInt(tx.Value)
		denominator.Add(denominator, big.NewFloat(0.000001))
		density := new(big.Float).Quo(revenue, denominator)

		txsWithDensity = append(txsWithDensity, &TxWithDensity{
			BrokerRawMeg: brokerRawMeg,
			Density:      density,
		})
	}

	sort.Slice(txsWithDensity, func(i, j int) bool {
		return txsWithDensity[i].Density.Cmp(txsWithDensity[j].Density) > 0
	})

	result := make([]*message.BrokerRawMeg, 0)
	allocatedTxMap := make(map[*core.Transaction]bool)
	brokerIndex := 0

	for _, txWithDensity := range txsWithDensity {
		if brokerIndex >= len(brokerAddresses) {
			break
		}

		brokerRawMeg := txWithDensity.BrokerRawMeg
		tx := brokerRawMeg.Tx
		currentBroker := brokerAddresses[brokerIndex]
		usedCapacity := weightCapacities[currentBroker]
		totalCapacity := brokerBalance[currentBroker]

		tmpCapacity := new(big.Int).Add(usedCapacity, tx.Value)
		if tmpCapacity.Cmp(totalCapacity) > 0 {
			brokerIndex++
			continue
		}

		weightCapacities[currentBroker].Add(usedCapacity, tx.Value)

		allocatedBrokerRawMeg := &message.BrokerRawMeg{
			Broker: currentBroker,
			Tx:     tx,
		}
		result = append(result, allocatedBrokerRawMeg)
		allocatedTxMap[tx] = true
	}

	restBrokerRawMeg := make([]*message.BrokerRawMeg, 0)
	for _, brokerRawMeg := range brokerRawMegs {
		if !allocatedTxMap[brokerRawMeg.Tx] {
			restBrokerRawMeg = append(restBrokerRawMeg, brokerRawMeg)
		}
	}

	return result, restBrokerRawMeg
}

// BrokerChain - Shard-based broker allocation
// Evenly distribute broker funds across shards, then greedily allocate transactions
// Input: brokerRawMegs and BrokerBalance
// Output: allocated transactions and rest transactions
func BrokerChain(brokerRawMegs []*message.BrokerRawMeg, BrokerBalance map[string]map[uint64]*big.Int) ([]*message.BrokerRawMeg, []*message.BrokerRawMeg) {

	// 1. Filter cross-shard transactions
	crossShardTxs := make([]*message.BrokerRawMeg, 0)
	for _, brokerRawMeg := range brokerRawMegs {
		if utils.Addr2Shard(brokerRawMeg.Tx.Recipient) != utils.Addr2Shard(brokerRawMeg.Tx.Sender) {
			crossShardTxs = append(crossShardTxs, brokerRawMeg)
		}
	}

	// 2. Calculate total number of shards
	shardSet := make(map[uint64]bool)
	for _, shardAmounts := range BrokerBalance {
		for shardID := range shardAmounts {
			shardSet[shardID] = true
		}
	}
	shardNum := len(shardSet)
	if shardNum == 0 {
		return []*message.BrokerRawMeg{}, brokerRawMegs
	}

	// 3. Evenly distribute broker funds across shards
	brokerShardBalance := make(map[string]map[uint64]*big.Int)
	brokerAddresses := make([]string, 0)

	for brokerID, shardAmounts := range BrokerBalance {
		brokerAddresses = append(brokerAddresses, brokerID)

		// Calculate total funds
		totalAmount := big.NewInt(0)
		for _, amount := range shardAmounts {
			totalAmount.Add(totalAmount, amount)
		}

		// Average allocation to each shard
		avgAmount := new(big.Int).Div(totalAmount, big.NewInt(int64(shardNum)))
		brokerShardBalance[brokerID] = make(map[uint64]*big.Int)

		for shardID := range shardSet {
			brokerShardBalance[brokerID][shardID] = new(big.Int).Set(avgAmount)
		}
	}

	// 4. Sort transactions by fee/value ratio (greedy strategy)
	sort.Slice(crossShardTxs, func(i, j int) bool {
		div1 := new(big.Float).SetInt(crossShardTxs[i].Tx.Fee)
		denom1 := new(big.Float).SetInt(crossShardTxs[i].Tx.Value)
		denom1.Add(denom1, big.NewFloat(0.000001))
		ratio1 := div1.Quo(div1, denom1)

		div2 := new(big.Float).SetInt(crossShardTxs[j].Tx.Fee)
		denom2 := new(big.Float).SetInt(crossShardTxs[j].Tx.Value)
		denom2.Add(denom2, big.NewFloat(0.000001))
		ratio2 := div2.Quo(div2, denom2)

		return ratio1.Cmp(ratio2) > 0
	})

	// 5. Greedily allocate transactions
	result := make([]*message.BrokerRawMeg, 0)
	allocatedTxMap := make(map[*core.Transaction]bool)

	for _, brokerRawMeg := range crossShardTxs {
		tx := brokerRawMeg.Tx
		recipientShard := uint64(utils.Addr2Shard(tx.Recipient))

		// Try to find a broker with sufficient funds in recipient shard
		allocated := false
		for _, brokerID := range brokerAddresses {
			if brokerShardBalance[brokerID][recipientShard].Cmp(tx.Value) >= 0 {
				// Allocate to this broker
				brokerShardBalance[brokerID][recipientShard].Sub(
					brokerShardBalance[brokerID][recipientShard],
					tx.Value,
				)

				allocatedBrokerRawMeg := &message.BrokerRawMeg{
					Broker: brokerID,
					Tx:     tx,
				}
				result = append(result, allocatedBrokerRawMeg)
				allocatedTxMap[tx] = true
				allocated = true
				break
			}
		}

		if !allocated {
			// No broker can handle this transaction
			continue
		}
	}

	// 6. Calculate rest transactions
	restBrokerRawMeg := make([]*message.BrokerRawMeg, 0)
	for _, brokerRawMeg := range brokerRawMegs {
		if !allocatedTxMap[brokerRawMeg.Tx] {
			restBrokerRawMeg = append(restBrokerRawMeg, brokerRawMeg)
		}
	}

	return result, restBrokerRawMeg
}
