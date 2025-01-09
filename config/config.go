package config

import (
	"os"
)

// Config 定义配置结构体
type Config struct {
	MySQL struct {
		Host     string `yaml:"host"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"mysql"`
	Redis struct {
		Host     string `yaml:"host"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
}

// GetConfig 函数读取 YAML 文件并解析为 Config 结构体
func GetConfig(filePath string) (*Config, error) {
	var config *Config

	// 读取文件内容
	data, err := os.Open(filePath)
	if err != nil {
		return config, err
	}

	// 解析 YAML 数据
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return config, err
	}

	return config, nil
}
