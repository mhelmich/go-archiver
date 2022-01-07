package archiver

import (
	"compress/gzip"
	"io"
)

// GzipCompress fronts tar with a gzip compression stream.
func GzipCompress(source string, writer io.Writer) error {
	compressor := gzip.NewWriter(writer)
	defer compressor.Close()
	return Tar(source, compressor)
}

// GzipDecompress fronts tar with a gzip compression stream.
func GzipDecompress(destination string, r io.Reader) error {
	compressor, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	return Untar(destination, compressor)
}
