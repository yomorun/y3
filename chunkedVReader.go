package y3

import (
	"bytes"
	"io"
	"io/ioutil"
)

type chunkVReader struct {
	src        io.Reader     // the reader parts of V
	buf        *bytes.Buffer // the bytes parts of V
	totalSize  int           // size of whole buffer of this packet
	off        int           // last read op
	ChunkVSize int           // the size of chunked V
}

// Read implement io.Reader interface
func (r *chunkVReader) Read(p []byte) (n int, err error) {
	if r.src == nil {
		return 0, nil
	}

	if r.off >= r.totalSize {
		return 0, io.EOF
	}

	if r.off < r.totalSize-r.ChunkVSize {
		n, err := r.buf.Read(p)
		r.off += n
		if err != nil {
			if err == io.EOF {
				return n, nil
			} else {
				return 0, err
			}
		}
		return n, nil
	}
	n, err = r.src.Read(p)
	r.off += n
	if err != nil {
		return n, err
	}
	return n, nil
}

// WriteTo implement io.WriteTo interface
func (r *chunkVReader) WriteTo(w io.Writer) (n int64, err error) {
	if r.src == nil {
		return 0, nil
	}

	// first, write existed buffer
	m, err := w.Write(r.buf.Bytes())
	if err != nil {
		return 0, err
	}
	n += int64(m)

	// last, write from reader
	buf, err := ioutil.ReadAll(r.src)
	if err != nil && err != io.EOF {
		return 0, errWriteFromReader
	}
	m, err = w.Write(buf)
	if err != nil {
		return 0, err
	}

	n += int64(m)
	return n, nil
}
