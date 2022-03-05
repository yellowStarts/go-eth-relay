package main

import (
	"encoding/json"
	"eth-relay/tool"
	"fmt"
	"testing"
)

func TestNewETHRPCClient(t *testing.T) {
	// 首先是一个格式正确的链接测试初始化
	client2 := NewETHRPCClient("www.nihao.com").GetRpc()
	if client2 == nil {
		fmt.Println("初始化失败")
	}
	// 接着是 123://356 非法链接测试初始化
	client := NewETHRPCClient("123://456").GetRpc()
	if client == nil {
		fmt.Println("初始化失败")
	}
}

func Test_GetTransactionByHash(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71"
	txHash := "0xd34279f67e05c398a863177b73b735a6141deba3bda62342a8f2c91f36a22f8e"
	if txHash == "" || len(txHash) != 66 {
		// 这里演示在调用 rpc 接口函数的时候，要先进行入参的合法性判断
		fmt.Println("非法的交易哈希值")
		return
	}
	txInfo, err := NewETHRPCRequester(nodeUrl).GetTransactionByHash(txHash)
	if err != nil {
		// 查询失败，打印出信息
		fmt.Println("查询交易失败，信息是：", err.Error())
		return
	}
	// 查询成功，将 transaction 结果的结构体以 json 格式序列化，在以 string 格式输出
	json, _ := json.Marshal(txInfo)
	fmt.Println(string(json))
}

func Test_GetTransactions(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71"
	txHash_1 := "0xd34279f67e05c398a863177b73b735a6141deba3bda62342a8f2c91f36a22f8e"
	txHash_2 := "0x52a1dc843a9918b76e71334a034d46e4cf4834bcfa2409bc7286baa5bca91eed"
	txHash_3 := "0x52a1dc843a9918b76e71334a034d46e4cf4834bcfa2409bc7286baa5bca91exx"

	// taHash_1,taHash_2是存在的，taHash_3是伪造的
	txHashs := []string{}
	txHashs = append(txHashs, txHash_1, txHash_2, txHash_3)

	if txHashs == nil || len(txHashs) == 0 {
		// 这里演示在调用 rpc 接口函数的时候，要先进行入参的合法性判断
		fmt.Println("非法的交易哈希值数组")
		return
	}
	txInfos, err := NewETHRPCRequester(nodeUrl).GetTransactions(txHashs)
	if err != nil {
		// 查询失败，打印出信息
		fmt.Println("查询交易失败，信息是：", err.Error())
		return
	}
	// 查询成功，将 transaction 结果的结构体以 json 格式序列化，在以 string 格式输出
	json, _ := json.Marshal(txInfos)
	fmt.Println(string(json))
}

// 单笔交易的单元测试函数
func Test_GetETHBalance(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71"
	address := "0x4ad64983349c49defe8d7a4686202d24b25d0ce8"
	if address == "" || len(address) != 42 {
		// 这里演示在调用 rpc 接口函数的时候，要先进行入参的合法性判断
		fmt.Println("非法的交易地址值")
		return
	}
	balance, err := NewETHRPCRequester(nodeUrl).GetETHBalance(address)
	if err != nil {
		// 查询失败，打印出信息
		fmt.Println("查询 eth 余额失败，信息是：", err.Error())
		return
	}
	fmt.Println(balance)
}

// 批量交易的单元测试函数
func Test_GetETHBalances(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71"

	address1 := "0x4ad64983349c49defe8d7a4686202d24b25d0ce8"
	address2 := "0xcad621da75a66c7a8f4ff86d30a2bf981bfc8fdd"

	addresss := []string{address1, address2}

	balances, err := NewETHRPCRequester(nodeUrl).GetETHBalances(addresss)
	if err != nil {
		// 查询失败，打印出信息
		fmt.Println("查询 eth 余额失败，信息是：", err.Error())
		return
	}
	fmt.Println(balances)
}

