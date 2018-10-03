package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {

	// Set up the HTTP server:
	serverMUX := http.NewServeMux()
	serverMUX.HandleFunc("/upload", handlerUpload)

	server := &http.Server{}
	server.Addr = ":9999"
	server.Handler = serverMUX
	server.SetKeepAlivesEnabled(true)
	server.ReadTimeout = 60 * 120 * time.Second // 2 hours
	server.WriteTimeout = 16 * time.Second

	// Start the server:

	logrus.Info("The HTTP web server starts now on http://127.0.0.1" + server.Addr)
	if errHTTP := server.ListenAndServe(); errHTTP != nil {
		logrus.Info("Was not able to start the HTTP server: ", errHTTP)
		os.Exit(2)
	}
}

func handlerUpload(response http.ResponseWriter, request *http.Request) {
	var err error
	file, fileHeader, err := request.FormFile(`file`)
	if err != nil {
		logrus.Error("Was not able to access the uploaded file: ", err)
	} else {

		// Close the file afterwards:
		defer file.Close()

		// Get the original filename:
		sourceFilename := fileHeader.Filename

		// Read the entire file into memory:
		if data, err := ioutil.ReadAll(file); err != nil {
			logrus.Error("Error while reading file from client: ", err)
		} else {

			// Write the data into a new file on server's side:
			err = ioutil.WriteFile(sourceFilename, data, 0600)
			if err != nil {
				logrus.Error("ERROR:", err)
			}
			logrus.Info("Server: File was read from client and written to disk.")
		}
	}
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	} else {
		response.Write([]byte("OK"))
		return
	}
}

func SendResponse(w http.ResponseWriter, data interface{}) {

}
