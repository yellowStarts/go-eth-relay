package model

import "eth-relay/dao"

// 区块信息结构体
type FullBlock struct {
	Number           string            `json:"number"`           // 区块号
	Hash             string            `json:"hash"`             // 区块的哈希值
	ParentHash       string            `json:"parentHash"`       // 父区块的哈希值
	Nonce            string            `json:"nonce"`            // 区块的序列号
	Sha3Uncles       string            `json:"sha3Uncles"`       // 当前区块如果打包了叔块，那么它就是叔块的 sha3 加密值
	LogsBloom        string            `json:"logsBloom"`        // 当前区块的布隆过滤器日志
	TransactionsRoot string            `json:"transactionsRoot"` // 交易默克尔树的根部 hash 值
	ReceiptsRoot     string            `json:"stateRoot"`        // 收据默克尔树的根部 hash 值
	Miner            string            `json:"miner"`            // 挖出此区块的矿工的以太坊地址值
	Difficulty       string            `json:"difficulty"`       // 这个区块的难度值
	TotalDifficulty  string            `json:"totalDifficulty"`  // 这个块的链的总难度
	ExtraData        string            `json:"extraData"`        // 区块的附属数据
	Size             string            `json:"size"`             // 这个区块总数居量的大小
	GasLimit         string            `json:"gasLimit"`         // 区块的 GasLimit 注意他和交易的不一样
	GasUsed          string            `json:"gasUsed"`          // 当前该区块已经打包了的交易的总燃料费
	Timestamp        string            `json:"timestamp"`        // 区块被确认核实的时间戳，单位为秒
	Uncles           []string          `json:"uncles"`           // 叔块的哈希数组
	Transactions     []dao.Transaction `json:"transactions"`     // 所有被打包了的交易的数组
}
