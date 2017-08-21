package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type SimpleLogger struct {
	mutex sync.Mutex

	disable            bool
	w                  io.WriteCloser
	time_format        string
	level              Level
	enable_caller_info bool
	caller_path_number int

	System string
}

func NewDefaultSimpleLogger() *SimpleLogger {
	l := &SimpleLogger{
		w:                  os.Stdout,
		time_format:        default_time_format,
		level:              LevelDebug,
		enable_caller_info: true,
		caller_path_number: 3,
	}
	return l
}

func NewSimpleLogger(wc io.WriteCloser) *SimpleLogger {
	l := &SimpleLogger{
		w:                  wc,
		time_format:        default_time_format,
		level:              LevelDebug,
		caller_path_number: 3,
	}
	return l
}

func (l *SimpleLogger) Enable() {
	l.disable = false
}

func (l *SimpleLogger) Disable() {
	l.disable = true
}

func (l *SimpleLogger) SetWriter(w io.WriteCloser) {
	if w == nil {
		panic("SetWriter w is null")
	}

	l.mutex.Lock()
	old_w := l.w
	l.w = w
	l.mutex.Unlock()

	if old_w != nil {
		old_w.Close()
		old_w = nil
	}
}

func (l *SimpleLogger) SetLevel(lv Level) {
	l.level = lv
}

func (l *SimpleLogger) SetTimeFormat(format string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.time_format = format
}

func (l *SimpleLogger) EnableCallerInfo() {
	l.enable_caller_info = true
}

func (l *SimpleLogger) DisableCallerInfo() {
	l.enable_caller_info = false
}

func (l *SimpleLogger) write(level_str string, format string, a ...interface{}) {
	var s string

	if !l.enable_caller_info {
		if l.System != "" {
			s = fmt.Sprintf(time.Now().Format(l.time_format)+" ["+level_str+"] |"+l.System+"| "+format+"\n", a...)
		} else {
			s = fmt.Sprintf(time.Now().Format(l.time_format)+" ["+level_str+"] |"+l.System+"| "+format+"\n", a...)
		}
	} else {
		if l.System != "" {
			s = fmt.Sprintf(time.Now().Format(l.time_format)+" ["+level_str+"] |"+l.System+"| ["+get_caller_info(l.caller_path_number).String()+"] "+format+"\n", a...)
		} else {
			s = fmt.Sprintf(time.Now().Format(l.time_format)+" ["+level_str+"] |"+l.System+"| ["+get_caller_info(l.caller_path_number).String()+"] "+format+"\n", a...)
		}

	}

	l.w.Write(StringToReadonlySlice(&s))
}

func (l *SimpleLogger) Debug(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelDebug {
		l.write("DEBUG", format, a...)
	}
}

func (l *SimpleLogger) Info(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelInfo {
		l.write("INFO", format, a...)
	}
}

func (l *SimpleLogger) Warn(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelWarn {
		l.write("WARN", format, a...)
	}
}

func (l *SimpleLogger) Error(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelError {
		l.write("ERROR", format, a...)
	}
}

func (l *SimpleLogger) Critical(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelCrit {
		l.write("CRITICAL", format, a...)
	}
}

func (l *SimpleLogger) Close() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.w != nil {
		l.w.Close()
	}
}
