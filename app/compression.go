package main

import (
	"bytes"
	"compress/gzip"
)

func gzipCompress(data []byte) *bytes.Buffer {
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	if _, err := gz.Write(data); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}

	// fmt.Println("Hexadecimal Representation:", hex.EncodeToString(buf.Bytes()))
	return buf
}
