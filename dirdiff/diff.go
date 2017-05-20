package dirdiff

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
)

var ErrNotDir = errors.New("not dirrectory")

type Entry struct {
	IsSymlink     bool        `json:"is_symlink"`
	SymlinkTarget string      `json:"symlink_target"` //only for symbolic file
	Checksum      []byte      `json:"checksum"`
	Fileinfo      os.FileInfo `json:"fileinfo"`
}

const chunk_size int64 = 10 * 1024 * 1024

var ErrReadLess = errors.New("no enough data read")

func WalkDir(rootdir string, excludes []*FileMatcher) (map[string]Entry, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return nil, err
	}

	if err := os.Chdir(rootdir); err != nil {
		return nil, err
	}

	es := make(map[string]Entry)

	err = filepath.Walk(".", func(pth string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if pth == "." {
			return nil
		}

		for _, matcher := range excludes {
			if matched, err := matcher.Match(pth); err != nil {
				return err
			} else if matched {
				return nil
			}
		}

		var chksum []byte

		is_dir := info.IsDir()
		filesize := info.Size()

		e := Entry{
			Checksum: chksum,
			Fileinfo: info,
		}

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			e.IsSymlink = true
			e.SymlinkTarget, err = os.Readlink(pth)
			if err != nil {
				log.Printf("[dirdiff] %s %v", pth, err)
			}
		}

		//TODO: skip device or socket file
		if !is_dir && !e.IsSymlink {
			fp, err := os.OpenFile(pth, os.O_RDONLY, 0666)
			if err != nil {
				log.Printf("[dirdiff] %s error: %v", pth, err)
				return nil
			}

			h := md5.New()
			chunks := int64(math.Floor(float64(filesize) / float64(chunk_size)))

			var i int64
			for i = 0; i < chunks; i++ {
				n, err := io.CopyN(h, fp, chunk_size)
				if err != nil {
					fp.Close()
					log.Printf("[dirdiff] %s error: %v", pth, err)
					return nil
				} else if n != chunk_size {
					log.Printf("[dirdiff] %s error: %v", pth, ErrReadLess)
					return nil
				}

			}

			left := filesize - chunks*chunk_size
			if left > 0 {
				n, err := io.CopyN(h, fp, left)
				if err != nil {
					fp.Close()
					log.Printf("[dirdiff] %s error: %v", pth, err)
					return nil
				} else if n != left {
					log.Printf("[dirdiff] %s error: %v", pth, ErrReadLess)
					return nil
				}
			}

			e.Checksum = h.Sum(nil)
		}

		es[pth] = e
		return nil
	})

	err = os.Chdir(cwd)
	return es, err
}

type ChangeType uint8

const (
	ChangeTypeMode ChangeType = 1 << iota
	ChangeTypeModify
	ChangeTypeAdd
	ChangeTypeDelete
)

func (ct ChangeType) String() string {
	switch ct {
	case ChangeTypeMode:
		return "filemode"
	case ChangeTypeModify:
		return "modify"
	case ChangeTypeAdd:
		return "add"
	case ChangeTypeDelete:
		return "delete"
	}
	return ""
}

type EntryChange struct {
	Type  ChangeType `json:"type"`
	Entry Entry      `json:"entry"`
}

type FileChange struct {
	File  string      `json:"f"`
	Type  string      `json:"t"`
	Value interface{} `json:"v"`
}

func DiffDirs(olddir string, newdir string, excludes []string) (map[string]EntryChange, error) {
	var matchers []*FileMatcher
	for _, s := range excludes {
		matchers = append(matchers, NewFileMatcher(s))
	}

	oldents, err := WalkDir(olddir, matchers)
	if err != nil {
		return nil, err
	}

	newents, err := WalkDir(newdir, matchers)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]EntryChange)

	for pth, olde := range oldents {
		if newe, ok := newents[pth]; ok {
			if olde.Fileinfo.Size() != newe.Fileinfo.Size() || !bytes.Equal(olde.Checksum, newe.Checksum) {
				fmt.Printf("%s %x %x\n", pth, newe.Checksum, olde.Checksum)
				ret[pth] = EntryChange{
					Type:  ChangeTypeModify,
					Entry: newe,
				}
			} else if olde.Fileinfo.Mode() != newe.Fileinfo.Mode() {
				ret[pth] = EntryChange{
					Type:  ChangeTypeMode,
					Entry: newe,
				}
			}
		} else {
			ret[pth] = EntryChange{
				Type:  ChangeTypeDelete,
				Entry: olde,
			}
		}
	}

	for pth, newe := range newents {
		if _, ok := oldents[pth]; !ok {
			ret[pth] = EntryChange{
				Type:  ChangeTypeAdd,
				Entry: newe,
			}
		}
	}

	return ret, nil
}
