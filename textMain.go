package main

import (
	"net/http"
	"bytes"
	"fmt"
	"os"
	"mime/multipart"
	"path/filepath"
	"io"
)

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func main() {
	path, _ := os.Getwd()
	path += "/cn_windows_10_multiple_editions_x64_dvd_6848463.iso"
	extraParams := map[string]string{
		"title":       "My Document",
		"author":      "Matt Aimonetti",
		"description": "A document with all the Go programming language secrets",
	}
	fmt.Println("asdfsadfsdaf")
	request, err := newfileUploadRequest("http://10.10.141.71:8000/upload", extraParams, "uploadfile", path)
	if err != nil {
		fmt.Println("asd")
		fmt.Println(err.Error())
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		//log.Fatal(err)
		fmt.Println(err.Error())
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
		//	log.Fatal(err)
			fmt.Println("错误2")
		}
		resp.Body.Close()
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
		fmt.Println(body)
	}
}