package server

import (
	"io"
	"net"

	api "github.com/appscode/kubed/apis/kubed/v1alpha1"
	"github.com/appscode/kubed/pkg/operator"
	"github.com/appscode/kubed/pkg/server"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

const defaultEtcdPathPrefix = "/registry/kubed.appscode.com"

type KubedOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	OperatorOptions    *OperatorOptions

	StdOut io.Writer
	StdErr io.Writer
}

func NewKubedOptions(out, errOut io.Writer) *KubedOptions {
	o := &KubedOptions{
		// TODO we will nil out the etcd storage options.  This requires a later level of k8s.io/apiserver
		RecommendedOptions: genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix, server.Codecs.LegacyCodec(api.SchemeGroupVersion)),
		OperatorOptions:    NewOperatorOptions(),
		StdOut:             out,
		StdErr:             errOut,
	}
	o.RecommendedOptions.Etcd = nil

	return o
}

func (o *KubedOptions) AddFlags(fs *pflag.FlagSet) {
	o.RecommendedOptions.AddFlags(fs)
	o.OperatorOptions.AddFlags(fs)
}

func (o KubedOptions) Validate(args []string) error {
	return nil
}

func (o *KubedOptions) Complete() error {
	return nil
}

func (o KubedOptions) Config() (*server.KubedConfig, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, errors.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(server.Codecs)
	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	operatorConfig := operator.NewOperatorConfig(serverConfig.ClientConfig)
	if err := o.OperatorOptions.ApplyTo(operatorConfig); err != nil {
		return nil, err
	}

	config := &server.KubedConfig{
		GenericConfig:  serverConfig,
		OperatorConfig: operatorConfig,
	}
	return config, nil
}

func (o KubedOptions) Run(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	s, err := config.Complete().New()
	if err != nil {
		return err
	}

	return s.Run(stopCh)
}
