package util

import (
	"bytes"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

func ToUTF8(s string) string {
	var bufs bytes.Buffer
	wr := transform.NewWriter(&bufs, korean.EUCKR.NewDecoder())
	wr.Write([]byte(s))
	wr.Close()

	return bufs.String()
}

func ToEUCKR(s string) string {
	var bufs bytes.Buffer
	wr := transform.NewWriter(&bufs, korean.EUCKR.NewEncoder())
	wr.Write([]byte(s))
	wr.Close()

	return bufs.String()
}
