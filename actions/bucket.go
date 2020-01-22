package actions

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lbryio/lbry.go/extras/api"
	"github.com/lbryio/lbry.go/extras/errors"
	v "github.com/lbryio/ozzo-validation"
	"github.com/lbryio/ozzo-validation/is"
	"github.com/sirupsen/logrus"
)

// List generates a list of all the available buckets
func List(r *http.Request) api.Response {
	params := struct {
		Bucket *string
	}{}

	err := api.FormValues(r, &params, []*v.FieldRules{
		v.Field(&params.Bucket, is.PrintableASCII),
	})
	currDir, err := os.Getwd()
	if err != nil {
		return api.Response{Error: errors.Err(err)}
	}
	root := currDir + "/data"
	var buckets []*ftBucket
	var bucket *ftBucket
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		logrus.Info("Path: ", filepath.FromSlash(path), " Root: ", filepath.FromSlash(root))
		bucketName := strings.ReplaceAll(filepath.FromSlash(path), filepath.FromSlash(root), "")
		bucketName = strings.TrimPrefix(bucketName, string(os.PathSeparator))
		bucketName = strings.ReplaceAll(bucketName, string(os.PathSeparator), "-")

		if root == path {
			return nil
		}

		if info.IsDir() && (bucket == nil || bucket.Name != info.Name()) {
			bucket = &ftBucket{Name: bucketName, Files: make([]*ftBucketFile, 0)}
			if params.Bucket == nil || strings.Contains(path, *params.Bucket) {
				buckets = append(buckets, bucket)
			}
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
