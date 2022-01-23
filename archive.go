package archiver

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

func defaultOpts() *tarOptions {
	return &tarOptions{
		level: DefaultCompression,
	}
}

type tarOptions struct {
	honorGitIgnore bool
	ignoreDotGit   bool
	level          int
}

type TarOption func(*tarOptions)

// HonorGitIgnore will look for a .gitignore file in '.',
// parse it, and only archive files that are not matched by
// a rule in this .gitignore file.
// The current implementation does not support multiple
// .gitignore files in multiple folders.
func HonorGitIgnore() TarOption {
	return func(opts *tarOptions) {
		opts.honorGitIgnore = true
	}
}

// IgnoreDotGit does not archive any '.git' folders in any subdirectories.
func IgnoreDotGit() TarOption {
	return func(opts *tarOptions) {
		opts.ignoreDotGit = true
	}
}

// ArchiveGitRepo is a shorthand for HonorGitIgnore and IgnoreDotGit
func ArchiveGitRepo() TarOption {
	return func(opts *tarOptions) {
		opts.ignoreDotGit = true
		opts.honorGitIgnore = true
	}
}

// Tar takes a source and a writers and walks 'source' writing each file
// found to the tar writer.
// It
// * skips root
// * maintains empty folders
// * does not follow (symbolic) links
// * respects a .gitignore if it's found in the directory root
func Tar(source string, writer io.Writer, opts ...TarOption) error {
	tarOpts := defaultOpts()
	for _, opt := range opts {
		opt(tarOpts)
	}

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

	ignorer, err := newIgnorer(tarOpts, absSource)
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
		} else if ignorer != nil && ignorer.MatchesPath(relPath) {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// header name should be the path relative to the folder
		// specified to be archived
		// if the source folder is "./dir1" and dir1 contains
		// a file f1.txt, then header name should be "f1.txt"
		// and not "dir1/f1.txt"
		header.Name = relPath
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}

		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
}

// Untar takes a destination path and a reader. A tar reader loops over the tarfile
// creating the file structure at 'destination' along the way, and writing the files' contents.
// TODO: make sure that every file that tries to leave the box is skipped
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

func newIgnorer(tarOpts *tarOptions, absSource string) (*gitignore.GitIgnore, error) {
	var ignorer *gitignore.GitIgnore
	var err error
	if tarOpts.honorGitIgnore && tarOpts.ignoreDotGit {
		gitignorePath := filepath.Join(absSource, ".gitignore")
		ignorer, err = gitignore.CompileIgnoreFileAndLines(gitignorePath, "**/.git", ".gitignore")
	} else if tarOpts.honorGitIgnore && !tarOpts.ignoreDotGit {
		gitignorePath := filepath.Join(absSource, ".gitignore")
		ignorer, err = gitignore.CompileIgnoreFileAndLines(gitignorePath, ".gitignore")
	} else if !tarOpts.honorGitIgnore && tarOpts.ignoreDotGit {
		ignorer = gitignore.CompileIgnoreLines("**/.git")
	}
	return ignorer, err
}
