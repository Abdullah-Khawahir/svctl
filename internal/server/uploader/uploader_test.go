package uploader

import (
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
    http-headers:
     Accept: "text/json"
     Authentication: "abc:123"
`)
	c, err := InitializeArtifactConfig(passingFilepath) // err = nil
	assert.NoError(t, err, "the file contains all requirment it must not fail if it does update the requirment and update this message")
	h1, h2 := c.Handlers[0], c.Handlers[1]

	assert.Equal(t, h1.Name, "handler1")
	assert.Equal(t, h1.SourceRegex, "/path/to/source1")
	assert.Equal(t, h1.Destination, "http://path/to/destination1")

	assert.Equal(t, "handler2", h2.Name)
	assert.Equal(t, "/path/to/source2", h2.SourceRegex)
	assert.Equal(t, "http://path/to/destination1", h2.Destination)
	assert.Equal(t, map[string]string{
		"Accept":         "text/json",
		"Authentication": "abc:123",
	}, h2.Headers)

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

	assert.Equal(t, expected, got, "expected %s but got %s", expected, got)

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
	assert.Equal(t, expected, got, "expected %s but got %s", expected, got)

	defer deleteRecordFiles(handler)
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

func deleteRecordFiles(h ArtifactHandler) {
	os.Remove(h.getTrackingFailedFile())
	os.Remove(h.getTrackingSuccessfulFile())
}
