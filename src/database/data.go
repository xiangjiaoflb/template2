package database

import (
	"github.com/jinzhu/gorm"

	//加载sql库
	_ "github.com/go-sql-driver/mysql"
)

//
var (
	DB *gorm.DB
)

//Open 设置数据库连接
func Open(sqllink string) (db *gorm.DB, err error) {
	db, err = gorm.Open("mysql", sqllink)
	if DB != nil {
		DB.Close()
	}
	DB = db

	return
}
