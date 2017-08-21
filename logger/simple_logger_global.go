package logger

import (
	"io"
	"log"
	"os"
)

var simpleLg = NewDefaultSimpleLogger()

func Enable() {
	simpleLg.disable = false
}

func Disable() {
	simpleLg.disable = true
}

func SetLogFile(file string) error {
	fp, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	log.Println("set logfile:", file, err)
	if err != nil {
		return err
	}
	SetWriter(fp)
	return nil
}

func SetWriter(w io.WriteCloser) {
	simpleLg.SetWriter(w)
}

func SetLevel(lv Level) {
	simpleLg.level = lv
}

func SetTimeFormat(format string) {
	simpleLg.SetTimeFormat(format)
}

func EnableCallerInfo() {
	simpleLg.EnableCallerInfo()
}

func DisableCallerInfo() {
	simpleLg.DisableCallerInfo()
}

func Debug(format string, a ...interface{}) {
	simpleLg.Debug(format, a)
}

func Info(format string, a ...interface{}) {
	simpleLg.Info(format, a)
}

func Warn(format string, a ...interface{}) {
	simpleLg.Warn(format, a)
}

func Error(format string, a ...interface{}) {
	simpleLg.Error(format, a)
}

func Critical(format string, a ...interface{}) {
	simpleLg.Critical(format, a)
}

func Close() {
	simpleLg.Close()
}
