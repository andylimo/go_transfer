package jenkinsfile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/lbryio/ozzo-validation/is"

	"github.com/lbryio/lbry.go/extras/api"
	"github.com/lbryio/lbry.go/extras/errors"
	v "github.com/lbryio/ozzo-validation"
)

type jenkinsFile struct {
	Name       string
	Contents   string
	ModifiedAt time.Time
}

func List(r *http.Request) api.Response {
	files := make([]jenkinsFile, 0)
	currDir, err := os.Getwd()
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}
	root := currDir + "/jenkinsfiles"
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if root == path {
			return nil
		}
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Err(err)
		}
		file := jenkinsFile{info.Name(), string(contents), info.ModTime()}
		files = append(files, file)

		return nil
	})
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}

	return api.Response{Data: files}
}

func Publish(r *http.Request) api.Response {
	params := struct {
		Content    string
		Repository string
		Project    string
		Branch     string
		User       string
	}{}

	err := api.FormValues(r, &params, []*v.FieldRules{
		v.Field(&params.Content, is.ASCII),
		v.Field(&params.Repository, is.ASCII),
		v.Field(&params.User, is.ASCII),
		v.Field(&params.Project, is.ASCII),
		v.Field(&params.Branch, is.ASCII),
	})
	if err != nil {
		return api.Response{Error: err, Status: http.StatusBadRequest}
	}
	projectKey := params.Project
	repo := params.Repository
	path := "Jenkinsfile"
	bitbucketURL := os.Getenv("BITBUCKET_URL")
	url := fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s/browse/%s", bitbucketURL, projectKey, repo, path)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("content", params.Content)
	writer.WriteField("branch", params.Branch)
	writer.WriteField("message", "auto commit from file transfer by '"+params.User+"'")
	writer.WriteField("content", params.Content)
	err = writer.Close()
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}

	request, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	username := os.Getenv("BITBUCKET_USERNAME")
	password := os.Getenv("BITBUCKET_PASSWORD")
	request.SetBasicAuth(username, password)
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			return api.Response{Error: errors.Err(err)}
		}
		resp.Body.Close()
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
		fmt.Println(string(body.Bytes()))
		if resp.StatusCode >= 300 {
			e := make(map[string]interface{})
			err := json.Unmarshal(body.Bytes(), &e)
			if err != nil {
				return api.Response{Error: errors.Err(err)}
			}
			return api.Response{Error: errors.Err("failed to publish Jenkinsfile"), Data: e}
		}
	}
	return api.Response{Data: "OK"}
}
