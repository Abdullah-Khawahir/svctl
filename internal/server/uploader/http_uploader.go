package uploader

import (
	"bytes"
	"crypto/tls"
	"fmt"
	http "net/http"
	"os"
)

type HttpPostUploadStrategy struct{}
type HttpSecuredPostUploadStrategy struct{}

func (uploader HttpPostUploadStrategy) Upload(destination string, artifactPath string) error {

	fileBytes, err := os.ReadFile(artifactPath)
	if err != nil {
		return err
	}
	response, postErr := http.Post(destination, "text/plain", bytes.NewBuffer(fileBytes))
	if postErr != nil {
		return postErr
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
	response, err := client.Post(destination, "text/plain", bytes.NewBuffer(fileBytes))
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("uploading %s returned status code %v", artifactPath, http.StatusText(response.StatusCode))
	}
	defer response.Body.Close()
	return nil
}
