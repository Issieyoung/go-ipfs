package util

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	files "github.com/ipfs/go-ipfs-files"
	"io"
	"io/ioutil"
)

func Zip(ctx context.Context, file files.File) (files.File, error) {
	var b bytes.Buffer
	var buf = [1024]byte{}
	gw := gzip.NewWriter(&b)
	for {
		read, err := file.Read(buf[:])
		_, err = gw.Write(buf[:read])
		if err != nil {
			break
		}
		err = gw.Flush()
		if err != nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("")
		default:

		}
		if read != 1024 {
			break
		}
	}

	return files.NewBytesFile(b.Bytes()), nil
}

func Unzip(ctx context.Context, file io.Reader) (io.Reader, error) {
	r, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	all, _ := ioutil.ReadAll(r)

	return bytes.NewReader(all), nil
}
