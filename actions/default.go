package actions

import (
	"filetransfer/util"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/fatih/color"
	"github.com/lbryio/lbry.go/extras/api"
	"github.com/lbryio/lbry.go/extras/errors"
)

// Root Handler is the default handler
func Root(r *http.Request) api.Response {
	if r.URL.Path == "/" {
		return api.Response{Data: "Welcome to Simple File Transfer"}
	}
	return api.Response{Status: http.StatusNotFound, Error: errors.Err("404 Not Found")}
}

// Test simple test method for the handler
func Test(r *http.Request) api.Response {
	return api.Response{Data: "ok"}
}

// ConfigureAPIServer Handles the default configuration of the API Server
func ConfigureAPIServer() {
	api.TraceEnabled = util.Debugging

	api.IgnoredFormFields = []string{""}

	hs := make(map[string]string)

	hs["Server"] = "3ds.com"
	hs["Content-Type"] = "application/json; charset=utf-8"
	hs["Access-Control-Allow-Methods"] = "GET, PUT, POST, DELETE, OPTIONS"
	hs["Access-Control-Allow-Origin"] = "*"
	hs["X-Content-Type-Options"] = "nosniff"
	hs["X-Frame-Options"] = "deny"
	hs["Content-Security-Policy"] = "default-src 'none'"
	hs["X-XSS-Protection"] = "1; mode=block"
	hs["Referrer-Policy"] = "same-origin"

	api.ResponseHeaders = hs

	api.Log = func(request *http.Request, response *api.Response, err error) {
		consoleText := request.RemoteAddr + " [" + strconv.Itoa(response.Status) + "]: " + request.Method + " " + request.URL.Path
		if err == nil {
			logrus.Debug(color.GreenString(consoleText))
		} else {
			logrus.Error(color.RedString(consoleText + ": " + err.Error()))
		}
	}
}
