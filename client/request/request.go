
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	dc "k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
	utiltesting "k8s.io/client-go/util/testing"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"
)


func main() {
//	TestDoRequestGet()
//	TestDoRequestNewWayObj()
//	TestDoRequestNewWayReader()
//	TestBackoffLifecycle()
// testGet()
//	TestStream()
//	TestDiscovery()
	TestGetOpenAPISchema()

}

func TestDoRequestNewWayObj() {
	reqObj := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
	//reqBodyExpected, _ := runtime.Encode(scheme.Codecs.LegacyCodec(v1.SchemeGroupVersion), reqObj)
	expectedObj := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{
		Protocol:   "TCP",
		Port:       12345,
		TargetPort: intstr.FromInt(12345),
	}}}}
	expectedBody, _ := runtime.Encode(scheme.Codecs.LegacyCodec(v1.SchemeGroupVersion), expectedObj)
	fakeHandler := utiltesting.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(expectedBody),
	}
    //  Suffix("index.html")
	testServer := httptest.NewServer(&fakeHandler)
	defer testServer.Close()
	c := testRESTClient(testServer)

	r := c.Verb("POST").
		Suffix("index.html").
		Prefix("foo", "bar").
		Param("limit","500").
		Name("nginx").
		Resource("services").
		Timeout(time.Second).
		Body(reqObj)
	s := r.URL().String();
	fmt.Printf("namespace should be in order in path: %s", s)

	obj, err := r.DoRaw(context.Background())

	if err != nil {
		fmt.Printf("Unexpected error: %v %#v", err, err)
		return
	}
	if obj == nil {
		fmt.Printf("nil obj")
	}
	fmt.Printf("got %#v", string(obj))

}

func TestDoRequestNewWayReader() {

	reqObj := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
	reqBodyExpected, _ := runtime.Encode(scheme.Codecs.LegacyCodec(v1.SchemeGroupVersion), reqObj)
	expectedObj := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{
		Protocol:   "TCP",
		Port:       12345,
		TargetPort: intstr.FromInt(12345),
	}}}}

	expectedBody, _ := runtime.Encode(scheme.Codecs.LegacyCodec(v1.SchemeGroupVersion), expectedObj)
	fakeHandler := utiltesting.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(expectedBody),
	}
	testServer := httptest.NewServer(&fakeHandler)
	defer testServer.Close()
	c := testRESTClient(testServer)

	r := c.Verb("POST").
		Suffix("index.html").
		Prefix("foo", "bar").
		Param("limit","500").
		Name("nginx").
		Resource("services").
		Timeout(time.Second).
		Body(bytes.NewBuffer(reqBodyExpected))
	s := r.URL().String();
	fmt.Printf("namespace should be in order in path: %s", s)

	obj, err := r.DoRaw(context.Background())

	if err != nil {
		fmt.Printf("Unexpected error: %v %#v", err, err)
		return
	}
	if obj == nil {
		fmt.Printf("nil obj")
	} else if !apiequality.Semantic.DeepDerivative(expectedObj, obj) {
		fmt.Printf("Expected: %#v, got %#v", expectedObj, obj)
	}

}

// This test assumes that the client implementation backs off exponentially, for an individual request.
func TestBackoffLifecycle() {
	count := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		count++
		fmt.Printf("Attempt %d\n", count)
		if count == 5 || count == 9 {
			w.WriteHeader(http.StatusOK)
			fmt.Printf("http Status %v \n", "200")
			return
		}
		w.WriteHeader(http.StatusGatewayTimeout)
		fmt.Printf("http Status %v \n", "Timeout")
		return
	}))
	defer testServer.Close()
	c := testRESTClient(testServer)

	backoff := &rest.URLBackoff{
		// Use a fake backoff here to avoid flakes and speed the test up.
		Backoff: flowcontrol.NewBackOff(
			time.Duration(1)*time.Second,
			time.Duration(200)*time.Second,
		)}

	seconds := []int{0, 1, 2, 4, 8, 0, 1, 2, 4, 0}
	request := c.Verb("POST").Prefix("backofftest").Suffix("abc").BackOff(backoff)

	for _, sec := range seconds {
		thisBackoff := backoff.CalculateBackoff(request.URL())
		fmt.Printf("Current backoff %v ", thisBackoff)
		if thisBackoff != time.Duration(sec)*time.Second {
			fmt.Printf("Backoff is %v instead of %v", thisBackoff, sec)
		}
		request.DoRaw(context.Background())

	}

}

func TestStream() {
	expectedBody := "expected body"

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			panic("need flusher!")
		}
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedBody))
		flusher.Flush()
	}))
	defer testServer.Close()

	s := testRESTClient(testServer)
	readCloser, err := s.Get().Prefix("path/to/stream/thing").Stream(context.Background())
	if err != nil {
		fmt.Printf("unexpected error: %v", err)
	}
	defer readCloser.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(readCloser)
	resultBody := buf.String()

	if expectedBody != resultBody {
		fmt.Printf("Expected %s, got %s", expectedBody, resultBody)
	}
}

