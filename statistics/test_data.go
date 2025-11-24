package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"blockEmulator/params"
	"blockEmulator/utils"
)

// CountIntraShard reads transactions from a CSV file and counts intra-shard txs.
// It uses the existing utils.Addr2Shard function to determine shard IDs.
func CountIntraShard(csvPath string, maxRecords int) (itx int, vaild_total int, total_count int, ctxCount int, err error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	count := 0
	itxCount := 0
	vaild_total = 0
	ctxCount = 0
	for {
		if vaild_total > 0 && vaild_total >= maxRecords {
			break
		}
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return itxCount, vaild_total, count, ctxCount, err
		}

		count++
		// follow the same filtering / parsing rules used elsewhere in the repo
		// data[3] = sender, data[4] = recipient (with 0x prefix in source files)
		if record[6] != "0" || record[7] != "0" || len(record[3]) <= 16 || len(record[4]) <= 16 || record[3] == record[4] || len(record[10]) == 0 {
			continue
		}
		// strip possible 0x prefix
		sender := record[3][2:]
		recipient := record[4][2:]

		sSid := utils.Addr2Shard(sender)
		rSid := utils.Addr2Shard(recipient)
		vaild_total++
		if sSid == rSid {
			itxCount++
		}
		if sSid != rSid {
			ctxCount++
		}
	}

	return itxCount, vaild_total, count, ctxCount, nil
}

func main() {
	csvPath := flag.String("csv", params.FileInput, "path to transactions CSV")
	max := flag.Int("n", 0, "maximum number of records to process (0 = all)")
	flag.Parse()

	itx, vaild_total, total, ctxCount, err := CountIntraShard(*csvPath, *max)
	if err != nil {
		log.Fatalf("failed to count intra-shard txs: %v", err)
	}
	ratio := 0.0
	if total > 0 {
		ratio = float64(itx) / float64(vaild_total)
	}
	fmt.Printf("Processed records: %d\n", total)
	fmt.Printf("Valid records: %d\n", vaild_total)
	fmt.Printf("Intra-shard txs (itx): %d\n", itx)
	fmt.Printf("Cross-shard txs (ctx): %d\n", ctxCount)
	fmt.Printf("Ratio itx/valid_total: %.4f\n", ratio)
	fmt.Printf("Ratio itx/total: %.4f\n", float64(itx)/float64(total))
	fmt.Printf("theory ratio itx/total: %.4f\n", float64(1)/float64(params.ShardNum))
}
