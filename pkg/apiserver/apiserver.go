package apiserver

import (
	"sample-extension-apiserver/apis/somethingcontroller/install"
	//api "sample-extension-apiserver/apis/somethingcontroller/v1alpha1"
	//"sample-extension-apiserver/pkg/operator"
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/version"
	genericapiserver "k8s.io/apiserver/pkg/server"
	//restclient "k8s.io/client-go/rest"
	"k8s.io/apiserver/pkg/registry/rest"
	"github.com/golang/glog"

	//something_registry "sample-extension-apiserver/pkg/registry"
	//something_storage "sample-extension-apiserver/pkg/registry/something"
	restclient "k8s.io/client-go/rest"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"sample-extension-apiserver/pkg/registry/admissionreview"
	"fmt"
	"k8s.io/apimachinery/pkg/apimachinery"
	"strings"
)

var (
	groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
	registry             = registered.NewOrDie("")
	Scheme               = runtime.NewScheme()
	Codecs               = serializer.NewCodecFactory(Scheme)
)

type AdmissionHook interface {
	// Initialize is called as a post-start hook
	Initialize(kubeClientConfig *restclient.Config, stopCh <-chan struct{}) error
}

type ValidatingAdmissionHook interface {
	AdmissionHook
	ValidatingResource() (plural schema.GroupVersionResource, singular string)
	Validate(admissionSpec *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse
}

type MutatingAdmissionHook interface {
	AdmissionHook
	MutatingResource() (plural schema.GroupVersionResource, singular string)
	Admit(admissionSpec *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse
}

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
	//OperatorConfig operator.OperatorConfig
	ExtraConfig    ExtraConfig
}

type ExtraConfig struct {
	AdmissionHooks []AdmissionHook
	//ClientConfig   *restclient.Config
}

// ExtServer contains state for a Kubernetes cluster master/api server.
type ExtensionServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
	//Operator         *operator.Operator
}

func (op *ExtensionServer) Run(stopCh <-chan struct{}) error {
	//go op.Operator.Run(stopCh)
	return op.GenericAPIServer.PrepareRun().Run(stopCh)
}

type completedConfig struct {
	GenericConfig  genericapiserver.CompletedConfig
	//OperatorConfig *operator.OperatorConfig
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
		//&c.OperatorConfig,
		&c.ExtraConfig,
	}

	completedCfg.GenericConfig.Version = &version.Info{
		Major: "1",
		Minor: "1", //todo: 0 (may be 0)
	}

	return CompletedConfig{&completedCfg}
}

// New returns a new instance of ExtServer from the given config.
func (c completedConfig) New() (*ExtensionServer, error) {
	glog.Infoln("new completed config..........")
	genericServer, err := c.GenericConfig.New("Extension-api-server", genericapiserver.EmptyDelegate) // completion is done in Complete, no need for a second time
	if err != nil {
		return nil, err
	}
	//operator, err := c.OperatorConfig.New()
	//if err != nil {
	//	return nil, err
	//}

	s := &ExtensionServer{
		GenericAPIServer: genericServer,
		//Operator:         operator,
	}

	//apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(
	//	"somethingcontroller.kube-ac.com",
	//	registry, Scheme, metav1.ParameterCodec, Codecs,
	//)
	//apiGroupInfo.GroupMeta.GroupVersion = api.SchemeGroupVersion
	//
	//v1alpha1storage := map[string]rest.Storage{}
	//v1alpha1storage["somethings"] = something_registry.
	//	RESTInPeace(something_storage.NewREST(Scheme, c.GenericConfig.RESTOptionsGetter))
	//apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage
	//
	//if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
	//	return nil, err
	//}
	//
	//return s, nil

	inClusterConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}

	for _, versionMap := range admissionHooksByGroupThenVersion(c.ExtraConfig.AdmissionHooks...) {
		accessor := meta.NewAccessor()
		versionInterfaces := &meta.VersionInterfaces{
			ObjectConvertor:  Scheme,
			MetadataAccessor: accessor,
		}
		interfacesFor := func(version schema.GroupVersion) (*meta.VersionInterfaces, error) {
			if version != admissionv1beta1.SchemeGroupVersion {
				return nil, fmt.Errorf("unexpected version %v", version)
			}
			return versionInterfaces, nil
		}
		restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{admissionv1beta1.SchemeGroupVersion}, interfacesFor)
		// TODO we're going to need a later k8s.io/apiserver so that we can get discovery to list a different group version for
		// our endpoint which we'll use to back some custom storage which will consume the AdmissionReview type and give back the correct response
		apiGroupInfo := genericapiserver.APIGroupInfo{
			GroupMeta: apimachinery.GroupMeta{
				// filled in later
				//GroupVersion:  admissionVersion,
				//GroupVersions: []schema.GroupVersion{admissionVersion},

				SelfLinker:    runtime.SelfLinker(accessor),
				RESTMapper:    restMapper,
				InterfacesFor: interfacesFor,
				InterfacesByVersion: map[schema.GroupVersion]*meta.VersionInterfaces{
					admissionv1beta1.SchemeGroupVersion: versionInterfaces,
				},
			},
			VersionedResourcesStorageMap: map[string]map[string]rest.Storage{},
			// TODO unhardcode this.  It was hardcoded before, but we need to re-evaluate
			OptionsExternalVersion: &schema.GroupVersion{Version: "v1"},
			Scheme:                 Scheme,
			ParameterCodec:         metav1.ParameterCodec,
			NegotiatedSerializer:   Codecs,
		}

		for _, admissionHooks := range versionMap {
			for i := range admissionHooks {
				admissionHook := admissionHooks[i]
				admissionResource, singularResourceType := admissionHook.Resource()
				admissionVersion := admissionResource.GroupVersion()

				restMapper.AddSpecific(
					admissionv1beta1.SchemeGroupVersion.WithKind("AdmissionReview"),
					admissionResource,
					admissionVersion.WithResource(singularResourceType),
					meta.RESTScopeRoot)

				// just overwrite the groupversion with a random one.  We don't really care or know.
				apiGroupInfo.GroupMeta.GroupVersions = appendUniqueGroupVersion(apiGroupInfo.GroupMeta.GroupVersions, admissionVersion)

				admissionReview := admissionreview.NewREST(admissionHook.Admission)
				v1alpha1storage := map[string]rest.Storage{
					admissionResource.Resource: admissionReview,
				}
				apiGroupInfo.VersionedResourcesStorageMap[admissionVersion.Version] = v1alpha1storage
			}
		}

		// just prefer the first one in the list for consistency
		apiGroupInfo.GroupMeta.GroupVersion = apiGroupInfo.GroupMeta.GroupVersions[0]

		if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
			return nil, err
		}
	}

	for _, hook := range c.ExtraConfig.AdmissionHooks {
		postStartName := postStartHookName(hook)
		if len(postStartName) == 0 {
			continue
		}
		s.GenericAPIServer.AddPostStartHookOrDie(postStartName,
			func(context genericapiserver.PostStartHookContext) error {
				return hook.Initialize(inClusterConfig, context.StopCh)
			},
		)
	}

	return s, nil
}

