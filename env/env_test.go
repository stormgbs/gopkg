package env

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/docker/docker/pkg/testutils"
)

func TestEnvLenZero(t *testing.T) {
	env := &Env{}
	if env.Len() != 0 {
		t.Fatalf("%d", env.Len())
	}
}

func TestEnvLenNotZero(t *testing.T) {
	env := &Env{}
	env.Set("foo", "bar")
	env.Set("ga", "bu")
	if env.Len() != 2 {
		t.Fatalf("%d", env.Len())
	}
}

func TestEnvLenDup(t *testing.T) {
	env := &Env{
		"foo": "bar",
		"a":   "b",
	}
	if env.Len() != 2 {
		t.Fatalf("%d", env.Len())
	}
}

func TestSet(t *testing.T) {
	ev := mkEnv(t)
	ev.Set("foo", "bar")
	if val := ev.Get("foo"); val != "bar" {
		t.Fatalf("Get returns incorrect value: %s", val)
	}

	ev.Set("bar", "")
	if val := ev.Get("bar"); val != "" {
		t.Fatalf("Get returns incorrect value: %s", val)
	}
	if val := ev.Get("nonexistent"); val != "" {
		t.Fatalf("Get returns incorrect value: %s", val)
	}
}

func TestSetBool(t *testing.T) {
	ev := mkEnv(t)
	ev.SetBool("foo", true)
	if val := ev.GetBool("foo"); !val {
		t.Fatalf("GetBool returns incorrect value: %t", val)
	}

	ev.SetBool("bar", false)
	if val := ev.GetBool("bar"); val {
		t.Fatalf("GetBool returns incorrect value: %t", val)
	}

	if val := ev.GetBool("nonexistent"); val {
		t.Fatalf("GetBool returns incorrect value: %t", val)
	}
}

func TestSetInt(t *testing.T) {
	ev := mkEnv(t)

	ev.SetInt("foo", -42)
	if val, err := ev.GetInt("foo"); err != nil || val != -42 {
		t.Fatalf("GetInt returns incorrect value: %d", val)
	}

	ev.SetInt("bar", 42)
	if val, err := ev.GetInt("bar"); err != nil || val != 42 {
		t.Fatalf("GetInt returns incorrect value: %d", val)
	}
	if val, err := ev.GetInt("nonexistent"); err != ErrKeyNotFound {
		t.Fatalf("GetInt returns incorrect value: %d", val)
	}
}

func TestSetList(t *testing.T) {
	ev := mkEnv(t)

	ev.SetList("foo", []string{"bar"})
	if val, err := ev.GetList("foo"); err != nil || len(val) != 1 || val[0] != "bar" {
		t.Fatalf("GetList returns incorrect value: %v", val)
	}

	ev.SetList("bar", nil)
	if val, err := ev.GetList("bar"); err != nil || val != nil {
		t.Fatalf("GetList returns incorrect value: %v", val)
	}
	if val, err := ev.GetList("nonexistent"); err != ErrKeyNotFound || val != nil {
		t.Fatalf("GetList returns incorrect value: %v", val)
	}
}

func TestEnviron(t *testing.T) {
	ev := mkEnv(t)
	ev.Set("foo", "bar")
	if !ev.Exists("foo") {
		t.Fatalf("foo not found in the environ")
	}
	if val, err := ev.GetString("foo"); err != nil || val != "bar" {
		t.Fatalf("bar not found in the environ")
	}
}

// func TestMultiMap(t *testing.T) {
// 	e := &Env{}
// 	e.Set("foo", "bar")
// 	e.Set("bar", "baz")
// 	e.Set("hello", "world")
// 	m := e.MultiMap()
// 	e2 := &Env{}
// 	e2.Set("old_key", "something something something")
// 	e2.InitMultiMap(m)
// 	if v := e2.Get("old_key"); v != "" {
// 		t.Fatalf("%#v", v)
// 	}
// 	if v := e2.Get("bar"); v != "baz" {
// 		t.Fatalf("%#v", v)
// 	}
// 	if v := e2.Get("hello"); v != "world" {
// 		t.Fatalf("%#v", v)
// 	}
// }

func testMap(l int) [][2]string {
	res := make([][2]string, l)
	for i := 0; i < l; i++ {
		t := [2]string{testutils.RandomString(5), testutils.RandomString(20)}
		res[i] = t
	}
	return res
}

func BenchmarkSet(b *testing.B) {
	fix := testMap(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env := &Env{}
		for _, kv := range fix {
			env.Set(kv[0], kv[1])
		}
	}
}

func BenchmarkSetJson(b *testing.B) {
	fix := testMap(100)
	type X struct {
		f string
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env := &Env{}
		for _, kv := range fix {
			if err := env.SetJson(kv[0], X{kv[1]}); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkGet(b *testing.B) {
	fix := testMap(100)
	env := &Env{}
	for _, kv := range fix {
		env.Set(kv[0], kv[1])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, kv := range fix {
			env.Get(kv[0])
		}
	}
}

func BenchmarkGetJson(b *testing.B) {
	fix := testMap(100)
	env := &Env{}
	type X struct {
		f string
	}
	for _, kv := range fix {
		env.SetJson(kv[0], X{kv[1]})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, kv := range fix {
			if err := env.GetJson(kv[0], &X{}); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	fix := testMap(100)
	env := &Env{}
	type X struct {
		f string
	}
	// half a json
	for i, kv := range fix {
		if i%2 != 0 {
			if err := env.SetJson(kv[0], X{kv[1]}); err != nil {
				b.Fatal(err)
			}
			continue
		}
		env.Set(kv[0], kv[1])
	}
	var writer bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Encode(&writer)
		writer.Reset()
	}
}

func BenchmarkDecode(b *testing.B) {
	fix := testMap(100)
	env := &Env{}
	type X struct {
		f string
	}
	// half a json
	for i, kv := range fix {
		if i%2 != 0 {
			if err := env.SetJson(kv[0], X{kv[1]}); err != nil {
				b.Fatal(err)
			}
			continue
		}
		env.Set(kv[0], kv[1])
	}
	var writer bytes.Buffer
	env.Encode(&writer)
	denv := &Env{}
	reader := bytes.NewReader(writer.Bytes())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := denv.Decode(reader)
		if err != nil {
			b.Fatal(err)
		}
		reader.Seek(0, 0)
	}
}

func TestLongNumbers(t *testing.T) {
	type T struct {
		TestNum int64
	}
	v := T{67108864}
	var buf bytes.Buffer
	e := &Env{}
	e.SetJson("Test", v)
	if err := e.Encode(&buf); err != nil {
		t.Fatal(err)
	}
	res := make(map[string]T)
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res["Test"].TestNum != v.TestNum {
		t.Fatalf("TestNum %d, expected %d", res["Test"].TestNum, v.TestNum)
	}
}

func TestLongNumbersArray(t *testing.T) {
	type T struct {
		TestNum []int64
	}
	v := T{[]int64{67108864}}
	var buf bytes.Buffer
	e := &Env{}
	e.SetJson("Test", v)
	if err := e.Encode(&buf); err != nil {
		t.Fatal(err)
	}
	res := make(map[string]T)
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res["Test"].TestNum[0] != v.TestNum[0] {
		t.Fatalf("TestNum %d, expected %d", res["Test"].TestNum, v.TestNum)
	}
}
