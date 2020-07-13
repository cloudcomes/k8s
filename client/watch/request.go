
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	runtimejson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/apimachinery/pkg/watch"
	dc "k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
	restclientwatch "k8s.io/client-go/rest/watch"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"

	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)


func main() {
  TestWatch()
//TestEncodeDecodeRoundTrip()

}


// getDecoder mimics how k8s.io/client-go/rest.createSerializers creates a decoder
func getDecoder() runtime.Decoder {
	jsonSerializer := runtimejson.NewSerializer(runtimejson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, false)
	directCodecFactory := scheme.Codecs.WithoutConversion()
	return directCodecFactory.DecoderToVersion(jsonSerializer, v1.SchemeGroupVersion)
}


// getEncoder mimics how k8s.io/client-go/rest.createSerializers creates a encoder
func getEncoder() runtime.Encoder {
	jsonSerializer := runtimejson.NewSerializer(runtimejson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, false)
	directCodecFactory := scheme.Codecs.WithoutConversion()
	return directCodecFactory.EncoderForVersion(jsonSerializer, v1.SchemeGroupVersion)
}

func TestEncodeDecodeRoundTrip() {
	testCases := []struct {
		Type   watch.EventType
		Object runtime.Object
	}{
		{
			watch.Added,
			&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
		},
		{
			watch.Modified,
			&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
		},
		{
			watch.Deleted,
			&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
		},
		{
			watch.Bookmark,
			&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
		},
	}
	for i, testCase := range testCases {
		buf := &bytes.Buffer{}

		encoder := restclientwatch.NewEncoder(streaming.NewEncoder(buf, getEncoder()), getEncoder())
		if err := encoder.Encode(&watch.Event{Type: testCase.Type, Object: testCase.Object}); err != nil {
			fmt.Printf("%d: unexpected error: %v", i, err)
			continue
		}

		rc := ioutil.NopCloser(buf)
		decoder := restclientwatch.NewDecoder(streaming.NewDecoder(rc, getDecoder()), getDecoder())
		event, obj, err := decoder.Decode()
		if err != nil {
			fmt.Printf("%d: unexpected error: %v", i, err)
			continue
		}
		if !apiequality.Semantic.DeepDerivative(testCase.Object, obj) {
			fmt.Printf("%d: expected %#v, got %#v", i, testCase.Object, obj)
		}
		if event != testCase.Type {
			fmt.Printf("%d: unexpected type: %#v", i, event)
		}
	}
}

func TestWatch() {
	var table = []struct {
		t   watch.EventType
		obj runtime.Object
	}{
		{watch.Added, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "first"}}},
		{watch.Modified, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "second"}}},
		{watch.Deleted, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "last"}}},
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			panic("need flusher!")
		}

		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		//scheme := runtime.NewScheme()

		encoder := restclientwatch.NewEncoder(streaming.NewEncoder(w, getEncoder()), getEncoder())
		for _, item := range table {
			if err := encoder.Encode(&watch.Event{Type: item.t, Object: item.obj}); err != nil {
				panic(err)
			}
			flusher.Flush()
		}
	}))
	defer testServer.Close()


	s := testRESTClient(testServer)
	watching, err := s.Get().Prefix("path/to/watch/thing").Watch(context.Background())
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
	}

	for _, item := range table {
		got, ok := <-watching.ResultChan()
		if !ok {
			fmt.Printf("Unexpected early close")
		}
		if e, a := item.t, got.Type; e != a {
			fmt.Printf("Expected %v, got %v", e, a)
		}
		if e, a := item.obj, got.Object; !apiequality.Semantic.DeepDerivative(e, a) {
			fmt.Printf("Expected %v, got %v", e, a)
		}
	}

	_, ok := <-watching.ResultChan()
	if ok {
		fmt.Printf("Unexpected non-close")
	}
}

