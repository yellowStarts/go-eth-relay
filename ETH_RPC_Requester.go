package main

import (
	"errors"
	"eth-relay/model"
	"eth-relay/tool"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

type ETHRPCRequester struct {
	nonceManager *NonceManager // noce 管理器实例
	client       *ETHRPCClient // rpc 客户端
}

// NewETHRPCRequester 实例化
func NewETHRPCRequester(nodeUrl string) *ETHRPCRequester {
	requester := &ETHRPCRequester{}
	// 实例化 noce 管理器
	requester.nonceManager = NewNonceManager()
	// 实例化 rpc 客户端
	requester.client = NewETHRPCClient(nodeUrl)
	return requester
}

// GetTransactionByHash 根据交易的哈希值获取对应的交易信息
func (r *ETHRPCRequester) GetTransactionByHash(txHash string) (model.Transaction, error) {
	methodName := "eth_getTransactionByHash"
	result := model.Transaction{}
	// 下面 call 函数的 result 参数传入的是 model.Transaction 结构体的引用
	// 这样内部所设置的值在函数执行完之后才能有效果
	err := r.client.GetRpc().Call(&result, methodName, txHash)
	return result, err
}

// 根据交易哈希值字符串的数组批量获取对应的交易信息
func (r *ETHRPCRequester) GetTransactions(txHashs []string) ([]*model.Transaction, error) {
	name := "eth_getTransactionByHash"
	// 结果数组存储的是每个请求的结果指针，也就是引用
	results := []*model.Transaction{}
	// 获取好细致数组的长度，方便再循环中逐个实例化 BatchElem
	size := len(txHashs)

	requesters := []rpc.BatchElem{}
	for i := 0; i < size; i++ {
		result := model.Transaction{}
		// 实例化每个 BatchElem
		requester := rpc.BatchElem{
			Method: name,
			Args:   []interface{}{txHashs[i]},
			// &ret 传入单个请求的结果引用，保证它在函数内部被修改值后，回到函数外时仍然有效
			Result: &result,
		}
		// 将每个 BatchElem 添加到 BatchElem 数组
		requesters = append(requesters, requester)
		// 每个请求的结果引用添加到结果数组中
		results = append(results, &result)
	}
	// 传入 BatchElem 数组，发起批量请求
	err := r.client.GetRpc().BatchCall(requesters)
	return results, err
}

// 单笔查询，根据以太坊地址，查询以太坊 eth 的余额
func (r *ETHRPCRequester) GetETHBalance(address string) (string, error) {
	name := "eth_getBalance"
	result := ""
	// 对应文档，第一个参数就是要查询的以太坊地址，第二个参数是 latest
	err := r.client.GetRpc().Call(&result, name, address, "latest")
	if err != nil {
		return "", err
	}
	if result == "" {
		return "", errors.New("eth balance is null")
	}
	// 因为查询所返回的结果是一个十六进制的字符串
	// 为了方便阅读，我们在下面使用 go 的大数处理将其转换为十进制数
	// 并防止数位溢出
	ten, _ := new(big.Int).SetString(result[2:], 16)
	return ten.String(), nil
}

// 批量查询，根据以太坊地址数组，查询以太坊 eth 的余额
func (r *ETHRPCRequester) GetETHBalances(addresss []string) ([]string, error) {
	name := "eth_getBalance"
	// 结果数组存储的是每个请求的结果指针，也就是引用
	rets := []*string{}
	// 获取 address 数组的长度，方便在循环中逐个实例化 BatchElem
	size := len(addresss)
	reqs := []rpc.BatchElem{}
	for i := 0; i < size; i++ {
		ret := ""
		// 实例化每个 BatchElem
		req := rpc.BatchElem{
			Method: name,
			Args:   []interface{}{addresss[i], "latest"},
			// &ret 传入单个请求的结果引用，保证它在函数内部被修改值后，回到函数外时仍然有效
			Result: &ret,
		}
		// 将每个 BatchElem 添加到 BatchElem 数组
		reqs = append(reqs, req)
		// 每个请求的结果引用添加到结果数组中
		rets = append(rets, &ret)
	}
	// 传入 BatchElem 数组，发起批量请求
	err := r.client.GetRpc().BatchCall(reqs)
	if err != nil {
		return nil, err
	}
	// 查询每个请求有没有错误
	for _, req := range reqs {
		if req.Error != nil {
			return nil, req.Error
		}
	}
	finalRet := []string{}
	for _, item := range rets {
		ten, _ := new(big.Int).SetString((*item)[2:], 16)
		finalRet = append(finalRet, ten.String())
	}
	return finalRet, err
}

// ERC20BalanceRpcReq 是查询 ERC20 代币的参数集合结构体
type ERC20BalanceRpcReq struct {
	ContractAddress string // 合约的以太坊地址
	UserAddress     string // 用户的以太坊地址
	ContractDecimal int    // 合约所对应代币单位精确到小数点后的位数
}

// 批量查询：根据以太坊地址数组，查询 ERC20 代币的余额
func (r *ETHRPCRequester) GetERC20Balances(paramArr []ERC20BalanceRpcReq) ([]string, error) {
	name := "eth_call"
	methodId := "0x70a08231" // 这个是 balanceOf 的 methodId
	// 结果数组存储的是每个请求的结果指针，也就是引用
	rets := []*string{}
	// 获取 address 数组的长度，方便在循环中逐个实例化 BatchElem
	size := len(paramArr)
	reqs := []rpc.BatchElem{}
	for i := 0; i < size; i++ {
		ret := ""
		arg := &model.CallArg{}
		userAddress := paramArr[i].UserAddress
		// 下面是针对方位 balanceOf 时的必须参数，查询余额时不需要燃料费的，所以不需要设置 Gas
		arg.To = common.HexToAddress(paramArr[i].ContractAddress)
		arg.Data = methodId + "000000000000000000000000" + userAddress[2:]
		// 实例化每个 BatchElem
		req := rpc.BatchElem{
			Method: name,
			Args:   []interface{}{arg, "latest"},
			// &ret 传入单个请求的结果引用，保证它在函数内部被修改值后，回到函数外时仍然有效
			Result: &ret,
		}

		// 将每个 BatchElem 添加到 BatchElem 数组
		reqs = append(reqs, req)
		// 每个请求的结果引用添加到结果数组中
		rets = append(rets, &ret)
	}
	// 传入 BatchElem 数组，发起批量请求
	err := r.client.GetRpc().BatchCall(reqs)
	if err != nil {
		return nil, err
	}
	// 查询每个请求有没有错误
	for _, req := range reqs {
		if req.Error != nil {
			return nil, req.Error
		}
	}
	finalRet := []string{}
	for _, item := range rets {
		if *item == "" {
			continue
		}
		ten, _ := new(big.Int).SetString((*item)[2:], 16)
		finalRet = append(finalRet, ten.String())
	}
	return finalRet, err
}

// 获取以太坊最新生成区块的区块号
func (r *ETHRPCRequester) GetLatestBlockNumber() (*big.Int, error) {
	methodName := "eth_blockNumber"
	number := "" // 存储结果
	// eth_blockNumber 不需要参数
	err := r.client.client.Call(&number, methodName)
	if err != nil {
		return nil, fmt.Errorf("获取最新区块号失败！ %s", err.Error())
	}
	// 十六进制转换十进制
	ten, _ := new(big.Int).SetString(number[2:], 16)
	return ten, nil
}

// 根据区块号获取区块信息
func (r *ETHRPCRequester) GetBlockInfoByNumber(blockNumber *big.Int) (*model.FullBlock, error) {
	number := fmt.Sprintf("%#x", blockNumber) // 将 big.Int 转为 十六进制字符串
	methodName := "eth_getBlockByNumber"
	fullBlock := model.FullBlock{}
	// eth_getBlockByNumber 的第二个参数：
	// 若是 true 则返回完整的区块信息，若为false 则 transaction 部分只返回交易哈希数组
	err := r.client.client.Call(&fullBlock, methodName, number, true)
	if err != nil {
		return nil, fmt.Errorf("get block info failed! %s", err.Error())
	}
	if fullBlock.Number == "" {
		return nil, fmt.Errorf("block info is empty %s", blockNumber.String())
	}
	return &fullBlock, nil
}

// 根据区块哈希值获取区块信息
func (r *ETHRPCRequester) GetBlockInfoByHash(blockHash string) (*model.FullBlock, error) {
	methodName := "eth_getBlockByHash"
	fullBlock := model.FullBlock{}
	// eth_getBlockByHash 的第二个参数：
	// 若为true则返回完整的区块信息，为false则 transaction 部分只返回交易哈希值数组
	err := r.client.client.Call(&fullBlock, methodName, blockHash, true)
	if err != nil {
		return nil, fmt.Errorf("get block info failed! %s", err.Error())
	}
	if fullBlock.Number == "" {
		return nil, fmt.Errorf("block info is empty %s", blockHash)
	}
	return &fullBlock, nil
}

// 使用 eth_call 调用智能合约的函数
// 第一个参数是接受结果的结构体，第二个参数是 eth_call 参数集合结构体
func (r *ETHRPCRequester) ETHCall(result interface{}, arg model.CallArg) error {
	methodName := "eth_call"
	err := r.client.client.Call(result, methodName, arg, "latest")
	if err != nil {
		return fmt.Errorf("eth_call failed! %s", err.Error())
	}
	return nil
}

// 创建以太坊钱包
func (r *ETHRPCRequester) CreateETHWallet(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cant empty")
	}
	if len(password) < 6 {
		return "", errors.New("password's len must more than 6 words")
	}
	keydir := "./keystores" // 用来存储所创建的钱包的 keystore 文件的文件夹
	// StandardScryptN 是 Scrypt 加密算法的标准 N 参数
	// StandardScryptP 是 Scrypt 加密算法的标准 P 参数
	ks := keystore.NewKeyStore(keydir, keystore.StandardScryptN, keystore.StandardScryptP)
	wallet, err := ks.NewAccount(password) // 传入密码，创建钱包
	if err != nil {
		return "0x", err
	}
	return wallet.Address.String(), nil
}

