package main

import (
	"eth-relay/dao"
	"testing"
)

// 单元测试：区块扫描器，开始扫描区块
func TestBlockScanner_Start(t *testing.T) {
	// 初始化以太坊 rpc 请求者
	requester := NewETHRPCRequester("https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71")
	// 初始化数据库连接器配置对象，记得修改为本地数据库的参数
	option := dao.MySQLOptions{
		HostName:           "127.0.0.1",
		Port:               "3306",
		DbName:             "eth_reply",
		User:               "root",
		Password:           "",
		TablePrefix:        "eth_",
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    15,
	}
	// 添加数据表
	tables := []interface{}{}
	tables = append(tables, dao.Block{}, dao.Transaction{})
	// 根据上面定义的配置，初始化数据库连接器
	mysql := dao.NewMySQLConnector(&option, tables)
	// 初始化区块扫描器
	scanner := NewBlockScanner(*requester, mysql)
	err := scanner.Start() // 开始扫描
	if err != nil {
		panic(err)
	}
	// 使用 select 模拟阻塞主协程，等待上面的代码执行，因为扫描是在 gorutine 协程中进行的
	select {}
}
