

package main

import (
	"flag"
	"fmt"
	v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	//"os"
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"

	"github.com/davecgh/go-spew/spew"
	"time"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func main() {
//testGet()
TestWatch()
//	testDiscovery()
	//TestGetOpenAPISchema()
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
	// /<namespace>/<resource>/<name>
	// GET https://apiserver/api/v1/namespaces/perf/pods/netperf-655c567cf-fmw6v
	result := restclient.Get().
		Namespace("default").
		Resource("pods").
		Name("gohttpdemo-5cb6954c5c-drrmg").
		Do(context.Background())
	bytes, err := result.Raw()
	if err != nil {
		fmt.Printf("%s: %sn", err.Error(), bytes)
	} else {
		fmt.Printf("%sn", bytes)
	}
}

func TestWatch() {

	// RESTClientFor returns a RESTClient that satisfies the requested attributes on a client Config
	// object. Note that a RESTClient may require fields that are optional when initializing a Client.
	// A RESTClient created by this method is generic - it expects to operate on an API that follows
	// the Kubernetes conventions, but may not be the Kubernetes API.
	restclient, err := rest.RESTClientFor(getConfig())
	if err != nil {
		log.Fatalln(err)
	}
	// /<namespace>/<resource>/<name>
	// GET https://apiserver/api/v1/namespaces/perf/pods/netperf-655c567cf-fmw6v
	//r := restclient.Get().Prefix("watch/namespaces/default/pods/gohttpdemo-5cb6954c5c-drrmg")
	r := restclient.Get().Prefix("watch/namespaces/default/pods")
	s := r.URL().String();
	result, _ := r.Watch(context.Background())

	fmt.Printf("namespace should be in order in path: %s", s)

	bytes, _ := <-result.ResultChan()
	fmt.Printf("got %v", bytes)

}



func testDiscovery(){

	// NewDiscoveryClientForConfigOrDie creates a new DiscoveryClient for the given config. If
	// there is an error, it panics.
	discover := discovery.NewDiscoveryClientForConfigOrDie(getConfig())


	start := time.Now()
	//resources, err := discover.ServerPreferredNamespacedResources()
	//resources, err := discover.ServerPreferredResources()
	serverResources, err := discover.ServerResources()
	if err != nil {
		fmt.Printf("unexpected error: %v", err)
	}
	end := time.Now()
	if d := end.Sub(start); d > time.Second {
		fmt.Printf("took too long to perform discovery: %s", d)
	}

	fmt.Printf("%+v\n\n", serverResources)

	serverGroupVersions := groupVersions(serverResources)

	fmt.Printf("%vn", serverGroupVersions)

}


func TestGetOpenAPISchema() {

	discover := discovery.NewDiscoveryClientForConfigOrDie(getConfig())
	spew.Dump(discover)
	got, err := discover.OpenAPISchema()
	if err != nil {
		fmt.Printf("unexpected error getting openapi: %v", err)
	}

	//fmt.Printf("%+v\n\n", got)

	//_ = json.NewEncoder(os.Stdout).Encode(&got)
	js ,err := json.Marshal(&got)
	fmt.Println("Serialized:  ", string(js), err)

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

	//unable to decode watch event: no kind \"Pod\" is registered for version \"v1\"
	//we need mannually add the scheme: apis/apiextensions/v1beta1 like this:
	//clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	//_ = clientgoscheme.AddToScheme(scheme)

	_ = clientgoscheme.AddToScheme(scheme)

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


