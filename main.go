package main

import (
	"context"
	"filetransfer/actions"
	"filetransfer/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kabukky/httpscerts"
	"github.com/sirupsen/logrus"
)

func main() {
	err := findCreateCerts()
	if err != nil {
		logrus.Panic(err)
	}

	currDir, err := os.Getwd()
	if err != nil {
		logrus.Panic(err)
	}

	logrus.Infof("Current Working Directory: %s", currDir)

	// Set up routes -
	serverMUX := http.NewServeMux()
	routes := actions.GetRoutes()
	//Specialty Handlers for Data Upload/Download
	serverMUX.HandleFunc("/upload", handler.Upload)
	serverMUX.HandleFunc("/download", handler.Download)
	routes.Each(func(pattern string, handler http.Handler) {
		serverMUX.Handle(pattern, handler)
	})

	// Set up the HTTP server:
	server := &http.Server{}
	server.Addr = ":9999"
	server.Handler = serverMUX
	server.SetKeepAlivesEnabled(true)
	server.ReadTimeout = 15 * time.Minute
	server.WriteTimeout = 15 * time.Minute

	actions.ConfigureAPIServer()

	// Start the server:
	logrus.Printf("Listening on port %v", 9999)
	go func() {
		err := server.ListenAndServeTLS("cert.pem", "key.pem")
		if err != nil {
			//Normal graceful shutdown error
			if err.Error() == "http: Server closed" {
				logrus.Info(err)
			} else {
				log.Fatal(err)
			}
		}
	}()
	//Wait for shutdown signal, then shutdown api server. This will wait for all connections to finish.
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	<-interruptChan
	logrus.Debug("Shutting down API server...")
	err = server.Shutdown(context.Background())
	if err != nil {
		logrus.Error("Error shutting down server: ", err)
	}
	logrus.Debug("Simple File Transfer is shutting down...")

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
