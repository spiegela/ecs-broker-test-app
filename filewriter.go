package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type fileWriter struct {
	bucket     string
}

func NewFileWriter(bucket string) *fileWriter {
	return &fileWriter{bucket: bucket}
}

func (w fileWriter) filePath(key string) string {
	return filepath.Join("/", bucket, key)
}

func (w fileWriter) Delete(key string) ([]byte, error) {
	err := os.Remove(w.filePath(key))
	if err != nil {
		return nil, err
	}
	return []byte("OK"), nil
}

func (w fileWriter) Read(key string) ([]byte, error) {
	return ioutil.ReadFile(w.filePath(key))
}

func (w fileWriter) Write(r *http.Request, key string) ([]byte, error) {
	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		return nil, readErr
	}
	file := w.filePath(key)
	dirName := filepath.Dir(file)
	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return nil, err
	}
	f, err := os.Create(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	bytesWritten, err := f.Write(body)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("Wrote %d bytes\n", bytesWritten)), nil
}
