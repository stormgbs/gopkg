package dirdiff

import (
	"path/filepath"
	"strings"
)

type FileMatcher struct {
	patt string

	patt_fields []string
}

func NewFileMatcher(patt string) *FileMatcher {
	var p = filepath.Clean(patt)

	if rel, err := filepath.Rel("/", p); err == nil {
		p = rel
	}

	return &FileMatcher{
		patt:        p,
		patt_fields: strings.Split(p, "/"),
	}
}

func (m *FileMatcher) Match(fpath string) (bool, error) {
	p := filepath.Clean(fpath)
	p_fields := strings.Split(p, "/")

	p_fields_len := len(p_fields)
	patt_fields_len := len(m.patt_fields)

	if p_fields_len > patt_fields_len {
		p = strings.Join(p_fields[:patt_fields_len], "/")
	}

	return filepath.Match(m.patt, p)
}
