package archiver

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Tar takes a source and a writers and walks 'source' writing each file
// found to the tar writer. It skips root and maintains empty folders.
func Tar(source string, writer io.Writer) error {
	source = filepath.Clean(source)
	// ensure the source actually exists before trying to tar it
	sourceFi, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("unable to tar files - %s", err.Error())
	} else if !sourceFi.IsDir() {
		return fmt.Errorf("can only archive a directory")
	}

	tw := tar.NewWriter(writer)
	defer tw.Close()
	absSource, err := filepath.Abs(source)
	if err != nil {
		return err
	}

	return filepath.Walk(source, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if !fi.Mode().IsRegular() && !fi.Mode().IsDir() {
			return nil
		}

		absFile, err := filepath.Abs(file)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(absSource, absFile)
		if err != nil {
			return err
		} else if relPath == "." {
			return nil
		} else if !strings.HasPrefix(absFile, absSource) {
			return fmt.Errorf("illegal file path: [%s]", absFile)
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = relPath
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		defer f.Close()
		// copy file data into tar writer
		_, err = io.Copy(tw, f)
		return err
	})
}

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing the files' contents
func Untar(destination string, r io.Reader) error {
	// ensure the destination actually exists before trying to untar into it
	destinationFi, err := os.Stat(destination)
	if err != nil {
		return fmt.Errorf("unable to tar files - %s", err.Error())
	} else if !destinationFi.IsDir() {
		return fmt.Errorf("can only untar into a directory")
	}

	tr := tar.NewReader(r)
	absDestination, err := filepath.Abs(destination)
	if err != nil {
		return err
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		} else if header == nil {
			return nil
		}

		// the target location where the dir/file should be created
		target := filepath.Clean(filepath.Join(absDestination, header.Name))
		if !strings.HasPrefix(target, absDestination) {
			return fmt.Errorf("illegal file path: [%s]", target)
		}

		// check the file type
		if header.Typeflag == tar.TypeDir {
			_, err = os.Stat(target)
			if err != nil {
				err = os.MkdirAll(target, os.FileMode(header.Mode))
				if err != nil {
					return err
				}
			}
		} else if header.Typeflag == tar.TypeReg {
			err = writeFile(tr, target, header.Mode)
			if err != nil {
				return err
			}
		}
	}
}

func writeFile(tr *tar.Reader, target string, mode int64) error {
	f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(mode))
	if err != nil {
		return err
	}

	defer f.Close()
	// copy over contents
	_, err = io.Copy(f, tr)
	return err
}
