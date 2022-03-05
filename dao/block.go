package dao

// 存储区块信息的区块结构体
type Block struct {
	Id          int64  `json:"id"`           // 主键
	BlockNumber string `json:"block_number"` // 区块号
	BlockHash   string `json:"block_hash"`   // 区块的哈希值
	ParentHash  string `json:"parent_hash"`  // 父区块的哈希值
	CreateTime  int64  `json:create_time`    // 区块的生成时间
	Fork        bool   `json:"fork"`         // 是否为分叉区块
}
