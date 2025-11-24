package params

var (
	Block_Interval      = 5000  // generate new block interval
	MaxBlockSize_global = 500   // the block contains the maximum number of transactions
	InjectSpeed         = 5000  // the transaction inject speed
	TotalDataSize       = 11000 // the total number of txs
	BatchSize           = 5000  // supervisor read a batch of txs then send them, it should be larger than inject speed
	BrokerNum           = 200
	NodesInShard        = 4
	ShardNum            = 6
	IterNum_B2E         = 5
	Brokerage           = 0.1
	DataWrite_path      = "./result/"                               // measurement data result output path
	LogWrite_path       = "./log"                                   // log output path
	SupervisorAddr      = "127.0.0.1:18800"                         //supervisor ip address
	FileInput           = `./1000000to1999999_BlockTransaction.csv` //the raw BlockTransaction data path
)
