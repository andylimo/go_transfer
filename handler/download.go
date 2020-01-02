package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/tiger5226/filetransfer/util"

	"github.com/lbryio/lbry.go/extras/errors"
	"github.com/sirupsen/logrus"
)

// Download Handles a server request to download content from one of the project buckets
func Download(response http.ResponseWriter, request *http.Request) {
	//First of check if Get is set in the URL
	Filename := "data/" + request.URL.Query().Get("file")
	if Filename == "" {
		logrus.Error("Get 'file' not specified in url.")
		//Get not set, send a 400 bad request
		http.Error(response, "Get 'file' not specified in url.", 400)
		return
	}
	fmt.Println("Client requests: " + Filename)

	//Check if file exists and open
	Openfile, err := os.Open(Filename)
	defer util.CloseOSFile(Openfile) //Close after function return
	if err != nil {
		logrus.Error(err)
		//File not found, send 404
		http.Error(response, "File not found.", 404)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	_, err = Openfile.Read(FileHeader)
	if err != nil {
		http.Error(response, errors.FullTrace(err), http.StatusInternalServerError)
		return
	}

	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	response.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	response.Header().Set("Content-Type", FileContentType)
	response.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already so we reset the offset back to 0
	_, err = Openfile.Seek(0, 0)
	if err != nil {
		http.Error(response, errors.FullTrace(err), http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(response, Openfile) //'Copy' the file to the client
	if err != nil {
		http.Error(response, errors.FullTrace(err), http.StatusInternalServerError)
		return
	}
}
