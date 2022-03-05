package dao

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"xorm.io/core"
)

type MySQLOptions struct {
	HostName           string // 数据库服务器域名
	Port               string // 端口
	User               string // 数据库用户
	Password           string // 数据库密码
	DbName             string //数据库名称
	TablePrefix        string // 数据库表前缀
	MaxOpenConnections int    // 数据库最大连接数
	MaxIdleConnections int    // 数据库最大空闲连接数
	ConnMaxLifetime    int    // 空闲连接多长时间被回收，单位为秒
}

// MySQL 连接器结构体
type MySQLConnector struct {
	options *MySQLOptions // 数据库配置结构指针
	tables  []interface{} // 数据库表的结构体集合
	Db      *xorm.Engine  // xorm 框架指针
}

// tables 是数据表的结构体实例数组
func NewMySQLConnector(options *MySQLOptions, tables []interface{}) MySQLConnector {
	var connector MySQLConnector
	connector.options = options
	connector.tables = tables
	// 设置数据库连接的 url
	url := ""
	if options.HostName == "" || options.HostName == "127.0.0.1" {
		url = fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True", options.User, options.Password, options.DbName)
	} else {
		url = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", options.User, options.Password, options.HostName, options.Port, options.DbName)
	}
	db, err := xorm.NewEngine("mysql", url) // 以 MySQL 数据可类型实例化
	if err != nil {
		panic(fmt.Errorf("数据库初始化失败 %s", err.Error()))
	}
	tbMapper := core.NewPrefixMapper(core.SnakeMapper{}, options.TablePrefix)
	db.SetTableMapper(tbMapper)
	db.DB().SetConnMaxLifetime(time.Duration(options.ConnMaxLifetime) * time.Second)
	db.DB().SetMaxIdleConns(options.MaxIdleConnections)
	db.DB().SetMaxOpenConns(options.MaxOpenConnections)
	// db.ShowSQL(true) // 是否开启打印 SQL 日志到控制台
	if err = db.Ping(); err != nil {
		panic(fmt.Errorf("数据库连接失败 %s", err.Error()))
	}
	connector.Db = db
	// 创建数据表，策略是不存在则创建
	if err := connector.createTables(); err != nil {
		panic(fmt.Errorf("创建数据表失败 %s", err.Error()))
	}
	return connector
}

// 创建数据表
func (s *MySQLConnector) createTables() error {
	if len(s.tables) == 0 {
		// 没有数据表则需要创建
		return nil
	}
	if err := s.Db.CreateTables(s.tables...); err != nil {
		return fmt.Errorf("create mysql table error: %s", err.Error())
	}
	// 同步数据表的修改
	if err := s.Db.Sync2(s.tables...); err != nil {
		return fmt.Errorf("sync table error: %s", err.Error())
	}
	return nil
}
