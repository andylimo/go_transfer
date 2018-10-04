package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

func HandlerDownload(response http.ResponseWriter, request *http.Request) {
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
	defer Openfile.Close() //Close after function return
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
	Openfile.Read(FileHeader)
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
	Openfile.Seek(0, 0)
	io.Copy(response, Openfile) //'Copy' the file to the client
	return

}
