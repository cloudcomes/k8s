package discovery

import (
	"encoding/json"
	"fmt"
	openapi_v2 "github.com/googleapis/gnostic/OpenAPIv2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"mime"
	"net/http"
	"net/http/httptest"
	"reflect"


	"time"

	"github.com/gogo/protobuf/proto"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	//"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/version"
	dc "k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
)

func TestGetServerVersion() {
	expect := version.Info{
		Major:     "foo",
		Minor:     "bar",
		GitCommit: "baz",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		output, err := json.Marshal(expect)
		if err != nil {
			fmt.Println("unexpected encoding error: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(output)
	}))
	defer server.Close()
	client := dc.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL})

	got, err := client.ServerVersion()
	//fmt.Printf("%v %+v %#v\n", expect, expect, expect)
	fmt.Printf("%v %+v %#v\n", got, got, got)
	if err != nil {
		fmt.Println("unexpected encoding error: %v", err)
	}
	if e, a := expect, *got; !reflect.DeepEqual(e, a) {
		fmt.Println("expected %v, got %v", e, a)
	}
}

func TestGetServerResources() {
	stable := metav1.APIResourceList{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{
			{Name: "pods", Namespaced: true, Kind: "Pod"},
			{Name: "services", Namespaced: true, Kind: "Service"},
			{Name: "namespaces", Namespaced: false, Kind: "Namespace"},
		},
	}
	beta := metav1.APIResourceList{
		GroupVersion: "extensions/v1beta1",
		APIResources: []metav1.APIResource{
			{Name: "deployments", Namespaced: true, Kind: "Deployment"},
			{Name: "ingresses", Namespaced: true, Kind: "Ingress"},
			{Name: "jobs", Namespaced: true, Kind: "Job"},
		},
	}
	beta2 := metav1.APIResourceList{
		GroupVersion: "extensions/v1beta2",
		APIResources: []metav1.APIResource{
			{Name: "deployments", Namespaced: true, Kind: "Deployment"},
			{Name: "ingresses", Namespaced: true, Kind: "Ingress"},
			{Name: "jobs", Namespaced: true, Kind: "Job"},
		},
	}

	appsbeta1 := metav1.APIResourceList{GroupVersion: "apps/v1beta1", APIResources: []metav1.APIResource{{Name: "deployments", Namespaced: true, Kind: "Deployment"}}}
	appsbeta2 := metav1.APIResourceList{GroupVersion: "apps/v1beta2", APIResources: []metav1.APIResource{{Name: "deployments", Namespaced: true, Kind: "Deployment"}}}

	tests := []struct {
		resourcesList *metav1.APIResourceList
		path          string
		request       string
		expectErr     bool
	}{
		{
			resourcesList: &stable,
			path:          "/api/v1",
			request:       "v1",
			expectErr:     false,
		},
		{
			resourcesList: &beta,
			path:          "/apis/extensions/v1beta1",
			request:       "extensions/v1beta1",
			expectErr:     false,
		},
		{
			resourcesList: &stable,
			path:          "/api/v1",
			request:       "foobar",
			expectErr:     true,
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var list interface{}
		switch req.URL.Path {
		case "/api/v1":
			list = &stable
		case "/apis/extensions/v1beta1":
			list = &beta
		case "/apis/extensions/v1beta2":
			list = &beta2
		case "/apis/apps/v1beta1":
			list = &appsbeta1
		case "/apis/apps/v1beta2":
			list = &appsbeta2
		case "/api":
			list = &metav1.APIVersions{
				Versions: []string{
					"v1",
				},
			}
		case "/apis":
			list = &metav1.APIGroupList{
				Groups: []metav1.APIGroup{
					{
						Name: "apps",
						Versions: []metav1.GroupVersionForDiscovery{
							{GroupVersion: "apps/v1beta1", Version: "v1beta1"},
							{GroupVersion: "apps/v1beta2", Version: "v1beta2"},
						},
					},
					{
						Name: "extensions",
						Versions: []metav1.GroupVersionForDiscovery{
							{GroupVersion: "extensions/v1beta1", Version: "v1beta1"},
							{GroupVersion: "extensions/v1beta2", Version: "v1beta2"},
						},
					},
				},
			}
		default:
			fmt.Println("unexpected request: %s", req.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		output, err := json.Marshal(list)
		if err != nil {
			fmt.Println("unexpected encoding error: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(output)
	}))
	defer server.Close()
	for _, test := range tests {
		client := dc.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL})
		got, err := client.ServerResourcesForGroupVersion(test.request)
		//fmt.Printf("%v %+v %#v\n", got, got, got)
		fmt.Printf("%+v\n", got)
		if test.expectErr {
			if err == nil {
				fmt.Println("unexpected non-error")
			}
			continue
		}
		if err != nil {
			fmt.Println("unexpected error: %v", err)
			continue
		}
		if !reflect.DeepEqual(got, test.resourcesList) {
			fmt.Println("expected:\n%v\ngot:\n%v\n", test.resourcesList, got)
		}
	}

	client := dc.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL})
	start := time.Now()
	serverResources, err := client.ServerResources()
	//fmt.Printf("%v %+v %#v\n", serverResources, serverResources, serverResources)
	fmt.Printf("%+v\n", serverResources)
	if err != nil {
		fmt.Println("unexpected error: %v", err)
	}
	end := time.Now()
	if d := end.Sub(start); d > time.Second {
		fmt.Println("took too long to perform discovery: %s", d)
	}
	serverGroupVersions := groupVersions(serverResources)
	expectedGroupVersions := []string{
		"v1",
		"apps/v1beta1",
		"apps/v1beta2",
		"extensions/v1beta1",
		"extensions/v1beta2",
	}
	if !reflect.DeepEqual(expectedGroupVersions, serverGroupVersions) {
		fmt.Println("unexpected group versions: %v", diff.ObjectReflectDiff(expectedGroupVersions, serverGroupVersions))
	}
}

