package logger

import (
	"errors"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
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

var ErrEmptyLogWriter = errors.New("empty log writer")

func StringToReadonlySlice(s *string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
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
