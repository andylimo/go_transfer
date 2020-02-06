package handler

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tiger5226/filetransfer/util"

	"github.com/sirupsen/logrus"
)

// Upload Handles a server request to upload content to one of the project buckets
func Upload(response http.ResponseWriter, request *http.Request) {
	hs := map[string]string{
		"Access-Control-Allow-Methods": "POST",
		"Access-Control-Allow-Origin":  "*"}

	for k, v := range hs {
		response.Header().Set(k, v)
	}

	if request.Method == "OPTIONS" {
		return
	}

	var err error
	file, fileHeader, err := request.FormFile("file")
	if err != nil {
		logrus.Error("Was not able to access the uploaded file: ", err)
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	bucket := request.FormValue("bucket")

	// Close the file afterwards:
	defer util.CloseMPFile(file)

	// Read the entire file into memory:
	data, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Error("Error while reading file from client: ", err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	currDir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}

	// Check if the directory exists!  If not, then we need to create it now
	fileName := fileHeader.Filename
	if bucket != "" {
		fileName = bucket + "/" + fileName
	}
	directory := filepath.Join(currDir, "data")
	pathElements := strings.Split(fileName, "/")
	bucket = strings.Join(pathElements[:len(pathElements)-1], "/")
	fileName = pathElements[len(pathElements)-1]
	logrus.Debug("Bucket: ", bucket)

	_, err = os.Stat(directory + "/" + bucket)
	if err != nil {
		logrus.Error(err)
	}
	if os.IsNotExist(err) {
		logrus.Debug("Server: Unable to find directory, '", directory, "'.  Creating now...")
		err := os.MkdirAll(directory+"/"+bucket, 0744) // http://permissions-calculator.org/decode/0744/
		if err != nil {
			logrus.Error(err)
		}
	}

	// Write the data into a new file on server's side:
	logrus.Debug("Directory:", directory+"/"+bucket)
	// Get the original filename:
	sourceFilename := bucket + "/" + fileName
	sourceFilePath := filepath.Join(directory, sourceFilename)
	logrus.Debug("Filename: ", fileName)
	err = ioutil.WriteFile(sourceFilePath, data, 0744)
	if err != nil {
		logrus.Error("ERROR:", err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	err = os.Chown(sourceFilePath, 65534, 65534)
	if err != nil {
		logrus.Error("ERROR:", err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.Debug("Server: File was read from client and written to disk.")
	response.WriteHeader(http.StatusOK)
}
