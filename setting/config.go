package setting

import (
	"blog/global"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type conf struct {
}

func (conf) Init() {
	// 构建配置文件路径
	configPath := filepath.Join(global.RootDir, "config", "app", "config.yaml")

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config file: %v", err))
	}

	// 解析 YAML 配置
	err = yaml.Unmarshal(data, global.Config)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse config file: %v", err))
	}

	// 验证必要的配置项
	if global.Config.App.Port == "" {
		panic("App.Port is required in config file")
	}
	if global.Config.MySQL.DSN == "" {
		panic("MySQL.DSN is required in config file")
	}

	fmt.Printf("Config loaded successfully from: %s\n", configPath)
}
