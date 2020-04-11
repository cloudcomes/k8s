package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	pkscorev1 "git.pacloud.io/pks/helm-operator/pkg/apis/core/v1"
	pksmetav1 "git.pacloud.io/pks/helm-operator/pkg/apis/meta/v1"
	pksv1 "git.pacloud.io/pks/helm-operator/pkg/apis/pks/v1"
	"github.com/go-openapi/spec"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	openapinamer "k8s.io/apiserver/pkg/endpoints/openapi"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/klog"
	"k8s.io/kube-openapi/pkg/builder"
	"k8s.io/kube-openapi/pkg/common"
	"net"
	"os"
)

type ResourceInfo struct {
	gvk             schema.GroupVersionKind
	obj             runtime.Object
	list            runtime.Object
	namespaceScoped bool
}

type TypeInfo struct {
	GroupVersion    schema.GroupVersion
	Resource        string
	Kind            string
	NamespaceScoped bool
}
type Config struct {
	Scheme             *runtime.Scheme
	Codecs             serializer.CodecFactory
	Info               spec.InfoProps
	OpenAPIDefinitions []common.GetOpenAPIDefinitions
	Resources          []TypeInfo
}

// 这个Storage是资源CRUD操作的提供者
type StandardStorage struct {
	cfg ResourceInfo
}

// 强制它实现以下接口
var _ rest.GroupVersionKindProvider = &StandardStorage{}
var _ rest.Scoper = &StandardStorage{}
//var _ rest.StandardStorage = &StandardStorage{}

// GroupVersionKindProvider
func (r *StandardStorage) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return r.cfg.gvk
}

// Scoper
func (r *StandardStorage) NamespaceScoped() bool {
	return r.cfg.namespaceScoped
}

// Getter
func (r *StandardStorage) New() runtime.Object {
	return r.cfg.obj
}

func (r *StandardStorage) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	return r.New(), nil
}

func (r *StandardStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.New(), nil
}

// Lister
func (r *StandardStorage) NewList() runtime.Object {
	return r.cfg.list
}

func (r *StandardStorage) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	return r.NewList(), nil
}

// CreaterUpdater
func (r *StandardStorage) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	return r.New(), true, nil
}

// GracefulDeleter
func (r *StandardStorage) Delete(ctx context.Context, name string, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	return r.New(), true, nil
}

// CollectionDeleter
func (r *StandardStorage) DeleteCollection(ctx context.Context, options *metav1.DeleteOptions, listOptions *metainternalversion.ListOptions) (runtime.Object, error) {
	return r.NewList(), nil
}

// Watcher
func (r *StandardStorage) Watch(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	return nil, nil
}



