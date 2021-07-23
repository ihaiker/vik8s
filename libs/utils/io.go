package utils

import (
	"io"
	"io/ioutil"
	"os"
)

func NotExists(path string) bool {
	_, err := os.Stat(path)
	return err != nil && os.IsNotExist(err)
}

func Exists(path string) bool {
	return !NotExists(path)
}

func FileBytes(path string) []byte {
	bs, err := ioutil.ReadFile(path)
	Panic(err, "read file %s", path)
	return bs
}

func Mkdir(dest string) error {
	return os.MkdirAll(dest, 0755)
}

func Copy(src, dest string) (err error) {
	var input io.ReadCloser
	var output io.WriteCloser

	if NotExists(src) {
		return Error("file not found: %s", src)
	}

	if output, err = os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0666); err != nil {
		return Error("open file %s", dest)
	}

	if input, err = os.Open(src); err != nil {
		return Error("open file %s", src)
	}

	if _, err = io.Copy(output, input); err != nil {
		return Error("copy file, from %s to %s", src, dest)
	}
	return
}
