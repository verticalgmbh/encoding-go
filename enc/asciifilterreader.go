package enc

import (
	"bufio"
	"io"
)

// ASCIIFilterReader stream which filters out any bytes which are no valid ASCII characters
type ASCIIFilterReader struct {
	// reader used to read data
	reader *bufio.Reader
}

// NewASCIIFilterReader creates a new utf8 filter reader
//
// **Parameters**
//   - inputreader: source stream to filter
//
// **Returns**
//   - *Utf8FilterReader: created reader
func NewASCIIFilterReader(inputreader io.Reader) *ASCIIFilterReader {
	return &ASCIIFilterReader{
		reader: bufio.NewReader(inputreader)}

}

// Read reads data from the original reader
//
// **Parameters**
//   - p: buffer to fill
//
// **Returns**
//   - int: number of bytes read
//   - error: error if any occured
func (stream *ASCIIFilterReader) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		data, err := stream.ReadByte()
		if err != nil {
			return i, err
		}

		p[i] = data
	}

	return len(p), nil
}

// ReadByte reads the next byte from the original reader and filters out any non ascii bytes
//
// **Returns**
//   - byte: byte read from stream
//   - error: error if any occured
func (stream *ASCIIFilterReader) ReadByte() (byte, error) {
	for {
		data, err := stream.reader.ReadByte()

		if err != nil {
			return data, err
		}

		if data > byte(127) {
			continue
		}

		return data, nil
	}
}
