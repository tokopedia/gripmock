package tool

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func UploadJsonFile(addr string, filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return 0, err
	}

	return Upload(addr, "application/json", b)
}

func ZipFolderAndUpload(addr string, folder string) (int, error) {
	b, err := ZipFolder(folder)
	if err != nil {
		return 0, err
	}

	return Upload(addr, "binary/octet-stream", b)
}

func ZipFolder(folder string) ([]byte, error) {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	dir, err := filepath.Abs(folder)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// ignore top level folder
		if path == dir {
			return nil
		}
		// remove top level folder from path
		relPath := path[len(dir)+1:]

		if info.IsDir() {
			relPath += "/"
		}

		f, err := w.Create(relPath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(b))
		return err
	})
	if err != nil {
		return nil, err
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Upload(addr string, mimeType string, payload []byte) (int, error) {
	res, err := http.Post(addr, mimeType, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	message, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, errors.New(string(message))
	}

	return res.StatusCode, nil
}

func UploadAsJson(addr string, payload interface{}) (int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	return Upload(addr, "application/json", data)
}
