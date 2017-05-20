package dirdiff

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func Tar(rootdir string, fc map[string]EntryChange, dst_path string) (errs []error) {

	cwd, err := os.Getwd()
	if err != nil {
		errs = append(errs, err)
		return
	}

	cwd, err = filepath.Abs(cwd)
	if err != nil {
		errs = append(errs, err)
		return
	}

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

	if err := os.Chdir(rootdir); err != nil {
		errs = append(errs, err)
		return
	}

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

		fp, err := os.OpenFile(file, os.O_RDONLY, 0644)
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

	err = os.Chdir(cwd)
	if err != nil {
		errs = append(errs, err)
	}
	return
}

func getFileType(mod os.FileMode) byte {
	if mod&os.ModeDir == os.ModeDir {
		return tar.TypeDir
	}
	if mod&os.ModeSymlink == os.ModeSymlink {
		return tar.TypeSymlink
	}
	if mod&os.ModeDevice == os.ModeDevice {
		return tar.TypeBlock
	}
	if mod&os.ModeNamedPipe == os.ModeNamedPipe {
		return tar.TypeFifo
	}
	if mod&os.ModeCharDevice == os.ModeCharDevice {
		return tar.TypeChar
	}
	return tar.TypeReg
}

func TgzDir(srcdir string, target_name string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return err
	}

	if err := os.Chdir(srcdir); err != nil {
		return err
	}

	// es := make(map[string]Entry)

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

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if path == "." {
			return nil
		}

		symlink := ""
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			if symlink, err = os.Readlink(path); err != nil {
				return err
			}
		}

		hdr, err := tar.FileInfoHeader(info, symlink)
		if err != nil {
			return err
		}
		hdr.Name = path

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

		fp, err := os.OpenFile(path, os.O_RDONLY, 0644)
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
	return os.Chdir(cwd)
}

func TarDir(srcdir string, target_name string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return err
	}

	if err := os.Chdir(srcdir); err != nil {
		return err
	}

	// es := make(map[string]Entry)

	tarfile, err := os.OpenFile(target_name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tw := tar.NewWriter(tarfile)
	defer tw.Close()

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if path == "." {
			return nil
		}

		symlink := ""
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			if symlink, err = os.Readlink(path); err != nil {
				return err
			}
		}

		hdr, err := tar.FileInfoHeader(info, symlink)
		if err != nil {
			return err
		}
		hdr.Name = path

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

		fp, err := os.OpenFile(path, os.O_RDONLY, 0644)
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

	return os.Chdir(cwd)
}
