package tool

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// 单元测试：生成 methodId
func Test_MakeMethodId(t *testing.T) {
	// 加法智能合约的 abi 数据
	contractABI :=
		`[{"constant": true, "inputs": [{"name": "arg1", "type": "uint8"},
	{"name": "arg2", "type": "uint8"}], 
	"name": "add", "outputs":[{"name": "", "type": "uint8"}],
	"payable": false, "stateMutability": "pure", "type": "function"}]`
	methodName := "add" // 加法函数名称
	methodId, err := MakeMethodId(methodName, contractABI)
	if err != nil {
		fmt.Println("生成 methodId 失败", err.Error())
		return
	}
	fmt.Println("生成 methodId 成功", methodId)
}

func Test_UnlockETHWallet(t *testing.T) {
	address := "0x97376Cf11717ab4A9e9a94042e895640a6262e30"
	keyDir := "../keystores"
	// 第一次演示密码错误的情况
	err1 := UnlockETHWallet(keyDir, address, "789")
	if err1 != nil {
		fmt.Println("第一次解锁错误：", err1.Error())
	} else {
		fmt.Println("第一次解锁成功!")
	}
	err2 := UnlockETHWallet(keyDir, address, "123456")
	if err2 != nil {
		fmt.Println("第一次解锁错误：", err1.Error())
	} else {
		fmt.Println("第一次解锁成功!")
	}
	// 下面是签名的测试
	// 创建一个测试用的交易数据结构体
	tx := types.NewTransaction(
		123,                       // nonce 交易序列号
		common.Address{},          // to 接收者地址
		new(big.Int).SetInt64(10), // value 数值
		1000,                      // gasLimit
		new(big.Int).SetInt64(20), // gasPrice
		[]byte("交易"))              // data
	signTx, err := SignETHTransaction(address, tx)
	if err != nil {
		fmt.Println("签名失败!", err.Error())
		return
	}
	data, _ := json.Marshal(signTx)
	fmt.Println("签名成功\n", string(data))
}

