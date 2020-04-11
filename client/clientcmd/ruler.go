package main

import (

	"github.com/davecgh/go-spew/spew"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	clientCfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	spew.Dump(clientCfg, err)


}
//把config文件读进来，生产Config对象
/* clientCfg 的内容：
kubernetes-1.18.0-beta.1/staging/src/k8s.io/client-go/tools/clientcmd/api/types.go
// Config holds the information needed to build connect to remote kubernetes clusters as a given user
//type Config struct {
	// Legacy field from pkg/api/types.go TypeMeta.
	// TODO(jlowdermilk): remove this after eliminating downstream dependencies.
	// +k8s:conversion-gen=false
	// +optional
	Kind string `json:"kind,omitempty"`
	// Legacy field from pkg/api/types.go TypeMeta.
	// TODO(jlowdermilk): remove this after eliminating downstream dependencies.
	// +k8s:conversion-gen=false
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Preferences holds general information to be use for cli interactions
	Preferences Preferences `json:"preferences"`

	// Clusters is a map of referencable names to cluster configs
    // Cluster contains information about how to communicate with a kubernetes cluster
	Clusters map[string]*Cluster `json:"clusters"`

	// AuthInfos is a map of referencable names to user configs
    // AuthInfo contains information that describes identity information.  This is use to tell the kubernetes cluster who you are.
	AuthInfos map[string]*AuthInfo `json:"users"`

	// Contexts is a map of referencable names to context configs
    // Context is a tuple of references to a cluster (how do I communicate with a kubernetes cluster), a user (how do I identify myself),
    //and a namespace (what subset of resources do I want to work with)
	Contexts map[string]*Context `json:"contexts"`

	// CurrentContext is the name of the context that you would like to use by default
	CurrentContext string `json:"current-context"`

	// Extensions holds additional information. This is useful for extenders so that reads and writes don't clobber unknown fields
	// +optional
	Extensions map[string]runtime.Object `json:"extensions,omitempty"`
}
*/

