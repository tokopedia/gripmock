package tool

import "io"

type MultiWriter struct {
	writers []io.Writer
}

func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

func (w *MultiWriter) Write(p []byte) (n int, err error) {
	for _, v := range w.writers {
		n, err = v.Write(p)
		if err != nil {
			return
		}
	}
	return
}
