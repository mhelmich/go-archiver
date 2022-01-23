package archiver

import (
	"compress/gzip"
	"io"
)

// These constants are copied from the flate package, so that code that imports
// "archiver" does not also have to import "compress/gzip".
const (
	NoCompression      = gzip.NoCompression
	BestSpeed          = gzip.BestSpeed
	BestCompression    = gzip.BestCompression
	DefaultCompression = gzip.DefaultCompression
	HuffmanOnly        = gzip.HuffmanOnly
)

// CompressionLevel allows users to the compression level of gzip
func CompressionLevel(level int) TarOption {
	return func(opts *tarOptions) {
		opts.level = level
	}
}

// GzipCompress fronts tar with a gzip compression stream.
func GzipCompress(source string, writer io.Writer, opts ...TarOption) error {
	tarOpts := defaultOpts()
	for _, opt := range opts {
		opt(tarOpts)
	}

	compressor, err := gzip.NewWriterLevel(writer, tarOpts.level)
	if err != nil {
		return err
	}

	defer compressor.Close()
	return tarWithOpts(source, compressor, tarOpts)
}

// GzipDecompress fronts tar with a gzip decompression stream.
func GzipDecompress(destination string, r io.Reader) error {
	compressor, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	return Untar(destination, compressor)
}
