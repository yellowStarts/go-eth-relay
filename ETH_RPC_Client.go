package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rpc"
)

type ETHRPCClient struct {
	NodeUrl string      // 代表节点的 url 链接
	client  *rpc.Client // 代表 rpc 客户端句柄实例
}

// NewETHRPCClient 代表的是新建一个 RPC 客户端
// 参数 nodeUrl 是节点的链接，返回的是 ETHRPCClient 对象指针
func NewETHRPCClient(nodeUrl string) *ETHRPCClient {
	client := &ETHRPCClient{
		NodeUrl: nodeUrl,
	}
	client.initRpc() // 进行初始化  rpc 客户端句柄实例
	return client
}

// 初始化 rpc 客户端句柄
func (erc *ETHRPCClient) initRpc() {
	// 使用 go-ethereum 库中的 rpc 来初始化
	// DialHttp 的意思是使用 http 版本的 rpc 实现方式
	rpcClient, err := rpc.DialHTTP(erc.NodeUrl)
	if err != nil {
		// 初始化失败，终结程序，并将错误信息显示到控制台中
		errInfo := fmt.Errorf("初始化 rpc client 失败%s", err.Error()).Error()
		panic(errInfo)
	}
	// 初始化成功，将新实例化的 rpc 句柄赋值给 ETHRPCClient 结构体中的 client
	erc.client = rpcClient
}

// GetRpc 函数是为了方便外部能够获取 client *rpc.Client，以方便进行访问
func (erc *ETHRPCClient) GetRpc() *rpc.Client {
	if erc.client == nil {
		erc.initRpc()
	}
	return erc.client
}
