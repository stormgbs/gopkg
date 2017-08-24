package logger

import (
	"fmt"
	"sync"
	"time"
)

//Mon Jan 2 15:04:05 -0700 MST 2006

type LogFormator struct {
	mutex              sync.Mutex
	disable            bool
	time_format        string
	level              Level
	enable_caller_info bool
	caller_path_number int

	System string
}

func NewDefaultLogFormator() *LogFormator {
	l := &LogFormator{
		time_format:        default_time_format,
		level:              LevelDebug,
		enable_caller_info: false,
		caller_path_number: 3,
	}
	return l
}

func (l *LogFormator) Enable() {
	l.disable = false
}

func (l *LogFormator) Disable() {
	l.disable = true
}

func (l *LogFormator) SetLevel(lv Level) {
	l.level = lv
}

func (l *LogFormator) SetTimeFormat(format string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.time_format = format
}

func (l *LogFormator) EnableCallerInfo() {
	l.enable_caller_info = true
}

func (l *LogFormator) DisableCallerInfo() {
	l.enable_caller_info = false
}

func (l *LogFormator) format(level_str string, format string, a ...interface{}) []byte {
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

	return StringToReadonlySlice(&s)
}

func (l *LogFormator) Debug4(format string, a ...interface{}) []byte {
	if l.disable {
		return nil
	}
	if l.level <= LevelDebug4 {
		return l.format("DEBUG4", format, a...)
	}
	return nil
}

func (l *LogFormator) Debug3(format string, a ...interface{}) []byte {
	if l.disable {
		return nil
	}
	if l.level <= LevelDebug3 {
		return l.format("DEBUG3", format, a...)
	}
	return nil
}

func (l *LogFormator) Debug2(format string, a ...interface{}) []byte {
	if l.disable {
		return nil
	}
	if l.level <= LevelDebug2 {
		return l.format("DEBUG2", format, a...)
	}
	return nil
}

func (l *LogFormator) Debug1(format string, a ...interface{}) []byte {
	if l.disable {
		return nil
	}
	if l.level <= LevelDebug1 {
		return l.format("DEBUG1", format, a...)
	}
	return nil
}

func (l *LogFormator) Debug(format string, a ...interface{}) []byte {
	if l.disable {
		return nil
	}
	if l.level <= LevelDebug {
		return l.format("DEBUG", format, a...)
	}
	return nil
}

func (l *LogFormator) Info(format string, a ...interface{}) []byte {
	if l.disable {
		return nil
	}
	if l.level <= LevelInfo {
		return l.format("INFO", format, a...)
	}

	return nil
}

func (l *LogFormator) Warn(format string, a ...interface{}) []byte {
	if l.disable {
		return nil
	}
	if l.level <= LevelWarn {
		return l.format("WARN", format, a...)
	}
	return nil
}

func (l *LogFormator) Error(format string, a ...interface{}) []byte {
	if l.disable {
		return nil
	}
	if l.level <= LevelError {
		return l.format("ERROR", format, a...)
	}
	return nil
}

func (l *LogFormator) Critical(format string, a ...interface{}) []byte {
	if l.disable {
		return nil
	}
	if l.level <= LevelCrit {
		l.format("CRITICAL", format, a...)
	}
	return nil
}