func testGet() {

	// RESTClientFor returns a RESTClient that satisfies the requested attributes on a client Config
	// object. Note that a RESTClient may require fields that are optional when initializing a Client.
	// A RESTClient created by this method is generic - it expects to operate on an API that follows
	// the Kubernetes conventions, but may not be the Kubernetes API.
	restclient, err := rest.RESTClientFor(getConfig())

	if err != nil {
		log.Fatalln(err)
	}
    r := restclient.Get().
		Namespace("default").
		Resource("pods").
		Name("gohttpdemo-5cb6954c5c-42bkc")
	s := r.URL().String();
	fmt.Printf("namespace should be in order in path: %s", s)

	obj := r.Do(context.Background())

	// /<namespace>/<resource>/<name>
	// GET https://apiserver/api/v1/namespaces/default/pods/gohttpdemo-5cb6954c5c-42bkc

	bytes, err := obj.Raw()
	if err != nil {
		fmt.Printf("%s: %sn", err.Error(), bytes)
	} else {
		fmt.Printf("%sn", bytes)
	}
}





func TestDoRequestGet() {
	//GET https://49.0.11.44:6443/api/v1/namespaces/default/pods?limit=500

	c, err := rest.RESTClientFor(getConfig())
	r := c.Get().
		Namespace("default").
		Resource("pods").
		Name("gohttpdemo-5cb6954c5c-42bkc")
	s := r.URL().String();

	fmt.Printf("namespace should be in order in path: %s", s)

	obj,err := r.DoRaw(context.Background())

	if err != nil {
		fmt.Printf("Unexpected error: %v %#v", err, err)
		return
	}

	fmt.Printf("got %#v", string(obj))

}

func TestDiscovery(){
	groupVersion := "v1"
	resource, err := ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		fmt.Printf("%s: %sn", err.Error(), resource)
	}
	fmt.Printf("%+v\n", resource)
}

// ServerResourcesForGroupVersion returns the supported resources for a group and version.
func  ServerResourcesForGroupVersion(groupVersion string) (resources *metav1.APIResourceList, err error) {
	url := url.URL{}
	restclient, err := rest.RESTClientFor(getConfig())

	dcClient := dc.NewDiscoveryClient(restclient)

	if len(groupVersion) == 0 {
		return nil, fmt.Errorf("groupVersion shouldn't be empty")
	}
	if len(dcClient.LegacyPrefix) > 0 && groupVersion == "v1" {
		url.Path = dcClient.LegacyPrefix + "/" + groupVersion
	} else {
		url.Path = "/apis/" + groupVersion
	}
	resources = &metav1.APIResourceList{
		GroupVersion: groupVersion,
	}
	err = dcClient.RESTClient().Get().AbsPath(url.String()).Do(context.TODO()).Into(resources)
	if err != nil {
		// ignore 403 or 404 error to be compatible with an v1.0 server.
		if groupVersion == "v1" && (errors.IsNotFound(err) || errors.IsForbidden(err)) {
			return resources, nil
		}
		return nil, err
	}
	return resources, nil
}

func TestGetOpenAPISchema() {

	restclient, err := rest.RESTClientFor(getConfig())
    dcClient := dc.NewDiscoveryClient(restclient)

	got, err := dcClient.OpenAPISchema()
	if err != nil {
		fmt.Println("unexpected error getting openapi: %v", err)
	}

	fmt.Println("got %v", got)

}


func defaultResourcePathWithPrefix(prefix, resource, namespace, name string) string {
	var path string
	path = "/api/" + v1.SchemeGroupVersion.Version

	if prefix != "" {
		path = path + "/" + prefix
	}
	if namespace != "" {
		path = path + "/namespaces/" + namespace
	}
	// Resource names are lower case.
	resource = strings.ToLower(resource)
	if resource != "" {
		path = path + "/" + resource
	}
	if name != "" {
		path = path + "/" + name
	}
	return path
}

func defaultContentConfig() rest.ClientContentConfig {
	gvCopy := v1.SchemeGroupVersion
	var Scheme = runtime.NewScheme()
	var Codecs = serializer.NewCodecFactory(Scheme)

	return rest.ClientContentConfig{
		ContentType:  "application/json",
		GroupVersion: gvCopy,
		Negotiator:   runtime.NewClientNegotiator(Codecs.WithoutConversion(), gvCopy),
	}
}

func testRESTClient(srv *httptest.Server) *rest.RESTClient {
	contentConfig := defaultContentConfig()
	return testRESTClientWithConfig(srv, contentConfig)
}

func testRESTClientWithConfig(srv *httptest.Server, contentConfig rest.ClientContentConfig)  *rest.RESTClient{
	base, _ := url.Parse("http://localhost")
	if srv != nil {
		var err error
		base, err = url.Parse(srv.URL)
		if err != nil {
			fmt.Printf("failed to parse test URL: %v", err)
		}
	}
	versionedAPIPath := defaultResourcePathWithPrefix("", "", "", "")
	client, err := rest.NewRESTClient(base, versionedAPIPath, contentConfig, nil, nil)

	if err != nil {
		fmt.Printf("failed to create a client: %v", err)
	}
	return client

}

func getConfig() (*rest.Config) {
	log.SetFlags(log.Llongfile)
	kubeconfig := flag.String("kubeconfig", "./config", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalln(err)
	}

	spew.Dump(config, err)

	scheme := runtime.NewScheme()

	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/api"
	config.ContentType = runtime.ContentTypeJSON

	config.NegotiatedSerializer = serializer.NewCodecFactory(scheme).WithoutConversion()
	return config

}

func groupVersions(resources []*metav1.APIResourceList) []string {
	result := []string{}
	for _, resourceList := range resources {
		result = append(result, resourceList.GroupVersion)
	}
	return result
}


