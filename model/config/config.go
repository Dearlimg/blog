package config

type Config struct {
	App    AppConfig    `yaml:"app"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
	Log    LogConfig    `yaml:"log"`
	Ollama OllamaConfig `yaml:"ollama"`
	Eino   EinoConfig   `yaml:"eino"`
}

type AppConfig struct {
	Name      string `yaml:"name"`
	Version   string `yaml:"version"`
	StartTime string `yaml:"start_time"` // 使用字符串类型，避免解析问题
	Port      string `yaml:"port"`
}

type MySQLConfig struct {
	DSN string `yaml:"dns"` // 注意 yaml 标签对应配置中的 "dns"
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type LogConfig struct {
	Level      string `yaml:"level"`       // 日志级别: debug, info, warn, error
	Format     string `yaml:"format"`      // 日志格式: json, text
	Output     string `yaml:"output"`      // 输出位置: stdout, file, both
	FilePath   string `yaml:"file_path"`   // 日志文件路径（当 output 为 file 或 both 时）
	MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小（MB）
	MaxBackups int    `yaml:"max_backups"` // 保留的备份文件数量
	MaxAge     int    `yaml:"max_age"`     // 保留日志文件的天数
	Compress   bool   `yaml:"compress"`    // 是否压缩旧日志文件
}

type OllamaConfig struct {
	BaseURL string `yaml:"base_url"` // Ollama API 基础URL
	Model   string `yaml:"model"`    // 默认模型
	Timeout int    `yaml:"timeout"`  // 请求超时时间（秒）
}

type EinoConfig struct {
	MaxHistory int `yaml:"max_history"` // 最大历史记录数
}
