package operator

import (
	"path/filepath"
	"time"

	"github.com/appscode/kubed/pkg/eventer"
	rbin "github.com/appscode/kubed/pkg/recyclebin"
	resource_indexer "github.com/appscode/kubed/pkg/registry/resource"
	"github.com/appscode/kubed/pkg/syncer"
	"github.com/appscode/kutil/tools/fsnotify"
	srch_cs "github.com/appscode/searchlight/client/clientset/versioned"
	searchlightinformers "github.com/appscode/searchlight/client/informers/externalversions"
	scs "github.com/appscode/stash/client/clientset/versioned"
	stashinformers "github.com/appscode/stash/client/informers/externalversions"
	vcs "github.com/appscode/voyager/client/clientset/versioned"
	voyagerinformers "github.com/appscode/voyager/client/informers/externalversions"
	prom "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	kcs "github.com/kubedb/apimachinery/client/clientset/versioned"
	kubedbinformers "github.com/kubedb/apimachinery/client/informers/externalversions"
	"github.com/robfig/cron"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Config struct {
	ScratchDir        string
	ConfigPath        string
	OperatorNamespace string
	OpsAddress        string
	ResyncPeriod      time.Duration
}

type OperatorConfig struct {
	Config

	ClientConfig      *rest.Config
	KubeClient        kubernetes.Interface
	VoyagerClient     vcs.Interface
	SearchlightClient srch_cs.Interface
	StashClient       scs.Interface
	KubeDBClient      kcs.Interface
	PromClient        prom.MonitoringV1Interface
}

func NewOperatorConfig(clientConfig *rest.Config) *OperatorConfig {
	return &OperatorConfig{
		ClientConfig: clientConfig,
	}
}

func (c *OperatorConfig) New() (*Operator, error) {
	op := &Operator{
		Config:            c.Config,
		ClientConfig:      c.ClientConfig,
		KubeClient:        c.KubeClient,
		VoyagerClient:     c.VoyagerClient,
		SearchlightClient: c.SearchlightClient,
		StashClient:       c.StashClient,
		KubeDBClient:      c.KubeDBClient,
		PromClient:        c.PromClient,
	}

	op.recorder = eventer.NewEventRecorder(op.KubeClient, "kubed")
	op.trashCan = &rbin.RecycleBin{}
	op.eventProcessor = &eventer.EventForwarder{Client: op.KubeClient.Discovery()}
	op.configSyncer = syncer.New(op.KubeClient, op.recorder)

	op.cron = cron.New()
	op.cron.Start()

	// Enable full text indexing to have search feature
	indexDir := filepath.Join(c.ScratchDir, "indices")
	op.Indexer = resource_indexer.NewIndexer(indexDir)

	op.watcher = &fsnotify.Watcher{
		WatchDir: filepath.Dir(c.ConfigPath),
		Reload:   op.Configure,
	}

	// ---------------------------
	op.kubeInformerFactory = informers.NewSharedInformerFactory(op.KubeClient, c.ResyncPeriod)
	op.voyagerInformerFactory = voyagerinformers.NewSharedInformerFactory(op.VoyagerClient, c.ResyncPeriod)
	op.stashInformerFactory = stashinformers.NewSharedInformerFactory(op.StashClient, c.ResyncPeriod)
	op.searchlightInformerFactory = searchlightinformers.NewSharedInformerFactory(op.SearchlightClient, c.ResyncPeriod)
	op.kubedbInformerFactory = kubedbinformers.NewSharedInformerFactory(op.KubeDBClient, c.ResyncPeriod)
	// ---------------------------
	op.setupWorkloadInformers()
	op.setupNetworkInformers()
	op.setupConfigInformers()
	op.setupRBACInformers()
	op.setupCoreInformers()
	op.setupEventInformers()
	op.setupCertificateInformers()
	// ---------------------------
	op.setupVoyagerInformers()
	op.setupStashInformers()
	op.setupSearchlightInformers()
	op.setupKubeDBInformers()
	op.setupPrometheusInformers()
	// ---------------------------

	if err := op.Configure(); err != nil {
		return nil, err
	}
	return op, nil
}
