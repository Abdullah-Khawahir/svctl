package uploader

import (
	"fmt"
	"os"
	"strings"
)

const FailedUploadsFile = "FailedUploads.txt"
const SuccessfulUploadFile = "SuccessfulUploads.txt"

func SetFileAsSent(filepath string) {
	file, err := os.OpenFile(SuccessfulUploadFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintln(filepath))
}

func GetUploadedFiles() []string {
	fileBytes, err := os.ReadFile(SuccessfulUploadFile)
	if err != nil {
		return nil
	}

	return strings.Split(string(fileBytes), "\n")
}

func SetFileAsFailedToUpload(filepath string) {
	file, err := os.OpenFile(FailedUploadsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintln(filepath))
}
func GetFailedFiles() []string {
	fileBytes, err := os.ReadFile(FailedUploadsFile)
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimSpace(string(fileBytes)), "\n")
}