// 单元测试，批量获取代币值
func Test_GetERC20Balances(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71"

	address := "0x4ad64983349c49defe8d7a4686202d24b25d0ce8"
	contract1 := "0x53C8395465A84955c95159814461466053DedEDE"
	contract2 := "0x585fc93c81a261c834783ba6d4872e9d233c2513"

	params := []ERC20BalanceRpcReq{}
	item := ERC20BalanceRpcReq{}
	item.ContractAddress = contract1
	item.UserAddress = address
	item.ContractDecimal = 18

	params = append(params, item)

	item.ContractAddress = contract2
	params = append(params, item)

	balance, err := NewETHRPCRequester(nodeUrl).GetERC20Balances(params)
	if err != nil {
		// 查询失败，打印出信息
		fmt.Println("查询 eth 余额失败，信息是：", err.Error())
		return
	}
	fmt.Println(balance)
}

// 单元测试，获取以太坊最新生成区块的区块号
func Test_GetLatestBlockNumber(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71"
	number, err := NewETHRPCRequester(nodeUrl).GetLatestBlockNumber()
	if err != nil {
		// 查询失败，打印出信息
		fmt.Println("查询区块号失败，信息是：", err.Error())
		return
	}
	fmt.Println("10进制：", number.String())
}

// 单元测试：根据区块号获取区块信息
func Test_GetBlockInfoByNunber(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71"
	requester := NewETHRPCRequester(nodeUrl)
	number, _ := requester.GetLatestBlockNumber() // 获取区块号
	fmt.Println("区块号是：\n", number)
	fullBlock, err := requester.GetBlockInfoByNumber(number) // 获取区块信息
	if err != nil {
		// 查询失败，打印出信息
		fmt.Println("查询区块信息失败，信息是：", err.Error())
		return
	}
	// 查询成功
	json1, _ := json.Marshal(fullBlock)
	fmt.Println("根据区块号获取区块信息：\n", string(json1))
}

// 单元测试：根据区块哈希值获取区块信息
func Test_GetBlockInfoByHash(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71"
	requester := NewETHRPCRequester(nodeUrl)
	blockHash := "0xd5310fc253dab0060e3d7ae6d0b88eb72f117e6e9d37a5f7b1ca5250e08249b9"
	// 根据区块哈希获取区块信息
	fullBlock, err := requester.GetBlockInfoByHash(blockHash)
	if err != nil {
		// 查询失败，打印出信息
		fmt.Println("查询区块信息失败，信息是：", err.Error())
		return
	}
	// 查询成功
	json1, _ := json.Marshal(fullBlock)
	fmt.Println("根据区块号获取区块信息：\n", string(json1))
}

// 单元测试：使用 eth_call 访问智能合约的函数
// func Test_ETHCall(t *testing.T) {
// 	// 加法智能合约的 abi 数据
// 	contractABI :=
// 		`[{"constant": true, "inputs": [{"name": "arg1", "type": "uint8"},
// 	{"name": "arg2", "type": "uint8"}],
// 	"name": "add", "outputs":[{"name": "", "type": "uint8"}],
// 	"payable": false, "stateMutability": "pure", "type": "function"}]`
// 	methodName := "add" // 加法函数名称
// 	methodId, err := tool.MakeMethodId(methodName, contractABI)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// 下面进行 2+3 运算:
// 	arg1 := common.HexToHash("2").String()[2:] // 根据 data 中的参数格式，生成第一个参数
// 	arg2 := common.HexToHash("3").String()[2:] // 根据 data 中的参数格式，生成第一个参数
// 	constractAddress := "0x313dsfdsfs..."      // 这里填写智能合约的地址
// 	args := model.CallArg{
// 		To:   common.HexToAddress(constractAddress), // 对应的是合约地址，代表访问该合约
// 		Data: methodId + arg1 + arg2,                // 组合成 data 的完整格式
// 		// 下面的无关参数可以不进行赋值，让他们使用默认值
// 		// Gas: "0x0",
// 		// GasPrice: "0x0",
// 		// Value: "0x0",
// 		// Nonce: "0x0",
// 	}
// 	result := "" // 结果是一个十六进制字符串
// 	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71"
// 	requster := NewETHRPCRequester(nodeUrl)
// 	err = requster.ETHCall(&result, args)
// 	if err != nil {
// 		panic(err)
// 	}
// 	ten, _ := new(big.Int).SetString(result[2:], 16)
// 	fmt.Println("调用合约两数相加结果是：", ten.String())
// }

