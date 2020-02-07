/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webservice

import (
	"github.com/emicklei/go-restful"
	"io"
)

// legacyRootAPIHandler creates a webservice serving api group discovery.
type legacyRootAPIHandler struct {
	// addresses is used to build cluster IPs for discovery.

	apiPrefix  string

}

func NewLegacyRootAPIHandler( apiPrefix string) *legacyRootAPIHandler {
	// Because in release 1.1, /apis returns response with empty APIVersion, we
	// use stripVersionNegotiatedSerializer to keep the response backwards
	// compatible.

	return &legacyRootAPIHandler{

		apiPrefix:  apiPrefix,

	}
}

// AddApiWebService adds a service to return the supported api versions at the legacy /api.
func (s *legacyRootAPIHandler) WebService() *restful.WebService {

	ws := new(restful.WebService)
	ws.Path(s.apiPrefix)
	ws.Doc("get available API versions")
	ws.Route(ws.GET("/").To(s.handle))

	return ws
}

func (s *legacyRootAPIHandler) handle(req *restful.Request, resp *restful.Response) {
	io.WriteString(resp, "GoRestfulContainer hanlder\n")
}
