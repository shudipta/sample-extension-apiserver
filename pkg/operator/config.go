package operator
//
//import (
//	"time"
//
//	"k8s.io/client-go/informers"
//	"k8s.io/client-go/kubernetes"
//	cs "sample-extension-apiserver/client/clientset/versioned"
//	ext_server_informers "sample-extension-apiserver/client/informers/externalversions"
//	kext_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
//	"k8s.io/client-go/rest"
//)
//
//type Config struct {
//	ScratchDir        string
//	ConfigPath        string
//	OperatorNamespace string
//	OpsAddress        string
//	ResyncPeriod      time.Duration
//}
//
//type OperatorConfig struct {
//	Config
//
//	ClientConfig      *rest.Config
//	KubeClient        kubernetes.Interface
//	CRDClient    kext_cs.ApiextensionsV1beta1Interface
//	ExtServerClient   cs.Interface
//}
//
//func NewOperatorConfig(clientConfig *rest.Config) *OperatorConfig {
//	return &OperatorConfig{
//		ClientConfig: clientConfig,
//	}
//}
//
//func (c *OperatorConfig) New() (*Operator, error) {
//	op := &Operator{
//		Config:            c.Config,
//		ClientConfig:      c.ClientConfig,
//		KubeClient:        c.KubeClient,
//		CRDClient:    c.CRDClient,
//		ExtServerClient:   c.ExtServerClient,
//	}
//
//	// ---------------------------
//	op.kubeInformerFactory = informers.NewSharedInformerFactory(op.KubeClient, c.ResyncPeriod)
//	op.extServerInformerFactory = ext_server_informers.NewSharedInformerFactory(op.ExtServerClient, c.ResyncPeriod)
//	// ---------------------------
//	op.setupWorkloadInformers()
//	op.setupCertificateInformers()
//	// ---------------------------
//
//	if err := op.Configure(); err != nil {
//		return nil, err
//	}
//	return op, nil
//}