// 发送交易，根据传入 transaction 的不同变量设置，达到发送不同种类的交易
func (r *ETHRPCRequester) SendTransaction(address string, transaction *types.Transaction) (string, error) {
	// 对交易数据进行签名
	signTx, err := tool.SignETHTransaction(address, transaction)
	if err != nil {
		return "", fmt.Errorf("签名失败！ %s", err.Error())
	}
	// rlp 序列化
	txRlpData, err := rlp.EncodeToBytes(signTx)
	if err != nil {
		return "", fmt.Errorf("rlp 序列号失败！ %s", err.Error())
	}
	// 下面调用以太坊的 rpc 接口
	txHash := ""
	methodName := "eth_sendRawTransaction"
	err = r.client.client.Call(&txHash, methodName, common.Bytes2Hex(txRlpData))
	if err != nil {
		return "", fmt.Errorf("发送交易失败！ %s", err.Error())
	}
	oldNonce := r.nonceManager.GetNonce(address)
	if oldNonce == nil {
		r.nonceManager.SetNonce(address, new(big.Int).SetUint64(transaction.Nonce()))
	}
	r.nonceManager.PlusNonce(address) // 成功后，当前用户内存的 nonce 值加 1
	return txHash, nil                // 返回交易hash
}

// 获取地址的 noce 值
func (r *ETHRPCRequester) GetNonce(address string) (uint64, error) {
	methodName := "eth_getTransactionCount" // 指定接口名称
	nonce := ""
	// 因为我们要查询最新的，根据基于 eth_getTransactionCount 情况下的区块号关系，选取 pending
	err := r.client.client.Call(&nonce, methodName, address, "pending")
	if err != nil {
		return 0, fmt.Errorf("发送交易失败！ %s", err.Error())
	}
	n, _ := new(big.Int).SetString(nonce[2:], 16) // 16进制转10进制
	return n.Uint64(), nil
}

