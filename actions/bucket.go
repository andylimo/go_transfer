package actions

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/lbryio/lbry.go/extras/api"
	"github.com/lbryio/lbry.go/extras/errors"
)

// List generates a list of all the available buckets
func List(r *http.Request) api.Response {
	currDir, err := os.Getwd()
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}
	root := currDir + "/data" //
	var buckets []*ftBucket
	var bucket *ftBucket
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if root == path {
			return nil
		}
		if info.IsDir() && (bucket == nil || bucket.Name != info.Name()) {
			bucket = &ftBucket{Name: info.Name(), Files: make([]*ftBucketFile, 0)}
			buckets = append(buckets, bucket)
		} else {
			file := &ftBucketFile{info.Name(), info.Size(), info.ModTime()}
			bucket.Files = append(bucket.Files, file)
		}

		return nil
	})
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}

	return api.Response{Data: buckets}
}

type ftBucket struct {
	Name  string
	Files []*ftBucketFile
}

type ftBucketFile struct {
	Name       string
	Size       int64
	ModifiedAt time.Time
}
