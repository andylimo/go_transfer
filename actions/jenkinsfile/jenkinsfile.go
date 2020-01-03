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

	"github.com/lbryio/lbry.go/extras/api"
	"github.com/lbryio/lbry.go/extras/errors"
	v "github.com/lbryio/ozzo-validation"
	"github.com/lbryio/ozzo-validation/is"
	"github.com/sirupsen/logrus"
)

type jenkinsFile struct {
	Name       string
	Contents   string
	ModifiedAt time.Time
}

type formRequestValues struct {
	Content    string
	Repository string
	Project    string
	Branch     string
	User       string
}

type bitbucketInfo struct {
	server   string
	username string
	password string
	hookURL  string
}
type webhook struct {
	ID                 int    `json:"id"`
	Title              string `json:"title"`
	URL                string `json:"url"`
	CommittersToIgnore string `json:"-"` //Possibly set in the future?
	BranchesToIgnore   string `json:"-"` //Possibly set in the future?
	Enabled            bool   `json:"enabled"`
}
type webhooks []webhook

func (c webhook) String() string {
	return fmt.Sprintf("[%d](%s): %s", c.ID, c.Title, c.URL)
}

// List generates a list of all possible jenkinsfiles to use
func List(r *http.Request) api.Response {
	files := make([]jenkinsFile, 0)
	currDir, err := os.Getwd()
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}
	root := currDir + "/jenkinsfiles"
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if root == path {
			return nil
		}
		var contents []byte
		contents, err = ioutil.ReadFile(path)
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

// generateWebhook sends a http request to the Bitbucket Server to trigger the creation of a post webhook.
func generateWebhook(projectKey string, repo string, bbInfo bitbucketInfo) error {
	url := fmt.Sprintf("%s/rest/webhook/1.0/projects/%s/repos/%s/configurations", bbInfo.server, projectKey, repo)

	// write the fields
	body := &bytes.Buffer{}
	newWebhook := webhook{Title: "Jenkins DQCI Webhook", URL: bbInfo.hookURL, Enabled: true}
	jsonData, err := json.Marshal(newWebhook)
	if err != nil {
		return err
	}
	body.Write(jsonData)

	request, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(bbInfo.username, bbInfo.password)
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	_, err = handleResponseBasic(resp, "failed to create webhook")
	if err != nil {
		return err
	}

	return nil
}

// listWebhooks sends a http request to the Bitbucket Server for a list of all post webhooks in a repository
func listWebhooks(projectKey string, repo string) ([]webhook, error) {
	hooks := webhooks{}

	// Get BB creds
	bbInfo, err := getBitbucketEnvs(false)
	if err != nil {
		return nil, err
	}

	fakelist := []string{}
	err = json.Unmarshal([]byte(`[]`), &fakelist)
	fmt.Println(err)
	fmt.Println(fakelist)

	url := fmt.Sprintf("%s/rest/webhook/1.0/projects/%s/repos/%s/configurations", bbInfo.server, projectKey, repo)

	reqBody := &bytes.Buffer{}
	request, err := http.NewRequest(http.MethodGet, url, reqBody)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(bbInfo.username, bbInfo.password)
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	respBody, handleErr := handleResponseBasic(resp, "failed to get webhooks list")
	if handleErr != nil {
		return nil, handleErr
	}

	err = json.Unmarshal(respBody, &hooks)
	if err != nil {
		return nil, err
	}

	fmt.Println("===== Webhooks =====")
	for i, hook := range hooks {
		fmt.Println(i, hook)
	}

	return hooks, nil
}

// sendCommitFileRequest sends a http request to the BB Server to commit the contents of a file.  If the file already exits, an error is thrown.
func sendCommitFileRequest(projectKey string, repo string, content string, branch string, user string) error {

	// Get BB creds
	bbInfo, err := getBitbucketEnvs(false)
	if err != nil {
		return err
	}

	path := "Jenkinsfile"
	url := fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s/browse/%s", bbInfo.server, projectKey, repo, path)

	// Generate the body of the request
	body, writer, err := generateFormDataCommitBody(content, branch, user)
	if err != nil {
		return errors.Err(err)
	}

	request, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return errors.Err(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.SetBasicAuth(bbInfo.username, bbInfo.password)
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return errors.Err(err)
	}

	_, err = handleResponseBasic(resp, "failed to publish Jenkinsfile")
	if err != nil {
		return errors.Err(err)
	}

	return nil
}

// sendCreateWebhookRequest sends a http request to the BB Server to create a webhook.  If the hook already exists, it is not created.
func sendCreateWebhookRequest(projectKey string, repo string) error {
	// Retrieve the BB creds
	bbInfo, err := getBitbucketEnvs(true)
	if err != nil {
		return errors.Err(err)
	}

	// Generate the list of current webhooks
	webhooks, err := listWebhooks(projectKey, repo)
	if err != nil {
		return errors.Err(err)
	}

	// Check if the list contains our hook URL already
	if contains(webhooks, bbInfo.hookURL) {
		logrus.Info("Webhook has already been generated.")
		return nil
	}

	// we didn't exit yet, so we'll need to generate the webhook
	err = generateWebhook(projectKey, repo, bbInfo)
	if err != nil {
		return errors.Err(err)
	}

	return nil
}

