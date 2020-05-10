
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"regexp"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/client-go/discovery"
	dc "k8s.io/client-go/discovery"
	diskcached "k8s.io/client-go/discovery/cached/disk"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)



func main() {

//	TestDiscovery()
//  TestDiscoverydemo()
//	TestGetOpenAPISchema()
	RunAPIResources()

}



// ServerResourcesForGroupVersion returns the supported resources for a group and version.
func  TestDiscoverydemo()  {

	dcClient := dc.NewDiscoveryClientForConfigOrDie(getConfig())
	start := time.Now()
	serverResources, err := dcClient.ServerResources()
	//fmt.Printf("%v %+v %#v\n", serverResources, serverResources, serverResources)
	fmt.Printf("%+v\n", serverResources)
	if err != nil {
		fmt.Println("unexpected error: %v", err)
	}
	end := time.Now()
	if d := end.Sub(start); d > time.Second {
		fmt.Println("took too long to perform discovery: %s", d)
	}
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

// RunAPIResources does the work
func RunAPIResources()  {

    Cached := true

	discoveryclient, err := ToDiscoveryClient()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

	if !Cached {
		// Always request fresh data from the server
		discoveryclient.Invalidate()
	}

	errs := []error{}
	lists, err := discoveryclient.ServerPreferredResources()
	if err != nil {
		errs = append(errs, err)
	}

	for _, list := range lists {
		if len(list.APIResources) == 0 {
			continue
		}
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		fmt.Printf("%+v\n", gv)
		if err != nil {
			continue
		}
	}


}


// ToDiscoveryClient implements RESTClientGetter.
// Expects the AddFlags method to have been called.
// Returns a CachedDiscoveryInterface using a computed RESTConfig.
func ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	//config, err := f.ToRESTConfig()
	config := getConfig()

	// The more groups you have, the more discovery requests you need to make.
	// given 25 groups (our groups + a few custom resources) with one-ish version each, discovery needs to make 50 requests
	// double it just so we don't end up here again for a while.  This config is only used for discovery.
	config.Burst = 100

	// retrieve a user-provided value for the "cache-dir"
	// defaulting to ~/.kube/http-cache if no user-value is given.
	httpCacheDir := "./http-cache"

	discoveryCacheDir := computeDiscoverCacheDir(filepath.Join(homedir.HomeDir(), ".kube", "cache", "discovery"), config.Host)
	return diskcached.NewCachedDiscoveryClientForConfig(config, discoveryCacheDir, httpCacheDir, time.Duration(10*time.Minute))
}


// overlyCautiousIllegalFileCharacters matches characters that *might* not be supported.  Windows is really restrictive, so this is really restrictive
var overlyCautiousIllegalFileCharacters = regexp.MustCompile(`[^(\w/\.)]`)


// computeDiscoverCacheDir takes the parentDir and the host and comes up with a "usually non-colliding" name.
func computeDiscoverCacheDir(parentDir, host string) string {
	// strip the optional scheme from host if its there:
	schemelessHost := strings.Replace(strings.Replace(host, "https://", "", 1), "http://", "", 1)
	// now do a simple collapse of non-AZ09 characters.  Collisions are possible but unlikely.  Even if we do collide the problem is short lived
	safeHost := overlyCautiousIllegalFileCharacters.ReplaceAllString(schemelessHost, "_")
	return filepath.Join(parentDir, safeHost)
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


