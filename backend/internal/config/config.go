package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 统一配置结构
type Config struct {
	OSS     OSSConfig     `json:"oss"`
	Postgres PostgresConfig `json:"postgres"`
}

// OSSConfig OSS 配置
type OSSConfig struct {
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	ImagePrefix     string `json:"image_prefix"`
}

// PostgresConfig PostgreSQL 配置
type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
	TimeZone string `json:"timezone"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "configs/config.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// GetDSN 获取 PostgreSQL 连接字符串
func (c *PostgresConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode, c.TimeZone)
}

