package handler

import (
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

func HandlerUpload(response http.ResponseWriter, request *http.Request) {
	var err error
	file, fileHeader, err := request.FormFile(`file`)
	if err != nil {
		logrus.Error("Was not able to access the uploaded file: ", err)
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	} else {

		// Close the file afterwards:
		defer file.Close()

		// Get the original filename:
		sourceFilename := fileHeader.Filename

		// Read the entire file into memory:
		data, err := ioutil.ReadAll(file)
		if err != nil {
			logrus.Error("Error while reading file from client: ", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		} else {

			// Write the data into a new file on server's side:
			err = ioutil.WriteFile("data/"+sourceFilename, data, 0600)
			if err != nil {
				logrus.Error("ERROR:", err)
				http.Error(response, err.Error(), http.StatusInternalServerError)
				return
			}
			logrus.Info("Server: File was read from client and written to disk.")
			response.Write([]byte("OK"))
			return
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
