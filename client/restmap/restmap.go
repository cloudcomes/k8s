
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
	"k8s.io/kubectl/pkg/explain"
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

	preferredVersions = append(preferredVersions, gv.Version)

	mapping, _ := m.RESTMapping(gk, preferredVersions...)

	fmt.Println("mapping is: %v", mapping)

}


// Run executes the appropriate steps to print a model's documentation
func  RESTMapperR2K() error {
	recursive := o.Recursive
	apiVersionString := o.APIVersion

	// TODO: After we figured out the new syntax to separate group and resource, allow
	// the users to use it in explain (kubectl explain <group><syntax><resource>).
	// Refer to issue #16039 for why we do this. Refer to PR #15808 that used "/" syntax.
	inModel, fieldsPath, err := explain.SplitAndParseResourceRequest(args[0], o.Mapper)
	if err != nil {
		return err
	}

	// TODO: We should deduce the group for a resource by discovering the supported resources at server.
	fullySpecifiedGVR, groupResource := schema.ParseResourceArg(inModel)
	gvk := schema.GroupVersionKind{}
	if fullySpecifiedGVR != nil {
		gvk, _ = o.Mapper.KindFor(*fullySpecifiedGVR)
	}
	if gvk.Empty() {
		gvk, err = o.Mapper.KindFor(groupResource.WithVersion(""))
		if err != nil {
			return err
		}
	}

	if len(apiVersionString) != 0 {
		apiVersion, err := schema.ParseGroupVersion(apiVersionString)
		if err != nil {
			return err
		}
		gvk = apiVersion.WithKind(gvk.Kind)
	}

	schema := o.Schema.LookupResource(gvk)
	if schema == nil {
		return fmt.Errorf("Couldn't find resource for %q", gvk)
	}

	return explain.PrintModelDescription(fieldsPath, o.Out, schema, gvk, recursive)
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
