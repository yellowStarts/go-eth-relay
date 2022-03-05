package tool

import (
	"errors"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// 全局地保存了已经解锁成功的钱包 map 集合变量
var ETHUnlockMap map[string]accounts.Account

// 全局地对应 keystore 实例
var UnlockKs *keystore.KeyStore

// 解锁以太坊钱包，传入钱包地址和对应的 keystore 密码
func UnlockETHWallet(keysDir string, address, password string) error {
	if UnlockKs == nil {
		UnlockKs = keystore.NewKeyStore(
			// 服务端存储 keystore 文件的文件夹
			// 这些配置类的信息可以由配置文件指定
			keysDir,
			keystore.StandardScryptN,
			keystore.StandardScryptP)
		if UnlockKs == nil {
			return errors.New("ks is nil")
		}
	}
	unlock := accounts.Account{Address: common.HexToAddress(address)}
	// ks.Unlock 调用 keystore.go 的解锁函数，解锁出的私钥将存储再它里面的变量中
	if err := UnlockKs.Unlock(unlock, password); nil != err {
		return errors.New("unlock err : " + err.Error())
	}
	if ETHUnlockMap == nil {
		ETHUnlockMap = map[string]accounts.Account{}
	}
	ETHUnlockMap[address] = unlock // 解锁成功，存储
	return nil
}

// 根据函数的名称生成 methodId。abiStr 是智能合约的 "abi" 数据
func MakeMethodId(methodName string, abiStr string) (string, error) {
	abi := &abi.ABI{} // 实例化 "ABI" 结构体对象指针
	err := abi.UnmarshalJSON([]byte(abiStr))
	if err != nil {
		return "", err
	}
	// 根据 methodName 获取对应的 Method 对象
	method := abi.Methods[methodName]
	methodIdBytes := method.ID // 调用生成 methodId 的函数
	methodId := "0x" + common.Bytes2Hex(methodIdBytes)
	return methodId, nil
}

// type txData struct {
// 	AccountNonce uint64   `json:"nonce" gencodec:"required"`    // 交易序列号
// 	Price        *big.Int `json:"gasPrice" gencodec:"required"` // gasPrice
// 	GasLimit     uint64   `json:"gas" gencodec:"required"`      // gasLimit
// 	// to 交易的接收者地址，"nil means contract creation"的意思是，空意味着创建智能合约
// 	Recipient *common.Address `json:"to" rlp:"nil"`              // nil means contract creation
// 	Amount    *big.Int        `json:"value" gencodec:"required"` // 要交易的代币数值
// 	Payload   []byte          `json:"input" gencodec:"required"` // data 参数

// 	// Signature values
// 	// 下面的v,r,s签名时会赋值，其中保存的是签名后生成的数据
// 	V *big.Int `json:"v" gencodec:"required"`
// 	R *big.Int `json:"r" gencodec:"required"`
// 	S *big.Int `json:"s" gencodec:"required"`

// 	Hash *common.Hash `json:"hash" rlp:"-"`
// }

// 对交易数据结构体 types.Transaction 进行签名
func SignETHTransaction(address string, transaction *types.Transaction) (*types.Transaction, error) {
	if UnlockKs == nil {
		return nil, errors.New("you need to init keystore first")
	}
	account := ETHUnlockMap[address]
	if !common.IsHexAddress(account.Address.String()) {
		// 判断当前的地址钱包是否解锁了
		return nil, errors.New("account need to unlock first")
	}
	return UnlockKs.SignTx(account, transaction, nil) // 调用签名函数
}