// 发送 ETH 交易，或称转账 ETH
// 参数分别是交易发起地址、交易接收地址、ETH数量、燃料费设置
func (r *ETHRPCRequester) SendETHTransaction(fromStr, toStr, valueStr string, gasLimit, gasPrice uint64) (string, error) {
	if !common.IsHexAddress(fromStr) || !common.IsHexAddress(toStr) {
		return "", errors.New("invalid address")
	}

	to := common.HexToAddress(toStr) // 将字符串类型转为 address 类型
	gasPrice_ := new(big.Int).SetUint64(gasPrice)

	// value 乘上 10^decimal，得出真实的转账值，ETH 单位精确到小数点后 18 位
	realV := tool.GetRealDecimalValue(valueStr, 18)
	if realV == "" {
		return "", errors.New("invalid value")
	}
	amount, _ := new(big.Int).SetString(realV, 10)

	// 获取 nonce
	nonce := r.nonceManager.GetNonce(fromStr)
	if nonce == nil {
		// nonce 不存在，开始访问节点获取
		n, err := r.GetNonce(fromStr)
		if err != nil {
			return "", fmt.Errorf("获取 nonce 失败 %s", err.Error())
		}
		nonce = new(big.Int).SetUint64(n)
		r.nonceManager.SetNonce(fromStr, nonce) // 为当前的地址设置 nonce
	}
	// 构建 data，因为 eth 是交易转账类型，所有 data 是空的，我们设置空字符串即可
	data := []byte("")

	// 构建交易结构体
	transaction := types.NewTransaction(
		nonce.Uint64(),
		to,
		amount,
		gasLimit,
		gasPrice_,
		data)
	return r.SendTransaction(fromStr, transaction)
}

