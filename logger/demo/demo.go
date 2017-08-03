package main

import (
	"os"

	"github.com/stormgbs/gopkg/logger"
)

func myfunc(l *logger.SimpleLogger) {
	l.Error("test call")
}

func main() {
	fp, err := os.OpenFile("demo.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}

	lg := logger.NewSimpleLogger(fp)

	lg.System = "DEMO"

	lg.SetLevel(logger.LevelError)
	lg.Debug("debug: %d-%d", 1, 100)
	lg.Info("info: %d", 2)
	lg.Error("error: %d", 4)

	lg.SetLevel(logger.LevelInfo)
	lg.Debug("debug: %d-%d", 1, 100)
	lg.Info("info: %d", 2)
	lg.Error("error: %d", 4)

	lg.SetTimeFormat("01/02 15:04:05")
	lg.EnableCallerInfo()
	lg.Info("info: %d", 2)

	lg.EnableCallerInfo()
	lg.Critical("testcriti")

	myfunc(lg)

	lg.Close()
}
