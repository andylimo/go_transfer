package actions

import (
	"net/http"

	"github.com/tiger5226/filetransfer/actions/jenkinsfile"
	"github.com/tiger5226/filetransfer/orderedmap"

	"github.com/lbryio/lbry.go/extras/api"
)

// Routes an ordered map of the various endpoints for the API server
type Routes struct {
	m *orderedmap.Map
}

// Set Adds an API handler to the system based on a specific key
func (r *Routes) Set(key string, h api.Handler) {
	if r.m == nil {
		r.m = orderedmap.New()
	}
	r.m.Set(key, h)
}

// GetRoutes Generates the entire set of routes to be used by the API server
func GetRoutes() *Routes {
	routes := Routes{}

	routes.Set("/", Root)
	routes.Set("/test", Test)

	routes.Set("/bucket/list", List)
	routes.Set("/jenkinsfile/list", jenkinsfile.List)
	routes.Set("/jenkinsfile/publish", jenkinsfile.Publish)

	return &routes
}

// Each Iterates through all of the routes based on a lambda function
func (r *Routes) Each(f func(string, http.Handler)) {
	if r.m == nil {
		return
	}
	for _, k := range r.m.Keys() {
		a, _ := r.m.Get(k)
		f(k, a.(http.Handler))
	}
}

// Walk Iterates & updates all of the routes based on a lambda function
func (r *Routes) Walk(f func(string, http.Handler) http.Handler) {
	if r.m == nil {
		return
	}
	for _, k := range r.m.Keys() {
		a, _ := r.m.Get(k)
		r.m.Set(k, f(k, a.(http.Handler)))
	}
}
