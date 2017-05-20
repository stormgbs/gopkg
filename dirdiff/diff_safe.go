package dirdiff

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
)

func SafeWalkDir(rootdir string, excludes []*FileMatcher) (map[string]Entry, error) {
	abs_rootdir, err := filepath.Abs(rootdir)
	if err != nil {
		return nil, err
	}

	es := make(map[string]Entry)

	err = filepath.Walk(abs_rootdir, func(epath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if epath == "." || epath == abs_rootdir {
			return nil
		}

		rel, err := filepath.Rel(abs_rootdir, epath)
		if err != nil {
			return err
		}

		for _, ex := range excludes {
			if matched, err := ex.Match(rel); err != nil {
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
			e.SymlinkTarget, err = os.Readlink(epath)
			if err != nil {
				log.Printf("[dirdiff] %s %v", epath, err)
			}
		}

		//TODO: skip device or socket file
		if !is_dir && !e.IsSymlink {
			fp, err := os.OpenFile(epath, os.O_RDONLY, 0666)
			if err != nil {
				log.Printf("[dirdiff] %s error: %v", epath, err)
				return nil
			}

			h := md5.New()
			chunks := int64(math.Floor(float64(filesize) / float64(chunk_size)))

			var i int64
			for i = 0; i < chunks; i++ {
				n, err := io.CopyN(h, fp, chunk_size)
				if err != nil {
					fp.Close()
					log.Printf("[dirdiff] %s error: %v", epath, err)
					return nil
				} else if n != chunk_size {
					log.Printf("[dirdiff] %s error: %v", epath, ErrReadLess)
					return nil
				}

			}

			left := filesize - chunks*chunk_size
			if left > 0 {
				n, err := io.CopyN(h, fp, left)
				if err != nil {
					fp.Close()
					log.Printf("[dirdiff] %s error: %v", epath, err)
					return nil
				} else if n != left {
					log.Printf("[dirdiff] %s error: %v", epath, ErrReadLess)
					return nil
				}
			}

			e.Checksum = h.Sum(nil)
		}

		es[rel] = e
		return nil
	})

	return es, err
}

func SafeDiffDirs(olddir string, newdir string, excludes []string) (map[string]EntryChange, error) {
	var matchers []*FileMatcher
	for _, s := range excludes {
		matchers = append(matchers, NewFileMatcher(s))
	}

	oldents, err := SafeWalkDir(olddir, matchers)
	if err != nil {
		return nil, err
	}

	newents, err := SafeWalkDir(newdir, matchers)
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