// 发送 ERC20 代币交易，或称转账 ERC20 代币
// 参数分别是
// 交易的发起地址、代币的合约地址、交易接受地址、代币数量、燃料费设置、代币的 decimal 值
func (r *ETHRPCRequester) SendERC20Transaction(fromStr, contact, receiver, valueStr string, gasLimit, gasPrice uint64, decimal int) (string, error) {
	if !common.IsHexAddress(fromStr) || !common.IsHexAddress(contact) || !common.IsHexAddress(receiver) {
		return "", errors.New("invalid address")
	}

	to := common.HexToAddress(contact) // 将合约 contact 字符串类型转为 address 类型
	gasPrice_ := new(big.Int).SetUint64(gasPrice)

	// 结构体中的 value 字段为 0
	amount := new(big.Int).SetInt64(0)

	// 获取 nonce
	nonce := r.nonceManager.GetNonce(fromStr)
	if nonce == nil {
		// nonce 不存在，开始访问节点获取
		n, err := r.GetNonce(fromStr)
		if err != nil {
			return "", fmt.Errorf("获取 nonce 失败 %s", err.Error())
		}
		nonce = new(big.Int).SetUint64(n)
		r.nonceManager.SetNonce(fromStr, nonce) // 为当前的地址设置 nonce
	}

	// 构建 data，真实的 value 转账数值由 data 携带
	data := tool.BuildERC20TransferData(valueStr, receiver, decimal)
	dataBytes := common.FromHex(data) // 使用以太坊提供的函数将16进制转为字节

	// 构建交易结构体
	transaction := types.NewTransaction(
		nonce.Uint64(),
		to,
		amount,
		gasLimit,
		gasPrice_,
		dataBytes)
	return r.SendTransaction(fromStr, transaction)
}
