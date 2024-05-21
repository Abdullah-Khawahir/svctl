package uploader

import (
	"bytes"
	"crypto/tls"
	"fmt"
	http "net/http"
	"os"
)

type HttpPostUploadStrategy struct {
	httpHeaders map[string]string
}
type HttpSecuredPostUploadStrategy struct {
	httpHeaders map[string]string
}

func (uploader HttpPostUploadStrategy) Upload(destination string, artifactPath string) error {
	client := http.Client{}
	fileBytes, err := os.ReadFile(artifactPath)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", destination, bytes.NewBuffer(fileBytes))
	if err != nil {
		return err
	}
	for k, v := range uploader.httpHeaders {
		request.Header.Set(k, v)
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("uploading %s returned status code %v", artifactPath, http.StatusText(response.StatusCode))
	}
	return nil
}

func (uploader HttpSecuredPostUploadStrategy) Upload(destination string, artifactPath string) error {
	fileBytes, err := os.ReadFile(artifactPath)
	if err != nil {
		return err
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	request, err := http.NewRequest("POST", destination, bytes.NewBuffer(fileBytes))
	if err != nil {
		return err
	}
	for k, v := range uploader.httpHeaders {
		request.Header.Set(k, v)
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("uploading %s returned status code %v", artifactPath, http.StatusText(response.StatusCode))
	}
	defer response.Body.Close()
	return nil
}
