package uploader

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/jlaffaye/ftp"
)

type FtpUploadStrategy struct {
	Username string
	Password string
}
type FtpSecuredUploadStrategy struct {
	Username string
	Password string
}

func (uploader FtpUploadStrategy) Upload(destination string, artifactPath string) error {
	url, parsingErr := url.Parse(destination)
	if parsingErr != nil {
		return parsingErr
	}
	connection, connectionErr := ftp.Dial(fmt.Sprintf("%s:%s", url.Hostname(), url.Port()), ftp.DialWithTLS(&tls.Config{}))
	if connectionErr != nil {
		return connectionErr
	}

	loginErr := connection.Login(uploader.Username, uploader.Password)

	if loginErr != nil {
		return loginErr
	}

	fileContent, readErr := os.ReadFile(artifactPath)
	if readErr != nil {
		return readErr
	}
	storErr := connection.Stor(filepath.Base(artifactPath), bytes.NewBuffer(fileContent))
	if storErr != nil {
		return storErr
	}
	quitErr := connection.Quit()
	if quitErr != nil {
		return quitErr
	}
	return nil
}

func (uploader FtpSecuredUploadStrategy) Upload(destination string, artifactPath string) error {
	return fmt.Errorf("unimplemented")
}
