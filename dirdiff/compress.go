package dirdiff

import (
	"compress/gzip"
	"io"
	"math"
	"os"
)

func Gzip(tar string, dst string) error {
	if dst == "" {
		dst = tar + ".gz"
	}

	dst_fp, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dst_fp.Close()

	dst_fp.Truncate(0)

	gzfp, err := gzip.NewWriterLevel(dst_fp, gzip.DefaultCompression)
	if err != nil {
		return err
	}
	defer gzfp.Close()

	tar_fp, err := os.OpenFile(tar, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer tar_fp.Close()

	tar_finfo, err := tar_fp.Stat()
	if err != nil {
		return err
	}

	chunks := int64(math.Floor(float64(tar_finfo.Size()) / float64(chunk_size)))

	var i, n int64
	left := tar_finfo.Size()

	for i = 0; i <= chunks; i++ {
		if left < chunk_size {
			n, err = io.CopyN(gzfp, tar_fp, left)
		} else {
			n, err = io.CopyN(gzfp, tar_fp, chunk_size)
		}

		if err != nil {
			if err == io.EOF && n == 0 {
				err = nil
			} else {
				break
			}
		}
		left -= n
	}

	return err
}
