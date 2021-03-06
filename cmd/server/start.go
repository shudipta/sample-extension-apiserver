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
	//"time"
	"github.com/spf13/pflag"
	"github.com/golang/glog"
)

const defaultEtcdPathPrefix = "/registry/somethingcontroller.kube-ac.com"

type ServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	//OperatorOptions    *OperatorOptions

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
	//opt.RecommendedOptions.Etcd = nil
	//opt.RecommendedOptions.SecureServing.BindPort = 8443

	return opt
}

func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.RecommendedOptions.AddFlags(fs)
	//o.OperatorOptions.AddFlags(fs)
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
		//ExtraConfig: apiserver.ExtraConfig{
		//	ClientConfig:   serverConfig.ClientConfig,
		//},
	}
	return config, nil
}

func (o ServerOptions) Run(stopCh <-chan struct{}) error {
	glog.Infoln("running........")
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

	if err := s.GenericAPIServer.PrepareRun().Run(stopCh); err != nil {
		glog.Fatal(err)
	}

	//return s.Run(stopCh)
	return nil
}
