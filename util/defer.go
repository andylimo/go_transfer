package util

import (
	"mime/multipart"
	"os"

	"github.com/lbryio/lbry.go/extras/errors"
	"github.com/sirupsen/logrus"
)

//CloseOSFile closes the file safely while still reporting the error to the logs
func CloseOSFile(f *os.File) {
	err := f.Close()
	if err != nil {
		logrus.Error(errors.Err(err))
	}
}

//CloseMPFile closes the file safely while still reporting the error to the logs
func CloseMPFile(f multipart.File) {
	err := f.Close()
	if err != nil {
		logrus.Error(errors.Err(err))
	}
}
