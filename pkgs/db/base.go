package db

// db 的 orm 使用 gorm v2 , 具体使用方法参考链接 :
// https://www.kancloud.cn/sliver_horn/gorm/1861153

import (
	"database/sql"
	"fmt"
	"night-fury/pkgs/log"
	"os"
	"time"

	"github.com/pkg/errors"
	"gitlab.lanhuapp.com/gopkgs/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	connString string
	db         *gorm.DB
	Nil        = gorm.ErrRecordNotFound
)

type Model struct {
	ID        string     `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

func init() {
	initConnString()
	initDB()
	migrate()
}

func initConnString() {
	host := config.GetString("database.host")
	port := config.GetString("database.port")
	user := config.GetString("database.user")
	pass := config.GetString("database.pass")
	name := config.GetString("database.name")
	sslmode := config.GetString("database.sslmode")
	connString = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", host, port, user, name, pass, sslmode)
}

func initDB() {
	var err error

	db, err = gorm.Open(postgres.Open(connString), &gorm.Config{})

	if err != nil {
		log.Fatalf(log.TagDB, "init db error : %s", err)
		return
	}

	if os.Getenv("ENV") == "local" { // 开发环境调试
		logger.Default.LogMode(logger.Info)
	} else { // 生产环境静默
		logger.Default.LogMode(logger.Silent)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(errors.Wrap(err, "init db error"))
	}

	// 设置连接参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Minute * 10)
}

func migrate() {
	err := db.AutoMigrate(&User{})
	if err != nil {
		panic(err)
	}
}

func GetStats() sql.DBStats {
	sqlDB, _ := db.DB()
	return sqlDB.Stats()
}

func PingDB() error {
	sqlDB, _ := db.DB()
	return sqlDB.Ping()
}

type SortCond struct {
	Field string
	Sort  string
}

func (c *SortCond) Stringify() string {
	return fmt.Sprintf("%s %s", c.Field, c.Sort)
}

func getDB(dbs ...*gorm.DB) *gorm.DB {
	if len(dbs) > 0 {
		return dbs[0]
	}
	return db
}

func initTransaction() (*gorm.DB, func(*error), error) {
	var dberr error

	tx := db.Begin()
	if dberr = tx.Error; dberr != nil {
		return nil, nil, dberr
	}
	finally := func(err *error) {
		if *err == nil {
			err = &(tx.Commit().Error)
		}
		if *err != nil {
			tx.Rollback()
		}
	}
	return tx, finally, nil
}

// GetDb 获取db实例
func GetDb(dbs ...*gorm.DB) *gorm.DB {
	return getDB(dbs...)
}

// InitTransaction 初始化事务
func InitTransaction() (*gorm.DB, func(*error), error) {
	return initTransaction()
}
