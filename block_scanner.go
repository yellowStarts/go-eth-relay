package main

import (
	"encoding/json"
	"errors"
	"eth-relay/dao"
	"eth-relay/model"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

// 区块遍历器
type BlockScanner struct {
	ethRequester ETHRPCRequester    // 以太坊 rpc 请求者对象
	mysql        dao.MySQLConnector // 数据库连接器对象
	lastBlock    *dao.Block         // 用来存储每次遍历后上一次的区块
	lastNumber   *big.Int           // 上一次区块的区块号
	fork         bool               // 区块分叉标记位
	stop         chan bool          // 用来控制是否停止遍历的管道
	lock         sync.Mutex         // 互斥锁，控制并发
}

// 实例化 区块遍历器
func NewBlockScanner(requester ETHRPCRequester, mysql dao.MySQLConnector) *BlockScanner {
	return &BlockScanner{
		ethRequester: requester,
		mysql:        mysql,
		lastBlock:    &dao.Block{},
		fork:         false,
		stop:         make(chan bool),
		lock:         sync.Mutex{},
	}
}

// 整个区块扫码的启动函数
func (scanner *BlockScanner) Start() error {
	scanner.lock.Lock()
	init := func() error {
		// 寻找出上一次成功遍历的区块
		_, err := scanner.mysql.Db.
			Desc("create_time").
			Where("fork = ?", false).
			Get(scanner.lastBlock)
		if err != nil {
			return err
		}
		if scanner.lastBlock.BlockHash == "" {
			// 首次启动，从节点中获取，并初始化
			latestBlockNumber, err := scanner.ethRequester.GetLatestBlockNumber()
			if err != nil {
				return err
			}
			latestBlock, err := scanner.ethRequester.GetBlockInfoByNumber(latestBlockNumber)
			if err != nil {
				return err
			}
			if latestBlock.Number == "" {
				panic(latestBlockNumber.String())
			}
			scanner.lastBlock.BlockHash = latestBlock.Hash
			scanner.lastBlock.ParentHash = latestBlock.ParentHash
			scanner.lastBlock.BlockNumber = latestBlock.Number
			scanner.lastBlock.CreateTime = scanner.hexToTen(latestBlock.Timestamp).Int64()
			scanner.lastNumber = latestBlockNumber
		} else {
			scanner.lastNumber, _ = new(big.Int).SetString(scanner.lastBlock.BlockNumber, 10)
			// 下面加 1，因为上一次数据库存的是已经遍历完了的
			scanner.lastNumber.Add(scanner.lastNumber, new(big.Int).SetInt64(1))
		}
		return nil
	}
	if err := init(); err != nil {
		return err
	}
	execute := func() {
		if err := scanner.scan(); nil != err {
			scanner.log(err.Error())
			return
		}
		time.Sleep(1 * time.Second) // 延迟一秒开始下一轮
	}
	// 启动一个协程来遍历区块
	go func() {
		for {
			select {
			case <-scanner.stop: // 监听是否退出遍历
				scanner.log("finish block scanner!")
				return
			default:
				if !scanner.fork {
					execute()
					continue
				}
				if err := init(); err != nil {
					scanner.log(err.Error())
					return
				}
				scanner.fork = false
			}
		}
	}()
	return nil
}

// 公有函数，可以供外部调用来停止区块遍历
func (scanner *BlockScanner) Stop() {
	scanner.lock.Unlock() // 解锁
	scanner.stop <- true
}

// 判断是否分叉的函数，若为 true 则是分叉
func (scanner *BlockScanner) isFork(currentBlock *dao.Block) bool {
	if currentBlock.BlockNumber == "" {
		panic("invalid block")
	}
	// scanner.lastBlock.BlockHash == currentBlock.ParentHash 判断上一次的区块哈希值是否是当前区块的父块哈希值
	if scanner.lastBlock.BlockHash == currentBlock.BlockHash || scanner.lastBlock.BlockHash == currentBlock.ParentHash {
		scanner.lastBlock = currentBlock // 没有发送分叉，更新上一次区块为当前被检测的区块
		return false
	}
	return true
}

// 获取分叉点区块
func (scanner *BlockScanner) getStartForkBlock(parentHash string) (*dao.Block, error) {
	// 获取当前区块的父区块,分叉从父区块开始
	parent := dao.Block{} // 定义一个 block 结构体实例，用来存储从数据库查询出的区块信息
	// 下面使用 xorm 框架提供的函数，根据 block_hash 去数据库获取区块信息，等同于 SQL语句：
	// select * from eth_block where block_hash=parentHash limit 1;
	_, err := scanner.mysql.Db.Where("block_hash=?", parentHash).Get(&parent)
	if err == nil && parent.BlockNumber != "" {
		return &parent, nil // 本地存在，直接返回分叉点区块
	}
	// 数据库没有父区块，准备从以太坊接口获取
	parentFull, err := scanner.retryGetBlockInfoByHash(parentHash)
	if err != nil {
		return nil, fmt.Errorf("分叉严重错误，需要重启区块扫码 %s", err.Error())
	}
	// 继续递归往上查询，直到数据库中有它的记录
	return scanner.getStartForkBlock(parentFull.ParentHash)
}

// 输出日志
func (scanner *BlockScanner) log(args ...interface{}) {
	fmt.Println(args...)
}

// 区块号存在，信息获取为空，可能是以太坊网络延时问题，重拾策略函数
func (scanner *BlockScanner) retryGetBlockInfoByNumber(targetNumber *big.Int) (*model.FullBlock, error) {

Retry:
	// 下面调用以太坊请求者 ethRequester 的 GetBlockInfoByNumber 函数
	fullBlock, err := scanner.ethRequester.GetBlockInfoByNumber(targetNumber)
	if err != nil {
		errInfo := err.Error()
		if strings.Contains(errInfo, "empty") {
			// 区块号存在，信息获取为空，可能是以太坊网络延时问题，直接重试
			scanner.log("获取区块信息，重试一次....", targetNumber.String())
			goto Retry
		}
		return nil, err
	}
	return fullBlock, nil
}

// 区块哈希存在，信息获取为空，可能是以太坊网络延时问题，重拾策略函数
func (scanner *BlockScanner) retryGetBlockInfoByHash(hash string) (*model.FullBlock, error) {

Retry:
	// 下面调用以太坊请求者 ethRequester 的 GetBlockInfoByHash 函数
	fullBlock, err := scanner.ethRequester.GetBlockInfoByHash(hash)
	if err != nil {
		errInfo := err.Error()
		if strings.Contains(errInfo, "empty") {
			// 区块哈希存在，信息获取为空，可能是以太坊网络延时问题，直接重试
			scanner.log("获取区块信息，重试一次....", hash)
			goto Retry
		}
		return nil, err
	}
	return fullBlock, nil
}

// 初始化，内部再开始遍历时赋值 lastBlock
func (scanner *BlockScanner) init() error {
	// 下面使用 xorm 提供的数据库函数从
	// 数据库中寻找出上一次成功遍历的且不是分叉的区块
	// 等同于SQL: select 8 from eth_block where fork=false order by create_time desc limit 1;
	_, err := scanner.mysql.Db.
		Desc("create_time"). // 根据时间排序
		Where("fork = ?", false).
		Get(scanner.lastBlock)
	if err != nil {
		return err
	}
	if scanner.lastBlock.BlockHash == "" {
		//区块哈希为空, 说明是整个程序的首次启动，那么从节点中获取最新生成的区块
		//GetLatestBlockNumber 获取最新区块的区块号
		latestBlockNumber, err := scanner.ethRequester.GetLatestBlockNumber()
		if err != nil {
			return err
		}
		// GetBlockInfoByNumber 根据区块好获取区块数据
		latestBlock, err := scanner.ethRequester.GetBlockInfoByNumber(latestBlockNumber)
		if err != nil {
			return err
		}
		if latestBlock.Number == "" {
			panic(latestBlockNumber.String())
		}
		// 下面是给区块遍历器的 lastBlock 变量赋值
		scanner.lastBlock.BlockHash = latestBlock.Hash
		scanner.lastBlock.ParentHash = latestBlock.ParentHash
		scanner.lastBlock.BlockNumber = latestBlock.Number
		scanner.lastBlock.CreateTime = scanner.hexToTen(latestBlock.Timestamp).Int64()
		scanner.lastNumber = latestBlockNumber
	} else {
		// 区块哈希值不为空，说明不是首次启动，而是后续的启动
		scanner.lastNumber, _ = new(big.Int).SetString(scanner.lastBlock.BlockNumber, 10)
		// 下面加 1，因为上一次数据库存的是已经遍历完了的区块，接下来是它的下一个区块
		scanner.lastNumber.Add(scanner.lastNumber, new(big.Int).SetInt64(1))
	}
	return nil
}

// 定义一个将 16 进制转为 10 进制的函数
func (scanner *BlockScanner) hexToTen(hex string) *big.Int {
	if !strings.HasPrefix(hex, "0x") {
		ten, _ := new(big.Int).SetString(hex, 10) // 本事就是十进制字符串，直接设置
		return ten
	}
	ten, _ := new(big.Int).SetString(hex[2:], 16)
	return ten
}

// 获取要扫描的区块号
func (scanner *BlockScanner) getScannerBlockNumber() (*big.Int, error) {
	// 调用以太坊请求者 ethRequester 获取公链上最新生成的区块的区块号
	newBlockNumber, err := scanner.ethRequester.GetLatestBlockNumber()
	if err != nil {
		return nil, err
	}
	latestNumber := newBlockNumber
	// 下面使用 new 的形式初始化并设置值，不要直接赋值
	// 否则会和 lastNumber 的内存地址一样，影响后面的获取区块信息
	targetNumber := new(big.Int).Set(scanner.lastNumber)
	// 比较区块号大小
	// -1 if x < y, 0 if x==y,+1 if x > y
	if latestNumber.Cmp(scanner.lastNumber) < 0 {
		// 最新的区块高度比设置的要小，则等待新区块高度 >= 设置的
	Next:
		for {
			select {
			case <-time.After(time.Duration(4 * time.Second)): // 延时 4 秒重新获取
				number, err := scanner.ethRequester.GetLatestBlockNumber()
				if err == nil && number.Cmp(scanner.lastNumber) >= 0 {
					break Next // 跳出循环
				}
			}
		}
	}
	return targetNumber, nil // 返回目标区块高度
}

// 扫描区块
func (scanner *BlockScanner) scan() error {
	// 获取公链上最新生成的区块
	newBlockNumber, err := scanner.ethRequester.GetLatestBlockNumber()
	if err != nil {
		return err
	}

	latestNumber := newBlockNumber
	targetNumber := scanner.lastNumber
	// 比较区块号大小
	// -1 if x <  y
	//  0 if x == y
	// +1 if x >  y
	if latestNumber.Cmp(scanner.lastNumber) < 0 {
		// 小，则等待新区块生成
	Next:
		for {
			select {
			case <-time.After(time.Duration(4 * time.Second)):
				number, err := scanner.ethRequester.GetLatestBlockNumber()
				if err == nil && number.Cmp(scanner.lastNumber) >= 0 {
					targetNumber = number
					break Next
				}
			}
		}
	}
	// 获取区块信息
	fullBlock, err := scanner.retryGetBlockInfoByNumber(targetNumber)
	if err != nil {
		return err
	}
	// 区块号自增 1
	scanner.lastNumber.Add(scanner.lastNumber, new(big.Int).SetInt64(1))

	// 开启数据库事务
	tx := scanner.mysql.Db.NewSession()
	defer tx.Close()

	// 准备保存区块信息，先判断当前区块记录是否已经存在
	block := dao.Block{}
	_, err = tx.Where("block_hash=?", fullBlock.Hash).Get(&block)
	if err == nil && block.Id == 0 {
		// 不存在，进行添加
		block.BlockNumber = scanner.hexToTen(fullBlock.Number).String()
		block.ParentHash = fullBlock.ParentHash
		block.CreateTime = scanner.hexToTen(fullBlock.Timestamp).Int64()
		block.BlockHash = fullBlock.Hash
		block.Fork = false
		if _, err := tx.Insert(&block); err != nil {
			tx.Rollback() // 事务回滚
			return err
		}
	}
	// 检查区块是否分叉
	if scanner.forkCheck(&block) {
		data, _ := json.Marshal(fullBlock)
		scanner.log("分叉！", string(data))
		tx.Commit()
		scanner.fork = true // 发生分叉
		return errors.New("fork check")
	}

	// 解析区块内数据，读取内部的 “transactions” 交易信息，分析得出各种合约事件
	scanner.log("scan block start ==> ", "number: ", scanner.hexToTen(fullBlock.Number), "hash: ", fullBlock.Hash)
	for index, transaction := range fullBlock.Transactions {
		// 下面的打印操作模拟自定义处理。对于每条 tx，我们是完全可以进一步从里面提取信息的！
		scanner.log("tx hash ==> ", transaction.Hash)
		if index > 5 {
			// 控制只打印 5 条
			break
		}
	}
	scanner.log("scan block finish \n=================")
	// 数据库保存交易信息
	if _, err = tx.Insert(&fullBlock.Transactions); err != nil {
		tx.Rollback() // 事务回滚
		return err
	}
	return tx.Commit()
}

// 检测分叉，返回 true 是分叉
func (scanner *BlockScanner) forkCheck(currentBlock *dao.Block) bool {
	if currentBlock.BlockNumber == "" {
		panic("invalid block")
	}
	if scanner.lastBlock.BlockHash == currentBlock.BlockHash || scanner.lastBlock.BlockHash == currentBlock.ParentHash {
		scanner.lastBlock = currentBlock // 更新
		return false
	}
	// 获取出最初开始分叉的那个区块
	forkBlock, err := scanner.getForkBlock(currentBlock.ParentHash)
	if err != nil {
		panic(err)
	}
	scanner.lastBlock = forkBlock // 更新。从这个区块开始，其之后的都是分叉的

	// 修改数据库记录，将分叉区块标记好
	numberEnd := ""
	if strings.HasPrefix(currentBlock.BlockNumber, "0x") {
		c, _ := new(big.Int).SetString(currentBlock.BlockNumber[2:], 16)
		numberEnd = c.String()
	} else {
		c, _ := new(big.Int).SetString(currentBlock.BlockNumber, 10)
		numberEnd = c.String()
	}
	numberFrom := forkBlock.BlockNumber
	_, err = scanner.mysql.Db.
		Table(dao.Block{}).
		Where("block_number > ? and block_number <= ?", numberFrom, numberEnd). // 区块号范围内
		Update(map[string]bool{"fork": true})
	if err != nil {
		panic(fmt.Errorf("update fork block failed %s", err.Error()))
	}
	return true
}

func (scanner *BlockScanner) getForkBlock(parentHash string) (*dao.Block, error) {
	// 获取当前区块的父区块，分叉从父区块开始
	parent := dao.Block{}
	_, err := scanner.mysql.Db.Where("block_hash=?", parentHash).Get(&parent)
	if err == nil && parent.BlockNumber != "" {
		return &parent, nil
	}
	// 数据库没有父区块记录，准备从以太坊接口获取
	parentFull, err := scanner.retryGetBlockInfoByHash(parentHash)
	if err != nil {
		return nil, fmt.Errorf("分叉严重错误，需要重启区块扫描 %s", err.Error())
	}
	// 继续递归往上查询，直到在数据库中有它的记录
	return scanner.getForkBlock(parentFull.ParentHash)
}
