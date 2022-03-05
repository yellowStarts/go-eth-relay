package dao

import (
	"fmt"
	"testing"
)

// 测试数据库连接
func Test_NewMySQLConnector(t *testing.T) {
	options := MySQLOptions{
		HostName:           "127.0.0.1", // 本地数据库
		Port:               "3306",      // 默认端口
		DbName:             "eth_reply", // 数据库名称
		User:               "root",      // 用户名
		Password:           "",          // 密码
		TablePrefix:        "eth_",      // 数据表前缀
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    15,
	}
	tables := []interface{}{}                       // 不创建数据表
	tables = append(tables, Block{}, Transaction{}) // 添加数据表的数据结构体
	mysql := NewMySQLConnector(&options, tables)
	if mysql.Db.Ping() == nil {
		fmt.Println("数据库连接成功")
	} else {
		fmt.Println("数据库连接成功")
	}
}
