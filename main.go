package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"fmt"
	"io"

	"github.com/kabukky/httpscerts"
	"github.com/sirupsen/logrus"
)

func main() {
	err := FindCreateCerts()
	if err != nil {
		logrus.Panic(err)
	}

	// Set up the HTTP server:
	serverMUX := http.NewServeMux()
	serverMUX.HandleFunc("/upload", handlerUpload)
	serverMUX.HandleFunc("/download", handlerDownload)
	serverMUX.HandleFunc("/echo", echoRequest)

	server := &http.Server{}
	server.Addr = ":9999"
	server.Handler = serverMUX
	server.SetKeepAlivesEnabled(true)
	server.ReadTimeout = 60 * 120 * time.Second // 2 hours
	server.WriteTimeout = 16 * time.Second

	// Start the server:

	logrus.Info("The HTTPS web server starts now on https://127.0.0.1" + server.Addr)
	if errHTTP := server.ListenAndServeTLS("cert.pem", "key.pem"); errHTTP != nil {
		logrus.Info("Was not able to start the HTTP server: ", errHTTP)
		os.Exit(2)
	}
}

func FindCreateCerts() error {
	err := httpscerts.Check("cert.pem", "key.pem")
	if err != nil {
		err = httpscerts.Generate("cert.pem", "key.pem", "127.0.0.1:9999")
		if err != nil {
			logrus.Fatal("Couldn't create https certs.", err)
			return err
		}
	}

	return nil
}

func echoRequest(response http.ResponseWriter, request *http.Request) {
	requestDump, err := httputil.DumpRequest(request, true)
	if err != nil {
		logrus.Error("ERROR DUMPING:", err)
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}
	logrus.Info("DUMPING REQUEST SUCCESS:", string(requestDump))
	response.Write([]byte(string(requestDump)))
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
			err = ioutil.WriteFile("/data/"+sourceFilename, data, 0600)
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

func handlerDownload(response http.ResponseWriter, request *http.Request) {
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

func SendResponse(w http.ResponseWriter, data interface{}) {

}
