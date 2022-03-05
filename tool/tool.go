package tool

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// 根据代币的 decimal 得出乘上 10^decimal 后的值
// value 是包含浮点数的，例如 0.5 个 ETH
func GetRealDecimalValue(value string, decimal int) string {
	if strings.Contains(value, ".") {
		// 小数
		arr := strings.Split(value, ".")
		if len(arr) != 2 {
			return ""
		}
		num := len(arr[1])
		left := decimal - num
		return arr[0] + arr[1] + strings.Repeat("0", left)
	} else {
		// 整数
		return value + strings.Repeat("0", decimal)
	}
}

// 构建符合 ERC20 标准的 transfer 合约函数的 data 入参
func BuildERC20TransferData(value, receiver string, decimal int) string {
	realValue := GetRealDecimalValue(value, decimal) // 将 value 乘上 10^decimal的格式
	valueBig, _ := new(big.Int).SetString(realValue, 10)

	// 构建
	methodId := "0xa9059cbb"                                    // "0xa9059cbb" 是 transfer 的 methodId
	param1 := common.HexToHash(receiver).String()[2:]           // 第一个参数，收款者地址
	param2 := common.BytesToHash(valueBig.Bytes()).String()[2:] // 第二个参数，交易的数值
	return methodId + param1 + param2
}