/*
(*api.Config)(0xc0001ce120)({
 Kind: (string) "",
 APIVersion: (string) "",
 Preferences: (api.Preferences) {
  Colors: (bool) false,
  Extensions: (map[string]runtime.Object) {
  }
 },
 Clusters: (map[string]*api.Cluster) (len=2) {
  (string) (len=15) "externalCluster": (*api.Cluster)(0xc0000bfb60)({
   LocationOfOrigin: (string) (len=25) "/Users/cloud/.kube/config",
   Server: (string) (len=26) "https://218.78.37.164:5443",
   InsecureSkipTLSVerify: (bool) true,
   CertificateAuthority: (string) "",
   CertificateAuthorityData: ([]uint8) <nil>,
   Extensions: (map[string]runtime.Object) {
   }
  }),
  (string) (len=15) "internalCluster": (*api.Cluster)(0xc0000bfaa0)({
   LocationOfOrigin: (string) (len=25) "/Users/cloud/.kube/config",
   Server: (string) (len=24) "https://192.168.0.3:5443",
   InsecureSkipTLSVerify: (bool) false,
   CertificateAuthority: (string) "",
   CertificateAuthorityData: ([]uint8) (len=1046 cap=1046) {
    00000000  2d 2d 2d 2d 2d 42 45 47  49 4e 20 43 45 52 54 49  |-----BEGIN CERTI|
    00000010  46 49 43 41 54 45 2d 2d  2d 2d 2d 0a 4d 49 49 43  |FICATE-----.MIIC|
    00000020  31 6a 43 43 41 62 36 67  41 77 49 42 41 67 49 42  |1jCCAb6gAwIBAgIB|
    00000030  41 44 41 4e 42 67 6b 71  68 6b 69 47 39 77 30 42  |ADANBgkqhkiG9w0B|
    00000040  41 51 73 46 41 44 41 62  4d 52 6b 77 46 77 59 44  |AQsFADAbMRkwFwYD|
    000003e0  46 38 4d 4c 45 62 35 49  35 56 0a 4c 45 6e 42 55  |F8MLEb5I5V.LEnBU|
    000003f0  42 35 44 4f 6c 79 63 38  51 3d 3d 0a 2d 2d 2d 2d  |B5DOlyc8Q==.----|
    00000400  2d 45 4e 44 20 43 45 52  54 49 46 49 43 41 54 45  |-END CERTIFICATE|
    00000410  2d 2d 2d 2d 2d 0a                                 |-----.|
   },
   Extensions: (map[string]runtime.Object) {
   }
  })
 },
 AuthInfos: (map[string]*api.AuthInfo) (len=1) {
  (string) (len=4) "user": (*api.AuthInfo)(0xc0001c6000)({
   LocationOfOrigin: (string) (len=25) "/Users/cloud/.kube/config",
   ClientCertificate: (string) "",
   ClientCertificateData: ([]uint8) (len=1440 cap=1440) {
    00000000  2d 2d 2d 2d 2d 42 45 47  49 4e 20 43 45 52 54 49  |-----BEGIN CERTI|
    00000010  46 49 43 41 54 45 2d 2d  2d 2d 2d 0a 4d 49 49 44  |FICATE-----.MIID|
    00000020  2b 7a 43 43 41 75 4f 67  41 77 49 42 41 67 49 49  |+zCCAuOgAwIBAgII|
    00000030  54 46 4c 48 36 31 7a 44  37 34 6f 77 44 51 59 4a  |TFLH61zD74owDQYJ|
    00000040  4b 6f 5a 49 68 76 63 4e  41 51 45 4c 42 51 41 77  |KoZIhvcNAQELBQAw|
    00000580  6f 64 2b 76 34 0a 2d 2d  2d 2d 2d 45 4e 44 20 43  |od+v4.-----END C|
    00000590  45 52 54 49 46 49 43 41  54 45 2d 2d 2d 2d 2d 0a  |ERTIFICATE-----.|
   },
   ClientKey: (string) "",
   ClientKeyData: ([]uint8) (len=1675 cap=1675) {
    00000000  2d 2d 2d 2d 2d 42 45 47  49 4e 20 52 53 41 20 50  |-----BEGIN RSA P|
    00000010  52 49 56 41 54 45 20 4b  45 59 2d 2d 2d 2d 2d 0a  |RIVATE KEY-----.|
    00000020  4d 49 49 45 6f 67 49 42  41 41 4b 43 41 51 45 41  |MIIEogIBAAKCAQEA|
    00000030  6f 4a 4d 39 70 4f 32 5a  37 32 68 39 37 73 4a 4d  |oJM9pO2Z72h97sJM|
    00000040  52 6c 44 41 2b 72 46 53  49 53 2b 72 70 4b 52 73  |RlDA+rFSIS+rpKRs|
    000000b0  38 39 45 4d 48 78 5a 6f  67 6b 75 61 38 4d 59 68  |89EMHxZogkua8MYh|
    00000670  2d 2d 45 4e 44 20 52 53  41 20 50 52 49 56 41 54  |--END RSA PRIVAT|
    00000680  45 20 4b 45 59 2d 2d 2d  2d 2d 0a                 |E KEY-----.|
   },
   Token: (string) "",
   TokenFile: (string) "",
   Impersonate: (string) "",
   ImpersonateGroups: ([]string) <nil>,
   ImpersonateUserExtra: (map[string][]string) <nil>,
   Username: (string) "",
   Password: (string) "",
   AuthProvider: (*api.AuthProviderConfig)(<nil>),
   Exec: (*api.ExecConfig)(<nil>),
   Extensions: (map[string]runtime.Object) {
   }
  })
 },
 Contexts: (map[string]*api.Context) (len=2) {
  (string) (len=8) "internal": (*api.Context)(0xc0000cd720)({
   LocationOfOrigin: (string) (len=25) "/Users/cloud/.kube/config",
   Cluster: (string) (len=15) "internalCluster",
   AuthInfo: (string) (len=4) "user",
   Namespace: (string) "",
   Extensions: (map[string]runtime.Object) {
   }
  }),
  (string) (len=8) "external": (*api.Context)(0xc0000cd860)({
   LocationOfOrigin: (string) (len=25) "/Users/cloud/.kube/config",
   Cluster: (string) (len=15) "externalCluster",
   AuthInfo: (string) (len=4) "user",
   Namespace: (string) "",
   Extensions: (map[string]runtime.Object) {
   }
  })
 },
 CurrentContext: (string) (len=8) "external",
 Extensions: (map[string]runtime.Object) {
 }
})

 */
