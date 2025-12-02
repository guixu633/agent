package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB 数据库连接实例
var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB(dsn string) error {
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 测试连接
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
