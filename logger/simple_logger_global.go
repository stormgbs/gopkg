package logger

import (
	"io"
	"os"
)

var simpleLg = NewDefaultSimpleLogger()

func init() {
	simpleLg.caller_path_number = 4
}

func Enable() {
	simpleLg.disable = false
}

func Disable() {
	simpleLg.disable = true
}

func SetLogFile(file string) error {
	fp, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		return err
	}
	SetWriter(fp)
	return nil
}

func SetWriter(w io.WriteCloser) {
	if w == nil {
		panic("SetWriter w is null")
	}

	simpleLg.mutex.Lock()
	old_w := simpleLg.w
	simpleLg.w = w
	simpleLg.mutex.Unlock()

	if old_w != nil {
		old_w.Close()
		old_w = nil
	}
}

func SetLevel(lv Level) {
	simpleLg.level = lv
}

func SetTimeFormat(format string) {
	simpleLg.mutex.Lock()
	defer simpleLg.mutex.Unlock()
	simpleLg.time_format = format
}

func EnableCallerInfo() {
	simpleLg.enable_caller_info = true
}

func DisableCallerInfo() {
	simpleLg.enable_caller_info = false
}

func Debug(format string, a ...interface{}) {
	if simpleLg.disable {
		return
	}

	if simpleLg.level <= LevelDebug {
		simpleLg.write("DEBUG", format, a...)
	}
}

func Info(format string, a ...interface{}) {
	if simpleLg.disable {
		return
	}
	if simpleLg.level <= LevelInfo {
		simpleLg.write("INFO", format, a...)
	}
}

func Warn(format string, a ...interface{}) {
	if simpleLg.disable {
		return
	}
	if simpleLg.level <= LevelWarn {
		simpleLg.write("WARN", format, a...)
	}
}

func Error(format string, a ...interface{}) {
	if simpleLg.disable {
		return
	}
	if simpleLg.level <= LevelError {
		simpleLg.write("ERROR", format, a...)
	}
}

func Critical(format string, a ...interface{}) {
	if simpleLg.disable {
		return
	}
	if simpleLg.level <= LevelCrit {
		simpleLg.write("CRITICAL", format, a...)
	}
}

func Close() {
	simpleLg.mutex.Lock()
	defer simpleLg.mutex.Unlock()

	if simpleLg.w != nil {
		simpleLg.w.Close()
	}
}
