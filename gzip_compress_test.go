package archiver

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressDecompress(t *testing.T) {
	file, err := ioutil.TempFile("", "TestCompressDecompress-file-")
	assert.Nil(t, err)

	err = GzipCompress("test-fixtures/tree1", file)
	assert.Nil(t, err)
	err = file.Close()
	assert.Nil(t, err)

	tempDir, err := ioutil.TempDir("", "TestCompressDecompress-dir-")
	assert.Nil(t, err)
	file, err = os.Open(file.Name())
	assert.Nil(t, err)
	defer file.Close()

	err = GzipDecompress(tempDir, file)
	assert.Nil(t, err)

	assertFoldersEqual(t, "test-fixtures/tree1", tempDir, 3, map[string]bool{})
}