// 单元测试：创建以太坊钱包
func Test_CreateETHWallet(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71"
	address1, err := NewETHRPCRequester(nodeUrl).CreateETHWallet("12345")
	// 演示密码太短的错误
	if err != nil {
		fmt.Println("第一次，创建钱包失败", err.Error())
	} else {
		fmt.Println("第一次，创建钱包成功，以太坊地址是：", address1)
	}
	address2, err := NewETHRPCRequester(nodeUrl).CreateETHWallet("123456")
	// 创建成功
	if err != nil {
		fmt.Println("第二次，创建钱包失败", err.Error())
	} else {
		fmt.Println("第二次，创建钱包成功，以太坊地址是：", address2)
	}
}

// 单元测试：获取 nonce
func Test_GetNonce(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71"
	address := "0x4ad64983349c49defe8d7a4686202d24b25d0ce8"
	if address == "" || len(address) != 42 {
		// 这里演示在调用 rpc 接口函数的时候，要先进行入参的合法性判断
		fmt.Println("非法的交易地址值")
		return
	}
	nonce, err := NewETHRPCRequester(nodeUrl).GetNonce(address)
	if err != nil {
		// 查询失败，打印出信息
		fmt.Println("查询 nonce 失败，信息是：", err.Error())
		return
	}
	fmt.Println(nonce)
}

// 单元测试：转账 ETH
func Test_SendETHTransaction(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71" // ropsten 测试网络的节点链接
	from := "0x4ad64983349c49defe8d7a4686202d24b25d0ce8"                       // 这里找一个获取测试代币的地址
	if from == "" || len(from) != 42 {
		// 这里演示在调用 rpc 接口函数的时候，要先进行入参的合法性判断
		fmt.Println("非法的交易地址值")
		return
	}
	to := "0x97376cf11717ab4a9e9a94042e895640a6262e3"
	value := "0.2" // 发送 0.2个 ETH
	gasLimit := uint64(100000)
	gasPrice := uint64(36000000000)
	// 当前这笔交易消耗的燃料费最大值是 (gasLimit * gasPrice)/10^18 ETH
	err := tool.UnlockETHWallet("./keystores", from, "123456") // 解锁钱包
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 下面发起交易转账
	txHash, err := NewETHRPCRequester(nodeUrl).SendETHTransaction(from, to, value, gasLimit, gasPrice)
	if err != nil {
		// 转账失败，打印出信息
		fmt.Println("ETH 转账失败，信息是：", err.Error())
		return
	}
	fmt.Println(txHash) // 打印出当前交易的哈希值
}

// 单元测试：转账 ERC20 代币
func Test_SendERC20Transaction(t *testing.T) {
	nodeUrl := "https://mainnet.infura.io/v3/70888e737c7b4306aa7f386af25aca71" // ropsten 测试网络的节点链接
	from := "0x4ad64983349c49defe8d7a4686202d24b25d0ce8"                       // 这里找一个获取测试代币的地址
	if from == "" || len(from) != 42 {
		// 这里演示在调用 rpc 接口函数的时候，要先进行入参的合法性判断
		fmt.Println("非法的交易地址值")
		return
	}
	to := "0x99BD856a01210D3B4b76A6f8c6fFf3eCdC485758"      // 在测试网络上发布的 MTC代币智能合约
	amount := "10"                                          // 转账 ERC20 代币的数值：10个MFTC
	decimal := 18                                           // MFTC 代币单位精确到小数点后的位数
	receiver := "0x97376cf11717ab4a9e9a94042e895640a6262e3" // 接收者的以太坊地址
	gasLimit := uint64(50000)
	gasPrice := uint64(24000000000)
	// 当前这笔交易消耗的燃料费最大值是 (gasLimit * gasPrice)/10^18 ETH
	err := tool.UnlockETHWallet("./keystores", from, "123456") // 解锁钱包
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 下面发起交易转账
	txHash, err := NewETHRPCRequester(nodeUrl).SendERC20Transaction(from, to, receiver, amount, gasLimit, gasPrice, decimal)
	if err != nil {
		// 转账失败，打印出信息
		fmt.Println("ETH 转账失败，信息是：", err.Error())
		return
	}
	fmt.Println(txHash) // 打印出当前交易的哈希值
}
