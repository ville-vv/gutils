package dbs

import "testing"

func TestInitMysqlDB(t *testing.T) {
	mysqlCfg := MySqlConfig{
		MainDns:  "root:Root1234.@tcp(192.168.229.132:3306)/relay?charset=utf8mb4&parseTime=True&loc=Local",
		LogLevel: "info",
	}
	_ = InitMysqlDB(mysqlCfg)

}
