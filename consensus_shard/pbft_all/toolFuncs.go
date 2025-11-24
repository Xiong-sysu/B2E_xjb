package pbft_all

import (
	"blockEmulator/message"
	"blockEmulator/params"
	"blockEmulator/shard"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"log"
	"os"
	"strconv"
)

// set 2d map, only for pbft maps, if the first parameter is true, then set the cntPrepareConfirm map,
// otherwise, cntCommitConfirm map will be set
func (p *PbftConsensusNode) set2DMap(isPrePareConfirm bool, key string, val *shard.Node) {
	if isPrePareConfirm {
		if _, ok := p.cntPrepareConfirm[key]; !ok {
			p.cntPrepareConfirm[key] = make(map[*shard.Node]bool)
		}
		p.cntPrepareConfirm[key][val] = true
	} else {
		if _, ok := p.cntCommitConfirm[key]; !ok {
			p.cntCommitConfirm[key] = make(map[*shard.Node]bool)
		}
		p.cntCommitConfirm[key][val] = true
	}
}

// get neighbor nodes in a shard
func (p *PbftConsensusNode) getNeighborNodes() []string {
	receiverNodes := make([]string, 0)
	for _, ip := range p.ip_nodeTable[p.ShardID] {
		receiverNodes = append(receiverNodes, ip)
	}
	return receiverNodes
}

func (p *PbftConsensusNode) writeCSVline(str []string) {
	dirpath := params.DataWrite_path + "pbft_" + strconv.Itoa(int(p.pbftChainConfig.ShardNums))
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}

	targetPath := dirpath + "/Shard" + strconv.Itoa(int(p.ShardID)) + strconv.Itoa(int(p.pbftChainConfig.ShardNums)) + ".csv"
	f, err := os.Open(targetPath)
	if err != nil && os.IsNotExist(err) {
		file, er := os.Create(targetPath)
		if er != nil {
			panic(er)
		}
		defer file.Close()

		w := csv.NewWriter(file)
		title := []string{"blockHeight", "txpool size", "BlockBodyLength", "txExcuted", "broker1Txs", " broker2Txs", "AllocatedTxs", "BAT_byte_Size", "block_byte_Size", "BAT_byte_ratio", "Relay1TxNum", "Relay2TxNum", "innerTxs", "relayer_1_2_txs", "HasBroker_count"}
		w.Write(title)
		w.Flush()
		w.Write(str)
		w.Flush()
	} else {
		file, err := os.OpenFile(targetPath, os.O_APPEND|os.O_RDWR, 0666)

		if err != nil {
			log.Panic(err)
		}
		defer file.Close()
		writer := csv.NewWriter(file)
		err = writer.Write(str)
		if err != nil {
			log.Panic()
		}
		writer.Flush()
	}

	f.Close()
}

// writeCSV_txTime writes multiple transaction timing rows to the tx_makespan CSV.
// Accepts rows as [][]string (each inner slice is a CSV row).
func (p *PbftConsensusNode) writeCSV_txTime(rows [][]string) {
	if len(rows) == 0 {
		return
	}
	dirpath := params.DataWrite_path + "pbft_" + strconv.Itoa(int(p.pbftChainConfig.ShardNums))
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}

	targetPath := dirpath + "/Shard" + strconv.Itoa(int(p.ShardID)) + strconv.Itoa(int(p.pbftChainConfig.ShardNums)) + "tx_makespan" + ".csv"

	// If file does not exist, create and write header + all rows.
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		file, er := os.Create(targetPath)
		if er != nil {
			panic(er)
		}
		defer file.Close()

		w := csv.NewWriter(file)
		title := []string{"blockHeight", "tx_proposetime", "tx_committime", "tx_makespan(ms)", "is_bat", "is_broker1Tx", "is_broker2Tx", "is_hasbroker", "is_normal_itx", "is_relay1", "is_relay2"}
		_ = w.Write(title)
		_ = w.WriteAll(rows)
		w.Flush()
		return
	}

	// Otherwise open file and append all rows in one shot to reduce I/O.
	file, err := os.OpenFile(targetPath, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	err = writer.WriteAll(rows)
	if err != nil {
		log.Panic(err)
	}
	writer.Flush()
}

// get the digest of request
func getDigest(r *message.Request) []byte {
	b, err := json.Marshal(r)
	if err != nil {
		log.Panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}
