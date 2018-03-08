package server

import (
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"sample-extension-apiserver/pkg/apiserver"
	"io"
	api "sample-extension-apiserver/apis/somethingcontroller/v1alpha1"
	"fmt"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"net"
	//"sample-extension-apiserver/pkg/operator"
	//"time"
	//"github.com/appscode/go/runtime"
	//
	//"k8s.io/client-go/kubernetes"
	//"github.com/appscode/kutil/meta"
	//cs "sample-extension-apiserver/client/clientset/versioned"
	//kext_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"time"
)

const defaultEtcdPathPrefix = "/registry/somethingcontroller.kube-ac.com"


type OperatorOptions struct {
	ConfigPath          string
	OpsAddress          string
	ControllerNamespace string

	QPS          float32
	Burst        int
	ResyncPeriod time.Duration
}

func NewOperatorOptions() *OperatorOptions {
	return &OperatorOptions{
		ConfigPath: "/home/ac/go/src/sample-extension-apiserver/hack/deploy/config.yaml",
		OpsAddress: ":56790",
		// ref: https://github.com/kubernetes/ingress-nginx/blob/e4d53786e771cc6bdd55f180674b79f5b692e552/pkg/ingress/controller/launch.go#L252-L259
		// High enough QPS to fit all expected use cases. QPS=0 is not set here, because client code is overriding it.
		QPS: 1e6,
		// High enough Burst to fit all expected use cases. Burst=0 is not set here, because client code is overriding it.
		Burst:              1e6,
		ResyncPeriod:       10 * time.Minute,
	}
}

//func (s *OperatorOptions) ApplyTo(cfg *operator.OperatorConfig) error {
//	var err error
//
//	cfg.OperatorNamespace = meta.Namespace()
//	cfg.ClientConfig.QPS = s.QPS
//	cfg.ClientConfig.Burst = s.Burst
//	cfg.ResyncPeriod = s.ResyncPeriod
//
//	if cfg.KubeClient, err = kubernetes.NewForConfig(cfg.ClientConfig); err != nil {
//		return err
//	}
//	if cfg.CRDClient, err = kext_cs.NewForConfig(cfg.ClientConfig); err != nil {
//		return err
//	}
//	if cfg.ExtServerClient, err = cs.NewForConfig(cfg.ClientConfig); err != nil {
//		return err
//	}
//
//	cfg.OpsAddress = s.OpsAddress
//	cfg.ConfigPath = s.ConfigPath
//
//	return nil
//}

type ServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	OperatorOptions    *OperatorOptions

	StdOut                io.Writer
	StdErr                io.Writer
}

func NewOptions(out, errOut io.Writer) *ServerOptions {
	opt := &ServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(
			defaultEtcdPathPrefix,
			apiserver.Codecs.LegacyCodec(api.SchemeGroupVersion),
		),
		//OperatorOptions:    NewOperatorOptions(),
		StdOut:             out,
		StdErr:             errOut,
	}
	opt.RecommendedOptions.Etcd = nil
	opt.RecommendedOptions.SecureServing.BindPort = 8443

	return opt
}

func (o ServerOptions) Validate(args []string) error {
	errors := []error{}
	errors = append(errors, o.RecommendedOptions.Validate()...)
	return utilerrors.NewAggregate(errors)
}

func (o *ServerOptions) Complete() error {
	return nil
}

func (o *ServerOptions) Config() (*apiserver.Config, error) {
	// register admission plugins
	//banflunder.Register(o.RecommendedOptions.Admission.Plugins)

	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	//operatorConfig := operator.NewOperatorConfig(serverConfig.ClientConfig)
	//if err := o.OperatorOptions.ApplyTo(operatorConfig); err != nil {
	//	return nil, err
	//}

	config := &apiserver.Config{
		GenericConfig:  serverConfig,
		//OperatorConfig: *operatorConfig,
		ExtraConfig: apiserver.ExtraConfig{
			ClientConfig:   serverConfig.ClientConfig,
		},
	}
	return config, nil
}

func (o ServerOptions) Run(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	s, err := config.Complete().New()
	if err != nil {
		return err
	}
	//s.GenericAPIServer.AddPostStartHook("start-sample-server-informers", func(context genericapiserver.PostStartHookContext) error {
	//	config.GenericConfig.SharedInformerFactory.Start(context.StopCh)
	//	return nil
	//})

	//return s.GenericAPIServer.PrepareRun().Run(stopCh)
	return s.Run(stopCh)
}
