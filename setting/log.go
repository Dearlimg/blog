package setting

import (
	"blog/pkg/logger"
	"fmt"
)

type logConfig struct{}

func (logConfig) Init() {
	if err := logger.InitLogger(); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger.Info("Logger initialized successfully")
}
