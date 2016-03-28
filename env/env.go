package env

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var ErrKeyNotFound = errors.New("env: key not found")

type Env map[string]string

func (env *Env) Get(key string) (value string) {
	var ok bool
	if value, ok = (*env)[key]; !ok {
		return ""
	}
	return value
}

func (env *Env) Exists(key string) bool {
	_, exists := (*env)[key]
	return exists
}

func (env *Env) Len() int {
	return len(*env)
}

func (env *Env) Init(src *Env) {
	(*env) = make(map[string]string)
	for k, v := range *src {
		(*env)[k] = v
	}
}

func (env *Env) GetBool(key string) (value bool) {
	s := strings.ToLower(strings.Trim(env.Get(key), " \t"))
	if s == "" || s == "0" || s == "no" || s == "false" || s == "none" {
		return false
	}
	return true
}

func (env *Env) SetBool(key string, value bool) {
	if value {
		env.Set(key, "true")
	} else {
		env.Set(key, "false")
	}
}

func (env *Env) GetString(key string) (string, error) {
	v, ok := (*env)[key]
	if !ok {
		return "", fmt.Errorf("key %s not found", key)
	}
	return v, nil
}

func (env *Env) GetInt(key string) (int, error) {
	v, err := env.GetInt64(key)
	return int(v), err
}

func (env *Env) GetInt64(key string) (int64, error) {
	if !env.Exists(key) {
		return 0, fmt.Errorf("key %s not found", key)
	}
	s := strings.Trim(env.Get(key), " \t")
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (env *Env) GetFloat32(key string) (float32, error) {
	v, err := env.GetFloat64(key)
	return float32(v), err
}

func (env *Env) GetFloat64(key string) (float64, error) {
	if !env.Exists(key) {
		return 0.0, fmt.Errorf("key %s not found", key)
	}
	s := strings.Trim(env.Get(key), " \t")
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, err
	}
	return val, nil
}

func (env *Env) SetInt(key string, value int) {
	env.Set(key, fmt.Sprintf("%d", value))
}

func (env *Env) SetInt64(key string, value int64) {
	env.Set(key, fmt.Sprintf("%d", value))
}

func (env *Env) SetFloat32(key string, value float32) {
	env.Set(key, fmt.Sprintf("%f", value))
}

func (env *Env) SetFloat64(key string, value float64) {
	env.Set(key, fmt.Sprintf("%f", value))
}

// Returns nil if key not found
func (env *Env) GetList(key string) ([]string, error) {
	if !env.Exists(key) {
		return nil, fmt.Errorf("key %s not found", key)
	}

	sval := env.Get(key)
	if sval == "" {
		return nil, fmt.Errorf("key %s not found", key)
	}
	l := make([]string, 0, 1)
	if err := json.Unmarshal([]byte(sval), &l); err != nil {
		l = append(l, sval)
	}
	return l, nil
}

func (env *Env) GetJson(key string, iface interface{}) error {
	if !env.Exists(key) {
		return fmt.Errorf("key %s not found", key)
	}
	sval := env.Get(key)
	if sval == "" {
		return nil
	}
	return json.Unmarshal([]byte(sval), iface)
}

func (env *Env) SetJson(key string, value interface{}) error {
	sval, err := json.Marshal(value)
	if err != nil {
		return err
	}
	env.Set(key, string(sval))
	return nil
}

func (env *Env) SetList(key string, value []string) error {
	return env.SetJson(key, value)
}

func (env *Env) Set(key, value string) {
	(*env)[key] = value
}

func NewDecoder(src io.Reader) *Decoder {
	return &Decoder{
		json.NewDecoder(src),
	}
}

type Decoder struct {
	*json.Decoder
}

func (decoder *Decoder) Decode() (*Env, error) {
	m := make(map[string]interface{})
	if err := decoder.Decoder.Decode(&m); err != nil {
		return nil, err
	}
	env := &Env{}
	for key, value := range m {
		env.SetAuto(key, value)
	}
	return env, nil
}

// DecodeEnv decodes `src` as a json dictionary, and adds
// each decoded key-value pair to the environment.
//
// If `src` cannot be decoded as a json dictionary, an error
// is returned.
func (env *Env) Decode(src io.Reader) error {
	m := make(map[string]interface{})
	if err := json.NewDecoder(src).Decode(&m); err != nil {
		return err
	}
	for k, v := range m {
		env.SetAuto(k, v)
	}
	return nil
}

func (env *Env) SetAuto(k string, v interface{}) {
	// FIXME: we fix-convert float values to int, because
	// encoding/json decodes integers to float64, but cannot encode them back.
	// (See http://golang.org/src/pkg/encoding/json/decode.go#L46)
	if sval, ok := v.(string); ok {
		env.Set(k, sval)
	} else if val, err := json.Marshal(v); err == nil {
		env.Set(k, string(val))
	} else {
		env.Set(k, fmt.Sprintf("%v", v))
	}
}

func changeFloats(v interface{}) interface{} {
	switch v := v.(type) {
	case float64:
		return int(v)
	case map[string]interface{}:
		for key, val := range v {
			v[key] = changeFloats(val)
		}
	case []interface{}:
		for idx, val := range v {
			v[idx] = changeFloats(val)
		}
	}
	return v
}

func (env *Env) Encode(dst io.Writer) error {
	m := make(map[string]interface{})
	for k, v := range *env {
		var val interface{}
		if err := json.Unmarshal([]byte(v), &val); err == nil {
			// FIXME: we fix-convert float values to int, because
			// encoding/json decodes integers to float64, but cannot encode them back.
			// (See http://golang.org/src/pkg/encoding/json/decode.go#L46)
			m[k] = changeFloats(val)
		} else {
			m[k] = v
		}
	}
	if err := json.NewEncoder(dst).Encode(&m); err != nil {
		return err
	}
	return nil
}

func (env *Env) WriteTo(dst io.Writer) (n int64, err error) {
	// FIXME: return the number of bytes written to respect io.WriterTo
	return 0, env.Encode(dst)
}

func (env *Env) Import(src interface{}) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("ImportEnv: %s", err)
		}
	}()
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	if err := env.Decode(&buf); err != nil {
		return err
	}
	return nil
}
