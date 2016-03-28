package env

import (
	"testing"
)

var globalTestID string

func mkEnv(t *testing.T) *Env {
	var ev Env
	ev = make(map[string]string)
	return &ev
}