func (c *Config) GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	out := map[string]common.OpenAPIDefinition{}
	for _, def := range c.OpenAPIDefinitions {
		for k, v := range def(ref) {
			out[k] = v
		}
	}
	return out
}
func RenderOpenAPISpec(cfg Config) (string, error) {

	// 总是需要向Scheme注册corev1、metav1
	metav1.AddToGroupVersion(cfg.Scheme, schema.GroupVersion{Version: "v1"})
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	cfg.Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)

	// API Server选项
	options := genericoptions.NewRecommendedOptions("/registry/pks.yun.pingan.com", cfg.Codecs.LegacyCodec(), &genericoptions.ProcessInfo{})
	options.SecureServing.BindPort = 6445
	options.Etcd = nil
	options.Authentication = nil
	options.Authorization = nil
	options.CoreAPI = nil
	options.Admission = nil
	// 自动生成的服务器证书的存放目录
	options.SecureServing.ServerCert.CertDirectory = "/tmp/helm-operator"

	// 启动的API Server，监听的地址
	publicAddr := "localhost"
	ips := []net.IP{net.ParseIP("127.0.0.1")}

	// 尝试自动生成服务器证书
	if err := options.SecureServing.MaybeDefaultWithSelfSignedCerts(publicAddr, nil, ips); err != nil {
		klog.Fatal(err)
	}

	// API Server配置
	serverConfig := genericapiserver.NewRecommendedConfig(cfg.Codecs)
	if err := options.ApplyTo(serverConfig, cfg.Scheme); err != nil {
		return "", err
	}
	serverConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(cfg.GetOpenAPIDefinitions, openapinamer.NewDefinitionNamer(cfg.Scheme))
	serverConfig.OpenAPIConfig.Info.InfoProps = cfg.Info
	//                             完成配置，  创建服务器
	genericServer, err := serverConfig.Complete().New("helm-operator-server", genericapiserver.NewEmptyDelegate())
	if err != nil {
		return "", err
	}

	// 这里处理本服务器需要Serve的资源列表
	table := map[schema.GroupVersion]map[string]rest.Storage{}
	{
		for _, ti := range cfg.Resources {
			// 对于每一种资源，根据其GV寻找存储库
			var resmap map[string]rest.Storage
			if m, found := table[ti.GroupVersion]; found {
				resmap = m
			} else {
				// 如果找不到，则创建存储库（每个GV一个存储库）
				resmap = map[string]rest.Storage{}
				table[ti.GroupVersion] = resmap
			}

			gvk := ti.GroupVersion.WithKind(ti.Kind)
			// 创建这种资源的一个对象
			obj, err := cfg.Scheme.New(gvk)
			if err != nil {
				return "", err
			}
			// 创建这种资源的列表对象
			list, err := cfg.Scheme.New(ti.GroupVersion.WithKind(ti.Kind + "List"))
			if err != nil {
				return "", err
			}

			// 为资源创建存储，并Put到它的GV的存储库中
			resmap[ti.Resource] = &StandardStorage{ResourceInfo{
				// GVK信息
				gvk: gvk,
				// 资源和资源列表的原型
				obj:  obj,
				list: list,
				// 提示此资源是否命名空间化
				namespaceScoped: ti.NamespaceScoped,
			}}
		}
	}
	for gv, resmap := range table {
		// 为每个组创建API组信息
		apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(gv.Group, cfg.Scheme, metav1.ParameterCodec, cfg.Codecs)
		storage := map[string]rest.Storage{}
		for r, s := range resmap {
			storage[r] = s
		}
		apiGroupInfo.VersionedResourcesStorageMap[gv.Version] = storage

		// 并安装到此服务器
		if err := genericServer.InstallAPIGroup(&apiGroupInfo); err != nil {
			return "", err
		}
	}
	// 构件API规范，也就是Swagger 2.0 Spec
	spec, err := builder.BuildOpenAPISpec(genericServer.Handler.GoRestfulContainer.RegisteredWebServices(), serverConfig.OpenAPIConfig)
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func main() {
	flag.Parse()
	if len(os.Args) <= 1 {
		panic("version required")
	}
	version := os.Args[1]

	scheme := runtime.NewScheme()
	// 将我们需要处理的类型加入到scheme
	scheme.AddKnownTypeWithName(schema.GroupVersion{Group: "pks.yun.pingan.com", Version: "v1"}.WithKind("Release"), &pksv1.Release{})
	scheme.AddKnownTypeWithName(schema.GroupVersion{Group: "pks.yun.pingan.com", Version: "v1"}.WithKind("ReleaseList"), &pksv1.Release{})
	scheme.AddKnownTypeWithName(schema.GroupVersion{Group: "pks.yun.pingan.com", Version: "v1"}.WithKind("WatchEvent"), &pksv1.Release{})

	spec, err := RenderOpenAPISpec(Config{
		Info: spec.InfoProps{
			Version: version,
			Title:   "Helm Operator OpenAPI",
		},
		Scheme: scheme,
		Codecs: serializer.NewCodecFactory(scheme),
		OpenAPIDefinitions: []common.GetOpenAPIDefinitions{
			pkscorev1.GetOpenAPIDefinitions,
			pksmetav1.GetOpenAPIDefinitions,
			pksv1.GetOpenAPIDefinitions,
		},
		Resources: []TypeInfo{
			{
				GroupVersion:    schema.GroupVersion{Group: "pks.yun.pingan.com", Version: "v1"},
				Kind:            "Release",
				Resource:        "Release",
				NamespaceScoped: true,
			},
		},
	})
	if err != nil {
		klog.Fatal(err.Error())
	}
	fmt.Println(spec)
}
