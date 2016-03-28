package logger

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type Level uint8

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelCrit
)

//Mon Jan 2 15:04:05 -0700 MST 2006
const default_time_format = "2006/01/02 15:04:05.000"

type Logger struct {
	mutex sync.Mutex

	disable bool

	w                  io.WriteCloser
	bw                 *bufio.Writer
	time_format        string
	level              Level
	enable_caller_info bool

	logbuf chan []byte
}

func NewDefaultLogger() *Logger {
	l := &Logger{
		w:                  os.Stdout,
		bw:                 bufio.NewWriter(os.Stdout),
		time_format:        default_time_format,
		level:              LevelDebug,
		enable_caller_info: true,

		logbuf: make(chan []byte, 200000),
	}
	go l.loop_write()
	return l
}

func NewLogger(wc io.WriteCloser) *Logger {
	l := &Logger{
		w:           wc,
		bw:          bufio.NewWriter(wc),
		time_format: default_time_format,
		level:       LevelDebug,

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
	old_bw := l.bw
	l.w = w
	l.bw = bufio.NewWriterSize(w, 1024*1024)
	l.mutex.Unlock()

	if old_w != nil && old_bw != nil {
		old_bw.Flush()
		old_bw = nil
		old_w.Close()
		old_w = nil
		old_bw = nil
	}
}

func StringToReadonlySlice(s *string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
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
		s = fmt.Sprintf(time.Now().Format(l.time_format)+" ["+level_str+"] ["+get_caller_info().String()+"] "+format+"\n", a...)
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
				l.bw.Write(logentry)
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

var ErrEmptyLogWriter = errors.New("empty log writer")

func (l *Logger) Flush() error {
	// l.mutex.Lock()
	// defer l.mutex.Unlock()
	l.mutex.Lock()
	var err error
	if l.bw != nil {
		err = l.bw.Flush()
	} else {
		err = ErrEmptyLogWriter
	}
	l.mutex.Unlock()
	return err
}

func (l *Logger) Close() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.bw.Flush()

	if l.w != nil {
		l.w.Close()
	}
}

type caller_info struct {
	pkg  string
	file string
	fnc  string
	line int
}

func (c caller_info) String() string {
	return fmt.Sprintf("%s:%s:%s(..):%d", c.pkg, c.file, c.fnc, c.line)
}

func get_caller_info() caller_info {
	pc, file, line, _ := runtime.Caller(3)
	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	packageName := ""
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		packageName = strings.Join(parts[0:pl-2], ".")
	} else {
		packageName = strings.Join(parts[0:pl-1], ".")
	}

	return caller_info{
		pkg:  packageName,
		file: fileName,
		fnc:  funcName,
		line: line,
	}
}