func appendUniqueGroupVersion(slice []schema.GroupVersion, elems ...schema.GroupVersion) []schema.GroupVersion {
	m := map[schema.GroupVersion]bool{}
	for _, gv := range slice {
		m[gv] = true
	}
	for _, e := range elems {
		m[e] = true
	}
	out := make([]schema.GroupVersion, 0, len(m))
	for gv := range m {
		out = append(out, gv)
	}
	return out
}

func postStartHookName(hook AdmissionHook) string {
	var ns []string
	if mutatingHook, ok := hook.(MutatingAdmissionHook); ok {
		gvr, _ := mutatingHook.MutatingResource()
		ns = append(ns, fmt.Sprintf("mutating-%s.%s.%s", gvr.Resource, gvr.Version, gvr.Group))
	}
	if validatingHook, ok := hook.(ValidatingAdmissionHook); ok {
		gvr, _ := validatingHook.ValidatingResource()
		ns = append(ns, fmt.Sprintf("validating-%s.%s.%s", gvr.Resource, gvr.Version, gvr.Group))
	}
	if len(ns) == 0 {
		return ""
	}
	return strings.Join(append(ns, "init"), "-")
}

func admissionHooksByGroupThenVersion(admissionHooks ...AdmissionHook) map[string]map[string][]admissionHookWrapper {
	ret := map[string]map[string][]admissionHookWrapper{}

	for i := range admissionHooks {
		if mutatingHook, ok := admissionHooks[i].(MutatingAdmissionHook); ok {
			gvr, _ := mutatingHook.MutatingResource()
			group, ok := ret[gvr.Group]
			if !ok {
				group = map[string][]admissionHookWrapper{}
				ret[gvr.Group] = group
			}
			group[gvr.Version] = append(group[gvr.Version], mutatingAdmissionHookWrapper{mutatingHook})
		}
		if validatingHook, ok := admissionHooks[i].(ValidatingAdmissionHook); ok {
			gvr, _ := validatingHook.ValidatingResource()
			group, ok := ret[gvr.Group]
			if !ok {
				group = map[string][]admissionHookWrapper{}
				ret[gvr.Group] = group
			}
			group[gvr.Version] = append(group[gvr.Version], validatingAdmissionHookWrapper{validatingHook})
		}
	}

	return ret
}

// admissionHookWrapper wraps either a validating or mutating admission hooks, calling the respective resource and admission method.
type admissionHookWrapper interface {
	Resource() (plural schema.GroupVersionResource, singular string)
	Admission(admissionSpec *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse
}

type mutatingAdmissionHookWrapper struct {
	hook MutatingAdmissionHook
}

func (h mutatingAdmissionHookWrapper) Resource() (plural schema.GroupVersionResource, singular string) {
	return h.hook.MutatingResource()
}

func (h mutatingAdmissionHookWrapper) Admission(admissionSpec *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse {
	return h.hook.Admit(admissionSpec)
}

type validatingAdmissionHookWrapper struct {
	hook ValidatingAdmissionHook
}

func (h validatingAdmissionHookWrapper) Resource() (plural schema.GroupVersionResource, singular string) {
	return h.hook.ValidatingResource()
}

func (h validatingAdmissionHookWrapper) Admission(admissionSpec *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse {
	return h.hook.Validate(admissionSpec)
}
