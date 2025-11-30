package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

// 账户交易计数结构
type AccountCount struct {
	Address string
	Count   int
}

// 从CSV文件中提取交易数据并统计账户交易数量
func extractTopAccounts(csvPath string, maxRecords int) ([]AccountCount, error) {
	// 打开CSV文件
	f, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 创建CSV读取器
	reader := csv.NewReader(f)

	// 用于统计每个账户的交易数量
	accountTxCount := make(map[string]int)
	count := 0
	validCount := 0

	// 逐行读取CSV数据
	for {
		if maxRecords > 0 && validCount >= maxRecords {
			break
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		count++
		// 使用与现有代码相同的过滤规则
		if record[6] != "0" || record[7] != "0" || len(record[3]) <= 16 || len(record[4]) <= 16 || record[3] == record[4] || len(record[10]) == 0 {
			continue
		}

		// 移除0x前缀并统计发送方和接收方的交易数量
		sender := record[3][2:]
		recipient := record[4][2:]

		accountTxCount[sender]++
		accountTxCount[recipient]++
		validCount++

		// 每处理100000条记录输出进度
		if validCount%100000 == 0 {
			fmt.Printf("已处理 %d 条有效交易记录...\n", validCount)
		}
	}

	// 将map转换为切片以便排序
	var accountCounts []AccountCount
	for addr, count := range accountTxCount {
		accountCounts = append(accountCounts, AccountCount{Address: addr, Count: count})
	}

	// 按交易数量降序排序
	sort.Slice(accountCounts, func(i, j int) bool {
		return accountCounts[i].Count > accountCounts[j].Count
	})

	fmt.Printf("总处理记录数: %d\n", count)
	fmt.Printf("有效交易记录数: %d\n", validCount)
	fmt.Printf("统计到的唯一账户数: %d\n", len(accountCounts))

	return accountCounts, nil
}

// 将前N个账户地址保存到文件
func saveTopAccountsToFile(accounts []AccountCount, n int, filename string) error {
	// 确保n不超过账户总数
	if n > len(accounts) {
		n = len(accounts)
	}

	// 创建输出文件
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// 写入前N个账户地址
	for i := 0; i < n; i++ {
		_, err := f.WriteString(accounts[i].Address + "\n")
		if err != nil {
			return err
		}
	}

	fmt.Printf("已将前 %d 个账户地址保存到 %s\n", n, filename)
	return nil
}

func main() {
	// 命令行参数
	// csvPath1 := flag.String("csv1", "0to999999_BlockTransaction.csv", "第一个交易CSV文件路径")
	csvPath := flag.String("csv2", "1000000to1999999_BlockTransaction.csv", "第二个交易CSV文件路径")
	maxRecords := flag.Int("n", 0, "每个文件处理的最大记录数 (0 = 全部)")
	flag.Parse()

	// fmt.Println("开始处理第一个交易文件...")
	// accounts1, err := extractTopAccounts(*csvPath1, *maxRecords)
	// if err != nil {
	// 	log.Fatalf("处理第一个文件失败: %v", err)
	// }

	fmt.Println("\n开始处理交易文件...")
	accounts, err := extractTopAccounts(*csvPath, *maxRecords)
	if err != nil {
		log.Fatalf("处理文件失败: %v", err)
	}

	// // 合并两个文件的统计结果
	// mergedAccounts := make(map[string]int)

	// 合并第一个文件的结果
	// for _, acc := range accounts1 {
	// 	mergedAccounts[acc.Address] = acc.Count
	// }

	// 合并第二个文件的结果
	// for _, acc := range accounts2 {
	// 	mergedAccounts[acc.Address] += acc.Count
	// }

	// 转换为排序的切片
	var finalAccounts []AccountCount
	for _, value := range accounts {
		finalAccounts = append(finalAccounts, AccountCount{Address: value.Address, Count: value.Count})
	}

	// 按交易数量降序排序
	sort.Slice(finalAccounts, func(i, j int) bool {
		return finalAccounts[i].Count > finalAccounts[j].Count
	})

	fmt.Printf("\n合并后的唯一账户总数: %d\n", len(finalAccounts))
	fmt.Println("\n前10个交易最多的账户:")
	for i := 0; i < 10 && i < len(finalAccounts); i++ {
		fmt.Printf("%d. %s - %d 笔交易\n", i+1, finalAccounts[i].Address, finalAccounts[i].Count)
	}

	// 需要提取的账户数量列表
	nValues := []int{200, 400, 500, 600, 800, 1000}

	// 创建输出目录
	outputDir := "top_accounts"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	// 保存不同数量的顶级账户
	for _, n := range nValues {
		filename := fmt.Sprintf("%s/broker%d.txt", outputDir, n)
		err = saveTopAccountsToFile(finalAccounts, n+1, filename)
		if err != nil {
			log.Printf("保存前 %d 个账户失败: %v", n, err)
		}
	}

	fmt.Println("\n所有文件已成功生成!")
}
