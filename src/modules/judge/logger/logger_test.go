package logger

import (
	"testing"
)

func Test_Logger(t *testing.T) {
	SetSeverity("DEBUG", 1)
	Debug(1, "hello")
	SetSeverity("INFO", 2)
	Debug(2, "hello")
	Infof(2, "%s", "hello")

	SetSeverity("WARNING")
	Debug(1, "hello")
	Info(2, "hello")
	Debug(0, "hello")
	Warning(0, "hello")
	Errorf(0, "%s", "hello")

	Close()
}
