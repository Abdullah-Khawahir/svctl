package uploader

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"slices"

	"github.com/goccy/go-yaml"
)

type UploadStrategy interface {
	Upload(destination string, artifactPath string) error
}

type ArtifactHandler struct {
	ArtifactList []string
	Name         string `yaml:"name"`
	SourceRegex  string `yaml:"path"`
	Destination  string `yaml:"destination"`
	Uploader     UploadStrategy
}

type ArtifactConfig struct {
	Handlers []ArtifactHandler `yaml:"artifacts"`
}

func InitializeArtifactConfig(configPath string) (*ArtifactConfig, error) {

	configFile, readErr := os.ReadFile(configPath)
	if readErr != nil {
		return nil, readErr
	}

	var config *ArtifactConfig = &ArtifactConfig{}
	unmarshalErr := yaml.Unmarshal(configFile, config)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	err := validate(config)
	if err != nil {
		return nil, err
	}

	uploaderError := config.assignUploaders()
	if uploaderError != nil {
		return nil, uploaderError
	}

	config.PopulateArtifactList()

	config.uploadFiles()

	return config, nil
}

func validate(config *ArtifactConfig) error {
	for i := range config.Handlers {
		handler := config.Handlers[i]
		if handler.Name == "" || len(handler.Name) == 0 {
			return errors.New("each handler must have a name")
		}
		if handler.Destination == "" || len(handler.Destination) == 0 {
			return errors.New("each handler must have a destination")
		}
		if handler.SourceRegex == "" || len(handler.SourceRegex) == 0 {
			return errors.New("each handler must have a path")
		}
	}
	return nil
}

func (config *ArtifactConfig) uploadFiles() {
	for i := range config.Handlers {
		handler := config.Handlers[i]
		slices.Sort(handler.ArtifactList)
		for ii := range handler.ArtifactList {
			file := handler.ArtifactList[ii]
			uploadedFiles := GetUploadedFiles(handler)
			failedFiles := GetFailedFiles(handler)
			if slices.Contains(uploadedFiles, file) {
				continue
			}

			err := handler.UploadFile(file)
			if err != nil {
				log.Printf("Error uploading file %s: %v", file, err)

				if !slices.Contains(failedFiles, file) {
					SetFileAsFailedToUpload(handler, file)
					failedFiles = GetFailedFiles(handler)
				}
			} else {
				SetFileAsSent(handler, file)
				uploadedFiles = GetUploadedFiles(handler)
			}
		}

	}
}

// TODO: test for Glob err
func (config *ArtifactConfig) PopulateArtifactList() {
	for i := 0; i < len(config.Handlers); i++ {
		handler := config.Handlers[i]
		files, err := filepath.Glob(config.Handlers[i].SourceRegex)
		if err != nil {
			continue
		}
		config.Handlers[i].ArtifactList = make([]string, 0)
		config.Handlers[i].ArtifactList = append(handler.ArtifactList, files...)

	}
}

func (handler *ArtifactHandler) UploadFile(file string) error {
	err := handler.Uploader.Upload(handler.Destination, file)
	if err != nil {
		return err
	}
	return nil
}

func (config *ArtifactConfig) assignUploaders() error {
	for i := 0; i < len(config.Handlers); i++ {
		handler := config.Handlers[i]
		destination := handler.Destination
		URL, err := url.Parse(destination)
		if err != nil {
			return err
		}

		switch URL.Scheme {
		case "http":
			config.Handlers[i].Uploader = HttpPostUploadStrategy{}
		case "https":
			config.Handlers[i].Uploader = HttpSecuredPostUploadStrategy{}
		case "ftp":
			{
				username := URL.User.Username()
				pass, _ := URL.User.Password()

				config.Handlers[i].Uploader = FtpUploadStrategy{
					Username: username,
					Password: pass,
				}
			}

		case "ftps":
			fallthrough

		default:
			return fmt.Errorf("%v is not supported", destination)
		}

	}
	return nil
}