func TestWatchNonDefaultContentType() {
	var table = []struct {
		t   watch.EventType
		obj runtime.Object
	}{
		{watch.Added, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "first"}}},
		{watch.Modified, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "second"}}},
		{watch.Deleted, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "last"}}},
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			panic("need flusher!")
		}

		w.Header().Set("Transfer-Encoding", "chunked")
		// manually set the content type here so we get the renegotiation behavior
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		encoder := restclientwatch.NewEncoder(streaming.NewEncoder(w, scheme.Codecs.LegacyCodec(v1.SchemeGroupVersion)), scheme.Codecs.LegacyCodec(v1.SchemeGroupVersion))
		for _, item := range table {
			if err := encoder.Encode(&watch.Event{Type: item.t, Object: item.obj}); err != nil {
				panic(err)
			}
			flusher.Flush()
		}
	}))
	defer testServer.Close()

	// set the default content type to protobuf so that we test falling back to JSON serialization
	contentConfig := defaultContentConfig()
	contentConfig.ContentType = "application/vnd.kubernetes.protobuf"
	s := testRESTClientWithConfig(testServer, contentConfig)
	watching, err := s.Get().Prefix("path/to/watch/thing").Watch(context.Background())
	if err != nil {
		fmt.Printf("Unexpected error")
	}

	for _, item := range table {
		got, ok := <-watching.ResultChan()
		if !ok {
			fmt.Printf("Unexpected early close")
		}
		if e, a := item.t, got.Type; e != a {
			fmt.Printf("Expected %v, got %v", e, a)
		}
		if e, a := item.obj, got.Object; !apiequality.Semantic.DeepDerivative(e, a) {
			fmt.Printf("Expected %v, got %v", e, a)
		}
	}

	_, ok := <-watching.ResultChan()

	if ok {
		fmt.Printf("Unexpected non-close")
	}
}

func TestWatchUnknownContentType() {
	var table = []struct {
		t   watch.EventType
		obj runtime.Object
	}{
		{watch.Added, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "first"}}},
		{watch.Modified, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "second"}}},
		{watch.Deleted, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "last"}}},
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			panic("need flusher!")
		}

		w.Header().Set("Transfer-Encoding", "chunked")
		// manually set the content type here so we get the renegotiation behavior
		w.Header().Set("Content-Type", "foobar")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		encoder := restclientwatch.NewEncoder(streaming.NewEncoder(w, scheme.Codecs.LegacyCodec(v1.SchemeGroupVersion)), scheme.Codecs.LegacyCodec(v1.SchemeGroupVersion))
		for _, item := range table {
			if err := encoder.Encode(&watch.Event{Type: item.t, Object: item.obj}); err != nil {
				panic(err)
			}
			flusher.Flush()
		}
	}))
	defer testServer.Close()

	s := testRESTClient(testServer)
	_, err := s.Get().Prefix("path/to/watch/thing").Watch(context.Background())
	if err == nil {
		fmt.Printf("Expected to fail due to lack of known stream serialization for content type")
	}
}


func testRESTClientWithConfig(srv *httptest.Server, contentConfig rest.ClientContentConfig) *rest.RESTClient {
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

func testRESTClient(srv *httptest.Server) *rest.RESTClient {
	contentConfig := defaultContentConfig()
	return testRESTClientWithConfig(srv, contentConfig)
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
	_ = clientgoscheme.AddToScheme(Scheme)

	return rest.ClientContentConfig{
		ContentType:  "application/json",
		GroupVersion: gvCopy,
		Negotiator:   runtime.NewClientNegotiator(Codecs.WithoutConversion(), gvCopy),
	}
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
	_ = clientgoscheme.AddToScheme(scheme)


	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/api"
	config.ContentType = runtime.ContentTypeJSON

	config.NegotiatedSerializer = serializer.NewCodecFactory(scheme).WithoutConversion()
	return config

}

func get127Config() (rest.Config) {
    config:= rest.Config{}

	scheme := runtime.NewScheme()

	gv := v1.SchemeGroupVersion

	config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(11, 12)
	config.GroupVersion = &gv
	config.APIPath = "/api"
	config.ContentType = runtime.ContentTypeJSON
	config.Host = "127.0.0.1"
	_ = clientgoscheme.AddToScheme(scheme)

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


