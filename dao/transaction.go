package dao

type Transaction struct {
	Id               int64  `json:"id"`                // 主键
	Hash             string `json:"hash"`              // 交易的哈希值
	Nonce            string `json:"nonce"`             // 交易的序列号
	BlockHash        string `json:"block_hash"`        // 当前交易被打包的区块的哈希值
	BlockNumber      string `json:"block_number"`      // 当前交易被打包在的区块的区块号
	TransactionIndex string `json:"transactionIndex"`  // 当前交易在区块已打包交易数组中的下标
	From             string `json:"from"`              // 交易发起者的地址
	To               string `json:"to"`                // 交易接收者的地址
	Value            string `json:"value"`             // 交易的数值
	GasPrice         string `json:"gasPrice"`          // gasPrice
	Gas              string `json:"gas"`               // gasLimit
	Input            string `xorm:"text" json:"input"` // data
}
