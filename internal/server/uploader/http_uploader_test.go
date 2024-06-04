package uploader

import (
	"fmt"
	"io"
	http "net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var test *testing.T

func TestHttpSecuredUpload(t *testing.T) {
	test = t
	files := []string{
		createFile(t, "file1.log", "log file"),
	}

	tls := httptest.NewTLSServer(http.HandlerFunc(handleLogFiles))
	defer tls.Close()

	time.Sleep(2 * time.Second)
	var config ArtifactConfig = ArtifactConfig{
		Handlers: []ArtifactHandler{
			{
				SourceRegex: "./file*.log",
				Destination: tls.URL,
				Headers: map[string]string{
					"testHeader": "testValue",
				},
			},
		},
	}

	config.populateArtifactList()

	err := config.assignUploaders()
	if err != nil {
		t.Errorf("expected to assign httpUploader but it failed : %s", err.Error())
	}

	config.UploadFiles()
	defer deleteRecordFiles(config.Handlers[0])
	defer func(files []string) {
		for _, file := range files {
			os.Remove(file)
		}
	}(files)
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

	var config ArtifactConfig = ArtifactConfig{
		Handlers: []ArtifactHandler{
			{
				SourceRegex: "./file*.log",
				Destination: "http://127.0.0.1:8080/log",
				Headers: map[string]string{
					"testHeader": "testValue",
				},
			},
			{
				SourceRegex: "./file*.txt",
				Destination: "http://127.0.0.1:8080/txt",
				Headers: map[string]string{
					"testHeader": "testValue",
				},
			},
		},
	}

	config.populateArtifactList()

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

	config.UploadFiles()
	defer deleteRecordFiles(config.Handlers[0])
	defer deleteRecordFiles(config.Handlers[1])
	defer func(files []string) {
		for _, file := range files {
			os.Remove(file)
		}
	}(files)
}

func handleTextFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "log file")
	body, _ := io.ReadAll(r.Body)
	assert.Equal(test, "testValue", r.Header.Get("testHeader"),
		"the header was not found it must be passed")
	if string(body) != "text file" {
		test.Errorf("expected text file but got %s", string(body))
	}
}

func handleLogFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "log file")
	body, _ := io.ReadAll(r.Body)
	assert.Equal(test, "testValue", r.Header.Get("testHeader"),
		"the header was not found it must be passed")
	if string(body) != "log file" {
		test.Errorf("expected log file but got %s", string(body))
	}
}
