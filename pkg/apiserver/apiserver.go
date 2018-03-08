package apiserver

import (
	"sample-extension-apiserver/apis/somethingcontroller/install"
	api "sample-extension-apiserver/apis/somethingcontroller/v1alpha1"
	"sample-extension-apiserver/pkg/operator"
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/version"
	genericapiserver "k8s.io/apiserver/pkg/server"
	restclient "k8s.io/client-go/rest"
	"k8s.io/apiserver/pkg/registry/rest"
)

var (
	groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
	registry             = registered.NewOrDie("")
	Scheme               = runtime.NewScheme()
	Codecs               = serializer.NewCodecFactory(Scheme)
)

func init() {
	install.Install(groupFactoryRegistry, registry, Scheme)

	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}
type Config struct {
	GenericConfig  *genericapiserver.RecommendedConfig
	OperatorConfig operator.OperatorConfig
	ExtraConfig    ExtraConfig
}

type ExtraConfig struct {
	//AdmissionHooks []hookapi.AdmissionHook
	ClientConfig   *restclient.Config
}

// ExtServer contains state for a Kubernetes cluster master/api server.
type ExtServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
	Operator         *operator.Operator
}

func (op *ExtServer) Run(stopCh <-chan struct{}) error {
	go op.Operator.Run(stopCh)
	return op.GenericAPIServer.PrepareRun().Run(stopCh)
}

type completedConfig struct {
	GenericConfig  genericapiserver.CompletedConfig
	OperatorConfig *operator.OperatorConfig
	ExtraConfig    *ExtraConfig
}

type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *Config) Complete() CompletedConfig {
	completedCfg := completedConfig{
		c.GenericConfig.Complete(),
		&c.OperatorConfig,
		&c.ExtraConfig,
	}

	completedCfg.GenericConfig.Version = &version.Info{
		Major: "1",
		Minor: "1",
	}

	return CompletedConfig{&completedCfg}
}

// New returns a new instance of ExtServer from the given config.
func (c completedConfig) New() (*ExtServer, error) {
	genericServer, err := c.GenericConfig.New("Extension-api-server", genericapiserver.EmptyDelegate) // completion is done in Complete, no need for a second time
	if err != nil {
		return nil, err
	}
	operator, err := c.OperatorConfig.New()
	if err != nil {
		return nil, err
	}

	s := &ExtServer{
		GenericAPIServer: genericServer,
		Operator:         operator,
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(
		"somethingcontroller.kube-ac.com",
		registry, Scheme, metav1.ParameterCodec, Codecs,
	)
	apiGroupInfo.GroupMeta.GroupVersion = api.SchemeGroupVersion

	v1alpha1storage := map[string]rest.Storage{}
	v1alpha1storage["somethings"] = NewREST()
	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}
