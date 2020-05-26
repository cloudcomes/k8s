
package main

import (
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	diskcached "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	mapper "k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"time"
)


func main() {

	//TestRESTMapperRESTMappingSelectsVersion()
	RESTMapperK2R()
}

func RESTMapperK2R() {

	discoveryclient, _ := ToDiscoveryClient()
	m := mapper.NewDeferredDiscoveryRESTMapper(discoveryclient)

	preferredVersions := []string{}

	gk := schema.GroupKind{Group: "", Kind: "Pod"}
	gv := schema.GroupVersion{Group: "", Version: "v1"}

	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	//gvk := schema.GroupVersionKind{Group: "app", Version: "v1", Kind: "Deployment"}

	gvk1, _ := m.KindFor(gvr)
	fmt.Println("gvk1 is: %v", gvk1)

    gvr1,_ := m.ResourceFor(gvr)
	fmt.Println("gvr1 is: %v", gvr1)


	preferredVersions = append(preferredVersions, gv.Version)
	mapping, _ := m.RESTMapping(gk, preferredVersions...)


	fmt.Println("mapping is: %v", mapping)

}

func TestRESTMapperRESTMappingSelectsVersion() {
	expectedGroupVersion1 := schema.GroupVersion{Group: "tgroup", Version: "test1"}
	expectedGroupVersion2 := schema.GroupVersion{Group: "tgroup", Version: "test2"}
	expectedGroupVersion3 := schema.GroupVersion{Group: "tgroup", Version: "test3"}

	otherObjectGK := schema.GroupKind{Group: "tgroup", Kind: "OtherObject"}

	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{expectedGroupVersion1, expectedGroupVersion2})
	mapper.Add(expectedGroupVersion1.WithKind("InternalObject"), meta.RESTScopeNamespace)
	mapper.Add(expectedGroupVersion2.WithKind("OtherObject"), meta.RESTScopeNamespace)

	// pick default matching object kind based on search order
	mapping, err := mapper.RESTMapping(otherObjectGK)
	if err != nil {
		fmt.Print("unexpected error: %v", err)
	}
	if mapping.Resource != expectedGroupVersion2.WithResource("otherobjects") || mapping.GroupVersionKind.GroupVersion() != expectedGroupVersion2 {
		fmt.Print("unexpected mapping: %#v", mapping)
	}

	mapping, err = mapper.RESTMapping(otherObjectGK, expectedGroupVersion3.Version, expectedGroupVersion2.Version)
	if err != nil {
		fmt.Print("unexpected error: %v", err)
	}
	if mapping.Resource != expectedGroupVersion2.WithResource("otherobjects") || mapping.GroupVersionKind.GroupVersion() != expectedGroupVersion2 {
		fmt.Print("unexpected mapping: %#v", mapping)
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
