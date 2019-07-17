package config

import (
	"fmt"
	"os"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/str"
)

// CryptoPass crypto password use salt
func CryptoPass(raw string) string {
	return str.MD5(Get().Salt + "<-*Uk30^96dY*->" + raw)
}

// InitLogger init logger toolkits
func InitLogger() {
	c := Get().Logger

	lb, err := logger.NewFileBackend(c.Dir)
	if err != nil {
		fmt.Println("cannot init logger:", err)
		os.Exit(1)
	}

	lb.SetRotateByHour(true)
	lb.SetKeepHours(c.KeepHours)

	logger.SetLogging(c.Level, lb)
}
