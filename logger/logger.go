package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mutex sync.Mutex

	disable bool

	w io.WriteCloser
	// bw                 *bufio.Writer
	time_format        string
	level              Level
	enable_caller_info bool
	caller_path_number int

	logbuf chan []byte
}

func NewDefaultLogger() *Logger {
	l := &Logger{
		w: os.Stdout,
		// bw:                 bufio.NewWriter(os.Stdout),
		time_format:        default_time_format,
		level:              LevelDebug,
		enable_caller_info: true,
		caller_path_number: 3,

		logbuf: make(chan []byte, 200000),
	}
	go l.loop_write()
	return l
}

func NewLogger(wc io.WriteCloser) *Logger {
	l := &Logger{
		w: wc,
		// bw:          bufio.NewWriter(wc),
		time_format:        default_time_format,
		level:              LevelDebug,
		caller_path_number: 3,

		logbuf: make(chan []byte, 200000),
	}
	go l.loop_write()
	return l
}

func (l *Logger) Enable() {
	l.disable = false
}

func (l *Logger) Disable() {
	l.disable = true
}

func (l *Logger) SetWriter(w io.WriteCloser) {
	if w == nil {
		panic("SetWriter w is null")
	}

	l.mutex.Lock()
	old_w := l.w
	// old_bw := l.bw
	l.w = w
	// l.bw = bufio.NewWriterSize(w, 1024*1024)
	l.mutex.Unlock()

	// if old_w != nil && old_bw != nil {
	if old_w != nil {
		// old_bw.Flush()
		// old_bw = nil
		old_w.Close()
		old_w = nil
		// old_bw = nil
	}
}

func (l *Logger) SetLevel(lv Level) {
	l.level = lv
}

func (l *Logger) SetTimeFormat(format string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.time_format = format
}

func (l *Logger) EnableCallerInfo() {
	l.enable_caller_info = true
}

func (l *Logger) DisableCallerInfo() {
	l.enable_caller_info = false
}

func (l *Logger) write(level_str string, format string, a ...interface{}) {
	var s string

	if !l.enable_caller_info {
		s = fmt.Sprintf(time.Now().Format(l.time_format)+" ["+level_str+"] "+format+"\n", a...)
	} else {
		s = fmt.Sprintf(time.Now().Format(l.time_format)+" ["+level_str+"] ["+get_caller_info(l.caller_path_number).String()+"] "+format+"\n", a...)
	}

	l.logbuf <- StringToReadonlySlice(&s)
	// l.logbuf <- []byte(s)
	// l.mutex.Lock()
	// n, err := l.bw.Write([]byte(s))
	// l.mutex.Unlock()
	// log.Printf("xxxlog.Info %p, %p,  %t %v (%d:%v)", l.bw, l.w, l.disable, l.level, n, err)
}

func (l *Logger) loop_write() {
	var count int

	for {
		count = 0

		l.mutex.Lock()
	innerfor:
		for {
			select {
			case logentry := <-l.logbuf:
				count++
				// l.bw.Write(logentry)
				l.w.Write(logentry)
				if count > 1000 {
					l.mutex.Unlock()
					time.Sleep(time.Millisecond)
					break innerfor
				}
			default:
				l.mutex.Unlock()
				time.Sleep(time.Millisecond * 100)
				break innerfor
			}
		}

	}
}

func (l *Logger) Debug4(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelDebug4 {
		l.write("DEBUG4", format, a...)
	}
}

func (l *Logger) Debug3(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelDebug3 {
		l.write("DEBUG3", format, a...)
	}
}

func (l *Logger) Debug2(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelDebug2 {
		l.write("DEBUG2", format, a...)
	}
}

func (l *Logger) Debug1(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelDebug1 {
		l.write("DEBUG1", format, a...)
	}
}

func (l *Logger) Debug(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelDebug {
		l.write("DEBUG", format, a...)
	}
}

func (l *Logger) Info(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelInfo {
		l.write("INFO", format, a...)
	}
}

func (l *Logger) Warn(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelWarn {
		l.write("WARN", format, a...)
	}
}

func (l *Logger) Error(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelError {
		l.write("ERROR", format, a...)
	}
}

func (l *Logger) Critical(format string, a ...interface{}) {
	if l.disable {
		return
	}
	if l.level <= LevelCrit {
		l.write("CRITICAL", format, a...)
	}
}

func (l *Logger) Flush() error {
	return nil
	// l.mutex.Lock()
	// var err error
	// if l.bw != nil {
	// 	err = l.bw.Flush()
	// } else {
	// 	err = ErrEmptyLogWriter
	// }
	// l.mutex.Unlock()
	// return err
}

func (l *Logger) Close() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// l.bw.Flush()

	if l.w != nil {
		l.w.Close()
	}
}
