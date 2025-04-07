package test

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestMd5(t *testing.T) {
	filepath := "D:\\Downloads\\4399naikuai.exe"

	open, _ := os.Open(filepath)
	defer open.Close()

	buf := make([]byte, 1024)
	hash := md5.New()

	for {
		n, err := open.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("read error")
			return
		}

		hash.Write(buf[:n])
	}
	fmt.Println(fmt.Sprintf("%x", hash.Sum(nil)))
}