var returnedOpenAPI = openapi_v2.Document{
	Definitions: &openapi_v2.Definitions{
		AdditionalProperties: []*openapi_v2.NamedSchema{
			{
				Name: "fake.type.1",
				Value: &openapi_v2.Schema{
					Properties: &openapi_v2.Properties{
						AdditionalProperties: []*openapi_v2.NamedSchema{
							{
								Name: "count",
								Value: &openapi_v2.Schema{
									Type: &openapi_v2.TypeItem{
										Value: []string{"integer"},
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "fake.type.2",
				Value: &openapi_v2.Schema{
					Properties: &openapi_v2.Properties{
						AdditionalProperties: []*openapi_v2.NamedSchema{
							{
								Name: "count",
								Value: &openapi_v2.Schema{
									Type: &openapi_v2.TypeItem{
										Value: []string{"array"},
									},
									Items: &openapi_v2.ItemsItem{
										Schema: []*openapi_v2.Schema{
											{
												Type: &openapi_v2.TypeItem{
													Value: []string{"string"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

func openapiSchemaFakeServer() (*httptest.Server, error) {
	var sErr error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/openapi/v2" {
			sErr = fmt.Errorf("Unexpected url %v", req.URL)
		}
		if req.Method != "GET" {
			sErr = fmt.Errorf("Unexpected method %v", req.Method)
		}
		decipherableFormat := req.Header.Get("Accept")
		if decipherableFormat != "application/com.github.proto-openapi.spec.v2@v1.0+protobuf" {
			sErr = fmt.Errorf("Unexpected accept mime type %v", decipherableFormat)
		}

		mime.AddExtensionType(".pb-v1", "application/com.github.googleapis.gnostic.OpenAPIv2@68f4ded+protobuf")

		output, err := proto.Marshal(&returnedOpenAPI)
		if err != nil {
			sErr = err
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(output)
	}))
	return server, sErr
}

func TestGetOpenAPISchema() {
	server, err := openapiSchemaFakeServer()
	if err != nil {
		fmt.Println("unexpected error starting fake server: %v", err)
	}
	defer server.Close()

	client := dc.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL})
	got, err := client.OpenAPISchema()
	if err != nil {
		fmt.Println("unexpected error getting openapi: %v", err)
	}
	if e, a := returnedOpenAPI, *got; !reflect.DeepEqual(e, a) {
		fmt.Println("expected %v, got %v", e, a)
	}
}

func groupVersions(resources []*metav1.APIResourceList) []string {
	result := []string{}
	for _, resourceList := range resources {
		result = append(result, resourceList.GroupVersion)
	}
	return result
}
