package uploader

import (
	"fmt"
	"os"
	"strings"
)

const FailedUploadsFile = "%s-failed.txt"
const SuccessfulUploadFile = "%s-uploaded.txt"

func (h ArtifactHandler) getTrackingFailedFile() string {
	return fmt.Sprintf(FailedUploadsFile, h.Name)
}

func (h ArtifactHandler) getTrackingSuccessfulFile() string {
	return fmt.Sprintf(SuccessfulUploadFile, h.Name)
}

func SetFileAsSent(handler ArtifactHandler, pathToSet string) {
	file, err := os.OpenFile(handler.getTrackingFailedFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintln(pathToSet))
}

func GetUploadedFiles(handler ArtifactHandler) []string {
	fileBytes, err := os.ReadFile(handler.getTrackingSuccessfulFile())
	if err != nil {
		return nil
	}

	return strings.Split(string(fileBytes), "\n")
}

func SetFileAsFailedToUpload(handler ArtifactHandler, pathToSet string) {
	file, err := os.OpenFile(handler.getTrackingSuccessfulFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintln(pathToSet))
}
func GetFailedFiles(handler ArtifactHandler) []string {
	fileBytes, err := os.ReadFile(handler.getTrackingSuccessfulFile())
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimSpace(string(fileBytes)), "\n")
}
