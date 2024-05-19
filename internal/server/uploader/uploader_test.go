package uploader

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var test *testing.T

func TestArtifactConfigValidation(t *testing.T) {
	// WARNING: this test will fail if the formatter formattes the strings in
	//	    the createFile function so be aware
	passingFilepath := createFile(t, "good.yaml", `
artifacts:
  - name: handler1
    path: /path/to/source1
    destination: http://path/to/destination1
  - name: handler2
    path: /path/to/source2
    destination: http://path/to/destination1
`)
	c, err := InitializeArtifactConfig(passingFilepath) // err = nil
	assert.NoError(t, err, "the file contains all requirment it must not fail if it does update the requirment and update this message")
	h1, h2 := c.Handlers[0], c.Handlers[1]
	assert.Equal(t, h1.Name, "handler1")
	assert.Equal(t, h2.Name, "handler2")

	assert.Equal(t, h1.SourceRegex, "/path/to/source1")
	assert.Equal(t, h2.SourceRegex, "/path/to/source2")

	assert.Equal(t, h1.Destination, "http://path/to/destination1")
	assert.Equal(t, h2.Destination, "http://path/to/destination1")

	failingFilePath := createFile(t, "bad.yaml", `
artifacts:
	- path: /path/to/source1
	destination: /path/to/destination1
	- name: handler2
	destination: http://path/to/destination1
	- name: handler3
	path: /path/to/source2
		`)

	_, err1 := InitializeArtifactConfig(failingFilePath) // err = somthing
	assert.EqualError(t, err1, "each handler must have a name")

	defer os.Remove(failingFilePath)
	defer os.Remove(passingFilepath)
}
func TestFailedFiles(t *testing.T) {
	handler := ArtifactHandler{
		Name:         "test",
		ArtifactList: []string{"f1.txt", "f2.txt"},
	}
	SetFileAsFailedToUpload(handler, "f1.txt")
	SetFileAsFailedToUpload(handler, "f2.txt")
	expected := []string{"f1.txt", "f2.txt"}
	got := GetFailedFiles(handler)
	if !slices.Equal(expected, got) {
		t.Errorf("expected %s but got %s", expected, got)
	}
	defer deleteRecordFiles(handler)
}
func TestSentFiles(t *testing.T) {
	os.Remove(SuccessfulUploadFile)
	handler := ArtifactHandler{
		Name:         "test",
		ArtifactList: []string{"f1.txt", "f2.txt"},
	}
	SetFileAsFailedToUpload(handler, "f1.txt")
	SetFileAsFailedToUpload(handler, "f2.txt")
	expected := []string{"f1.txt", "f2.txt"}
	got := GetFailedFiles(handler)
	if !slices.Equal(expected, got) {
		t.Errorf("expected %s but got %s", expected, got)
	}
	defer deleteRecordFiles(handler)
}
func TestSentFilesEmpty(t *testing.T) {
	os.Remove(SuccessfulUploadFile)
	handler := ArtifactHandler{
		Name:         "test",
		ArtifactList: []string{"f1.txt", "f2.txt"},
	}
	var expected []string = nil
	got := GetFailedFiles(handler)
	if !slices.Equal(expected, got) {
		t.Errorf("expected %s but got %s", expected, got)
	}
	defer deleteRecordFiles(handler)
}

func TestHttpSecuredUpload(t *testing.T) {
	test = t
	files := []string{
		createFile(t, "file1.log", "log file"),
	}
	defer func(files []string) {
		for _, file := range files {
			os.Remove(file)
		}
	}(files)

	tls := httptest.NewTLSServer(http.HandlerFunc(handleLogFiles))
	defer tls.Close()

	time.Sleep(2 * time.Second)
	var config ArtifactConfig = ArtifactConfig{
		Handlers: []ArtifactHandler{
			{
				SourceRegex: "./file*.log",
				Destination: tls.URL,
			},
		},
	}

	config.PopulateArtifactList()

	err := config.assignUploaders()
	if err != nil {
		t.Errorf("expected to assign httpUploader but it failed : %s", err.Error())
	}

	config.uploadFiles()
	defer deleteRecordFiles(config.Handlers[0])
}
func TestHttpUpload(t *testing.T) {
	test = t
	http.HandleFunc("/txt", handleTextFiles)
	http.HandleFunc("/log", handleLogFiles)
	go http.ListenAndServe(":8080", nil)

	time.Sleep(2 * time.Second)

	files := []string{
		createFile(t, "file1.log", "log file"),
		createFile(t, "file2.txt", "text file"),
	}

	defer func(files []string) {
		for _, file := range files {
			os.Remove(file)
		}
	}(files)
	var config ArtifactConfig = ArtifactConfig{
		Handlers: []ArtifactHandler{
			{
				SourceRegex: "./file*.log",
				Destination: "http://127.0.0.1:8080/log",
			},
			{
				SourceRegex: "./file*.txt",
				Destination: "http://127.0.0.1:8080/txt",
			},
		},
	}

	config.PopulateArtifactList()

	if !strings.EqualFold(config.Handlers[0].ArtifactList[0], files[0]) {
		t.Errorf("expected %s but found %s", files[0], config.Handlers[0].ArtifactList[0])
	}
	if config.Handlers[1].ArtifactList[0] != files[1] {
		t.Errorf("expected %s but found %s", files[1], config.Handlers[1].ArtifactList[0])
	}

	err := config.assignUploaders()
	if err != nil {
		t.Errorf("expected to assign httpUploader but it failed : %s", err.Error())
	}

	config.uploadFiles()
	defer deleteRecordFiles(config.Handlers[0])
	defer deleteRecordFiles(config.Handlers[1])
}

func createFile(t *testing.T, filename string, fileContent string) string {
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
		return ""
	}
	defer file.Close()
	file.Write([]byte(fileContent))

	return filename
}

func handleTextFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "log file")
	body, _ := io.ReadAll(r.Body)
	if string(body) != "text file" {
		test.Errorf("expected text file but got %s", string(body))
	}
}

func handleLogFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "log file")
	body, _ := io.ReadAll(r.Body)

	if string(body) != "log file" {
		test.Errorf("expected log file but got %s", string(body))
	}
}
func deleteRecordFiles(h ArtifactHandler) {
	os.Remove(h.getTrackingFailedFile())
	os.Remove(h.getTrackingSuccessfulFile())
}
