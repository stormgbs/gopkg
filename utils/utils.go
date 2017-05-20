package utils

import (
	"bytes"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/stormgbs/gopkg/logger"
)

var camelRegexp = regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")

func SnakeString(source string) string {
	var splitedFileds []string

	for _, subField := range camelRegexp.FindAllStringSubmatch(source, -1) {
		if subField[1] != "" {
			splitedFileds = append(splitedFileds, subField[1])
		}

		if subField[2] != "" {
			splitedFileds = append(splitedFileds, subField[2])
		}
	}

	return strings.ToLower(strings.Join(splitedFileds, "_"))
}

func NewRedisConnPool(addr, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func MysqlEscapeString(input string) string {
	type kv struct {
		old string
		new string
	}

	var escapes = []kv{
		kv{`\`, `\\`},
		kv{`/`, `\/`},
		kv{`"`, `\"`},
		kv{`'`, `\'`},
		kv{`|`, `\|`},
		kv{`-`, `\-`},
		kv{`;`, `\;`},
		kv{`[`, `\[`},
		kv{`]`, `\]`},
	}

	for _, e := range escapes {
		input = strings.Replace(input, e.old, e.new, -1)
	}

	return input
}

func RoundFloat64s(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}

func RandString(length int) string {
	if length <= 0 {
		return ""
	}

	bys := make([]byte, length, length)

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		c := rand.Intn(26) + 97
		bys[i] = byte(c)
	}
	return string(bys)
}

func GetLogLevel(str string) logger.Level {
	// debug, info, warn, error, critial
	switch strings.ToLower(str) {
	case "debug":
		return logger.LevelDebug
	case "info", "information":
		return logger.LevelInfo
	case "warn", "warning":
		return logger.LevelWarn
	case "err", "error":
		return logger.LevelError
	case "crit", "critial":
		return logger.LevelCrit
	default:
		return logger.LevelInfo
	}
}

func RunCommandByUser(who string, command string) ([]byte, error) {
	u, err := user.Lookup(who)
	if err != nil {
		return nil, err
	}

	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	cred := &syscall.Credential{
		Uid: uint32(uid),
		Gid: uint32(gid),
	}

	cmd := exec.Command("/bin/sh", "-c", command)
	buf := bytes.NewBuffer([]byte{})
	cmd.Stderr = buf
	cmd.Stdout = buf

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: cred,
		Setpgid:    true,
	}

	err = cmd.Run()
	return buf.Bytes(), err
}

func OpenOrCreateFile(file string) (*os.File, error) {
	if err := os.MkdirAll(path.Dir(file), 0755); err != nil && !os.IsExist(err) {
		return nil, err
	}

	return os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
}