// getBitbucketEnvs Validates and returns env variables as a single bitbucket object.
func getBitbucketEnvs(requiresHook bool) (bitbucketInfo, error) {
	info := bitbucketInfo{
		server:   os.Getenv("BITBUCKET_URL"),
		username: os.Getenv("BITBUCKET_USERNAME"),
		password: os.Getenv("BITBUCKET_PASSWORD"),
		hookURL:  os.Getenv("BITBUCKET_HOOKURL"),
	}

	if len(info.server) == 0 {
		return info, errors.Base("unable to find bitbucket URL from environment variables")
	}

	if len(info.username) == 0 || len(info.password) == 0 {
		return info, errors.Base("unable to find credentials for bitbucket")
	}

	if requiresHook && len(info.hookURL) == 0 {
		return info, errors.Base("unable to find credentials for bitbucket")
	}

	logrus.Info("==== Bitbucket Info ===")
	logrus.Info("server: ", info.server)
	logrus.Info("hookUrl: ", info.hookURL)

	return info, nil
}

// Publish Publishes both the jenkinsfile and webhook for DQCI
func Publish(r *http.Request) api.Response {
	params := formRequestValues{}

	err := api.FormValues(r, &params, []*v.FieldRules{
		v.Field(&params.Content, is.ASCII, v.Required),
		v.Field(&params.Repository, is.ASCII, v.Required),
		v.Field(&params.User, is.ASCII, v.Required),
		v.Field(&params.Project, is.ASCII, v.Required),
		v.Field(&params.Branch, is.ASCII, v.Required),
	})
	if err != nil {
		return api.Response{Error: err, Status: http.StatusBadRequest}
	}

	// First, publish the Jenkinsfile
	err = sendCommitFileRequest(params.Project, params.Repository, params.Content, params.Branch, params.User)
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}

	// Now that we have created the Jenkinsfile, we need to publish the webhook
	err = sendCreateWebhookRequest(params.Project, params.Repository)
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}

	return api.Response{Data: "OK"}
}

// PublishJenkinsfile Publishes a jenkinsfile based on the user's selection
func PublishJenkinsfile(r *http.Request) api.Response {
	params := formRequestValues{}

	err := api.FormValues(r, &params, []*v.FieldRules{
		v.Field(&params.Content, is.ASCII, v.Required),
		v.Field(&params.Repository, is.ASCII, v.Required),
		v.Field(&params.User, is.ASCII, v.Required),
		v.Field(&params.Project, is.ASCII, v.Required),
		v.Field(&params.Branch, is.ASCII, v.Required),
	})
	if err != nil {
		return api.Response{Error: err, Status: http.StatusBadRequest}
	}
	err = sendCommitFileRequest(params.Project, params.Repository, params.Content, params.Branch, params.User)
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}

	return api.Response{Data: "OK"}
}

// PublishWebhooks Checks to see if we have the DQCI webhook created on the repository.  If not, we create one.
func PublishWebhooks(r *http.Request) api.Response {
	params := formRequestValues{}

	// validate the parameters
	err := api.FormValues(r, &params, []*v.FieldRules{
		v.Field(&params.Repository, is.ASCII),
		v.Field(&params.Project, is.ASCII),

		// possible to be found, but not required.
		v.Field(&params.Content),
		v.Field(&params.User),
		v.Field(&params.Branch),
	})
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}

	err = sendCreateWebhookRequest(params.Project, params.Repository)
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}

	return api.Response{Data: "OK"}
}

// contains Checks to see if we have a matching url in a set of webhooks
func contains(hooks []webhook, url string) bool {
	for _, hook := range hooks {
		if hook.URL == url {
			return true
		}
	}
	return false
}

// handleResponseBasic Simple response handler, prints some simple information and will fail if any status is 300 or greater
func handleResponseBasic(resp *http.Response, failMessage string) ([]byte, error) {
	body := &bytes.Buffer{}
	_, err := body.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// Output some response information
	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header)
	fmt.Println(body.String())

	if resp.StatusCode >= 300 {
		e := make(map[string]interface{})
		err := json.Unmarshal(body.Bytes(), &e)
		if err != nil {
			return nil, err
		}
		return nil, errors.Err(failMessage, e)
	}

	return body.Bytes(), nil
}

// generateFormDataCommitBody Generates FormData body for a commit http request to a BitBucket server
func generateFormDataCommitBody(content string, branch string, user string) (*bytes.Buffer, *multipart.Writer, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("content", content)
	_ = writer.WriteField("branch", branch)
	_ = writer.WriteField("message", "auto commit from file transfer by '"+user+"'")
	err := writer.Close()
	if err != nil {
		return nil, writer, errors.Err(err)
	}

	return body, writer, nil
}
