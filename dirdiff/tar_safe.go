package dirdiff

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

func SafeTar(rootdir string, fc map[string]EntryChange, dst_path string) (errs []error) {
	dst, err := filepath.Abs(dst_path)
	if err != nil {
		errs = append(errs, err)
		return
	}

	md5file, err := os.OpenFile(dst+".md5sum", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		errs = append(errs, err)
		return
	}
	defer md5file.Close()

	tarfile, err := os.OpenFile(dst+".tar", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		errs = append(errs, err)
		return
	}

	err = tarfile.Truncate(0)
	if err != nil {
		errs = append(errs, err)
		return
	}
	defer tarfile.Close()

	tw := tar.NewWriter(tarfile)
	for file, e := range fc {
		if e.Type == ChangeTypeDelete {
			continue
		}

		if e.Type != ChangeTypeAdd && e.Type != ChangeTypeModify {
			continue
		}

		hdr, err := tar.FileInfoHeader(e.Entry.Fileinfo, e.Entry.SymlinkTarget)
		if err != nil {
			log.Printf("[dirdiff] Tar %s error: %v", file, err)
			errs = append(errs, err)
			continue
		}
		hdr.Name = file

		abs_file := path.Join(rootdir, file)

		err = tw.WriteHeader(hdr)
		if err != nil {
			log.Printf("[dirdiff] Tar %s error: %v", file, err)
			errs = append(errs, err)
			continue
		}

		if hdr.Typeflag != tar.TypeDir && !e.Entry.IsSymlink {
			md5file.WriteString(fmt.Sprintf("%x  %s\n", e.Entry.Checksum, file))
		}

		if hdr.Typeflag == tar.TypeDir || hdr.Typeflag == tar.TypeSymlink {
			continue
		}

		fp, err := os.OpenFile(abs_file, os.O_RDONLY, 0644)
		if err != nil {
			log.Printf("[dirdiff] Tar %s error: %v", file, err)
			errs = append(errs, err)
			continue
		}

		n, err := io.CopyN(tw, fp, e.Entry.Fileinfo.Size())
		if n != e.Entry.Fileinfo.Size() {
			fp.Close()
			log.Printf("[dirdiff] Tar %s error: %v", file, ErrReadLess)
			errs = append(errs, err)
			continue
		}
		if err != nil {
			fp.Close()
			log.Printf("[dirdiff] Tar %s error: %v", file, err)
			errs = append(errs, err)
			continue
		}

		fp.Close()

	}

	tw.Close()

	return
}

func SafeTgzDir(srcdir string, target_name string) error {
	abs_srcdir, err := filepath.Abs(srcdir)
	if err != nil {
		return err
	}

	tgzfile, err := os.OpenFile(target_name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer tgzfile.Close()

	gzfp, err := gzip.NewWriterLevel(tgzfile, gzip.DefaultCompression)
	if err != nil {
		return err
	}
	defer gzfp.Close()

	tw := tar.NewWriter(gzfp)

	err = filepath.Walk(abs_srcdir, func(epath string, info os.FileInfo, err error) error {
		if epath == "." || epath == abs_srcdir {
			return nil
		}

		symlink := ""
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			if symlink, err = os.Readlink(epath); err != nil {
				return err
			}
		}

		hdr, err := tar.FileInfoHeader(info, symlink)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(abs_srcdir, epath)
		if err != nil {
			return err
		}
		hdr.Name = rel

		if err = tw.WriteHeader(hdr); err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir, tar.TypeSymlink:
			return nil
		case tar.TypeReg, tar.TypeRegA:
		default:
			return fmt.Errorf("Unknown tar file type: %c", hdr.Typeflag)
		}

		fp, err := os.OpenFile(epath, os.O_RDONLY, 0644)
		if err != nil {
			return err
		}

		n, err := io.CopyN(tw, fp, info.Size())
		if err != nil {
			fp.Close()
			return err
		}

		if n != info.Size() {
			fp.Close()
			return ErrReadLess
		}

		fp.Close()
		return nil
	})

	tw.Close()
	return nil
}

func SafeTarDir(srcdir string, target_name string) error {
	abs_srcdir, err := filepath.Abs(srcdir)
	if err != nil {
		return err
	}

	tarfile, err := os.OpenFile(target_name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tw := tar.NewWriter(tarfile)
	defer tw.Close()

	err = filepath.Walk(abs_srcdir, func(epath string, info os.FileInfo, err error) error {
		if epath == "." || epath == abs_srcdir {
			return nil
		}

		symlink := ""
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			if symlink, err = os.Readlink(epath); err != nil {
				return err
			}
		}

		hdr, err := tar.FileInfoHeader(info, symlink)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(abs_srcdir, epath)
		if err != nil {
			return err
		}

		hdr.Name = rel

		if err = tw.WriteHeader(hdr); err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir, tar.TypeSymlink:
			return nil
		case tar.TypeReg, tar.TypeRegA:
		default:
			return fmt.Errorf("Unknown tar file type: %c", hdr.Typeflag)
		}

		fp, err := os.OpenFile(epath, os.O_RDONLY, 0644)
		if err != nil {
			return err
		}

		n, err := io.CopyN(tw, fp, info.Size())
		if err != nil {
			fp.Close()
			return err
		}

		if n != info.Size() {
			fp.Close()
			return ErrReadLess
		}

		fp.Close()
		return nil
	})

	return nil
}
