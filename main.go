package main

import (
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/tiger5226/filetransfer/handler"

	"github.com/kabukky/httpscerts"
	"github.com/sirupsen/logrus"
)

func main() {
	err := findCreateCerts() //
	if err != nil {
		logrus.Panic(err)
	}

	// Set up the HTTP server:
	serverMUX := http.NewServeMux()
	serverMUX.HandleFunc("/upload", handler.Upload)
	serverMUX.HandleFunc("/download", handler.Download)
	serverMUX.HandleFunc("/echo", echoRequest)

	server := &http.Server{}
	server.Addr = ":9999"
	server.Handler = serverMUX
	server.SetKeepAlivesEnabled(true)
	server.ReadTimeout = 60 * time.Second // 2 hours
	server.WriteTimeout = 15 * time.Minute

	// Start the server:

	logrus.Info("The HTTPS File Transfer web server starts now on https://127.0.0.1" + server.Addr)
	if errHTTP := server.ListenAndServeTLS("cert.pem", "key.pem"); errHTTP != nil {
		logrus.Info("Was not able to start the HTTP server: ", errHTTP)
		os.Exit(2)
	}
}

func findCreateCerts() error {
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
