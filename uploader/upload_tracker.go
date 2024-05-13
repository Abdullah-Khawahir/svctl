package uploader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const FailedUploadsFile = "FailedUploads.txt"
const SuccessfulUploadFile = "SuccessfulUploads.txt"

func  SetFileAsSent(path string) {
	abs ,_ := filepath.Abs(path)
	dir , _ := filepath.Split(abs)	
	
	file, err := os.OpenFile(dir + SuccessfulUploadFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintln(path))
}

func GetUploadedFiles(path string) []string {
	abs ,_ := filepath.Abs(path)
	dir , _ := filepath.Split(abs)
	fileBytes, err := os.ReadFile(dir + SuccessfulUploadFile)
	if err != nil {
		return nil
	}

	return strings.Split(string(fileBytes), "\n")
}

func SetFileAsFailedToUpload(path string) {
	abs ,_ := filepath.Abs(path)
	dir , _ := filepath.Split(abs)	
	
	file, err := os.OpenFile(dir + FailedUploadsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintln(path))
}
func GetFailedFiles(path string ) []string {
	abs ,_ := filepath.Abs(path)
	dir , _ := filepath.Split(abs)
	fileBytes, err := os.ReadFile(dir + FailedUploadsFile)
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimSpace(string(fileBytes)), "\n")
}
