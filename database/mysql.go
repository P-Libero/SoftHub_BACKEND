package database

import (
	"context"
	"fmt"
	"softhub/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect(cfg config.Database) (*gorm.DB, error) {

	DSN := fmt.Sprintf(
		//这段内容是 MySQL 连接字符串（DSN）后面的连接参数，用于告诉 Go 的 MySQL 驱动应该怎样处理字符编码、时间和时区。
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)
	db, err := gorm.Open(mysql.Open(DSN), &gorm.Config{}) //返回一个包装的gorm对象 这时候在gorm里 连接池已经建好了
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	//获取连接池
	mysqlDB, err := db.DB() //返回标准库的sql.DB 也就是连接池的对象
	if err != nil {
		return nil, fmt.Errorf("get sql mysql db: %w", err)
	}
	mysqlDB.SetMaxIdleConns(5)
	mysqlDB.SetMaxOpenConns(10)
	mysqlDB.SetConnMaxIdleTime(time.Hour)
	//主动验证是否连接 5s为期
	ctx, canel := context.WithTimeout(context.Background(), 5*time.Second)
	defer canel()
	if err := mysqlDB.PingContext(ctx); err != nil {
		_ = mysqlDB.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}
