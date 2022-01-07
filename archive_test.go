package archiver

import (
	"crypto/sha256"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUntarWithNonTarFile(t *testing.T) {
	file, err := os.Open("test-fixtures/tree2/f1.txt")
	assert.Nil(t, err)
	tempDir, err := ioutil.TempDir("", "TestUntarWithNonTarFile-dir-")
	assert.Nil(t, err)

	err = Untar(tempDir, file)
	assert.NotNil(t, err)
}

func TestTarUntar(t *testing.T) {
	tests := []struct {
		path         string
		options      []TarOption
		numEntries   int
		ignoredFiles map[string]bool
	}{
		{
			path:       "test-fixtures/tree1",
			numEntries: 3,
		},
		{
			path:       "test-fixtures/tree2",
			numEntries: 2,
		},
		{
			path:       "test-fixtures/tree3",
			numEntries: 0,
		},
		{
			path:       "test-fixtures/tree4",
			options:    []TarOption{HonorGitIgnore()},
			numEntries: 5,
			ignoredFiles: map[string]bool{
				"test-fixtures/tree4/d1/f11.txt": true,
			},
		},
	}

	for _, test := range tests {
		runTestArchiveUnArchive(t, test.path, test.options, test.numEntries, test.ignoredFiles)
	}
}

func runTestArchiveUnArchive(t *testing.T, path string, options []TarOption, numEntries int, ignoredFiles map[string]bool) {
	file, err := ioutil.TempFile("", "TestArchiveUnArchive-file-")
	assert.Nil(t, err)

	err = Tar(path, file, options...)
	assert.Nil(t, err)
	err = file.Close()
	assert.Nil(t, err)

	tempDir, err := ioutil.TempDir("", "TestArchiveUnArchive-dir-")
	assert.Nil(t, err)
	file, err = os.Open(file.Name())
	assert.Nil(t, err)
	defer file.Close()
	err = Untar(tempDir, file)
	assert.Nil(t, err)

	assertFoldersEqual(t, path, tempDir, numEntries, ignoredFiles)
}

func assertFoldersEqual(t *testing.T, dir1 string, dir2 string, numEntries int, ignoredFiles map[string]bool) {
	dir1Map, err := walkDir(dir1, ignoredFiles)
	assert.Nil(t, err)
	dir2Map, err := walkDir(dir2, map[string]bool{})
	assert.Nil(t, err)
	assert.Equal(t, len(dir1Map), len(dir2Map))
	assert.Equal(t, numEntries, len(dir1Map))

	for path, fi1 := range dir1Map {
		relPath, err := filepath.Rel(dir1, path)
		assert.Nil(t, err)

		path2 := filepath.Join(dir2, relPath)
		fi2, ok := dir2Map[path2]
		assert.True(t, ok)
		assertFileInfosEqual(t, fi1, fi2)

		if !fi1.IsDir() {
			absPath1, err := filepath.Abs(filepath.Join(dir1, relPath))
			assert.Nil(t, err)
			absPath2, err := filepath.Abs(filepath.Join(dir2, relPath))
			assert.Nil(t, err)

			hash1, err := hashFileContent(absPath1)
			assert.Nil(t, err)

			hash2, err := hashFileContent(absPath2)
			assert.Nil(t, err)

			assert.Equal(t, hash1, hash2)
		}
	}
}

func assertFileInfosEqual(t *testing.T, fi1 os.FileInfo, fi2 os.FileInfo) {
	assert.Equal(t, fi1.Mode(), fi2.Mode())
	assert.Equal(t, fi1.Name(), fi2.Name())
}

func hashFileContent(path string) (string, error) {
	hasher := sha256.New()
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer f.Close()
	_, err = io.Copy(hasher, f)
	if err != nil {
		return "", err
	}

	return string(hasher.Sum(nil)), nil
}

func walkDir(path string, ignoredFiles map[string]bool) (map[string]os.FileInfo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	m := map[string]os.FileInfo{}
	err = filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.Mode().IsRegular() && !fi.Mode().IsDir() {
			return nil
		}

		absFile, err := filepath.Abs(file)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(absPath, absFile)
		if err != nil {
			return err
		} else if relPath == "." {
			return nil
		}

		absPath := filepath.Clean(filepath.Join(path, relPath))
		m[absPath] = fi
		return nil
	})

	for key := range ignoredFiles {
		_, ok := m[key]
		if ok {
			delete(m, key)
		}
	}

	return m, err
}
