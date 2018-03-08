package operator

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/appscode/envconfig"
	"github.com/appscode/go/log"
	prom_util "github.com/appscode/kube-mon/prometheus/v1"
	api "github.com/appscode/kubed/apis/kubed/v1alpha1"
	_ "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/discovery"
	"github.com/appscode/kutil/tools/backup"
	"github.com/appscode/kutil/tools/fsnotify"
	"github.com/appscode/kutil/tools/queue"
	"github.com/appscode/pat"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	certificates "k8s.io/api/certificates/v1beta1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	storage_v1 "k8s.io/api/storage/v1"
	_ "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	core_informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	cs "sample-extension-apiserver/client/clientset/versioned"
	ext_server_informers "sample-extension-apiserver/client/informers/externalversions"
	kext_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"github.com/docker/docker/builder/dockerfile/shell"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Operator struct {
	Config

	ClientConfig *rest.Config

	KubeClient        kubernetes.Interface
	CRDClient  kext_cs.ApiextensionsV1beta1Interface
	ExtServerClient cs.Interface

	kubeInformerFactory        informers.SharedInformerFactory
	extServerInformerFactory ext_server_informers.SharedInformerFactory

	watcher *fsnotify.Watcher
	lock   sync.RWMutex
}

func (op *Operator) Configure() error {
	log.Infoln("configuring kubed ...")

	op.lock.Lock()
	defer op.lock.Unlock()

	var err error

	cfg, err := api.LoadConfig(op.ConfigPath)
	if err != nil {
		return err
	}
	err = cfg.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (op *Operator) setupWorkloadInformers() {
	deploymentInformer := op.kubeInformerFactory.Apps().V1beta1().Deployments().Informer()
	op.addEventHandlers(deploymentInformer, apps.SchemeGroupVersion.WithKind("Deployment"))
	deploymentInformer.AddEventHandler(op.extractDockerLabel.ExtractFromDeploymentHandler())

	rcInformer := op.kubeInformerFactory.Core().V1().ReplicationControllers().Informer()
	op.addEventHandlers(rcInformer, core.SchemeGroupVersion.WithKind("ReplicationController"))
	rcInformer.AddEventHandler(op.extractDockerLabel.ExtractFromReplicationControllerHandler())

	rsInformer := op.kubeInformerFactory.Extensions().V1beta1().ReplicaSets().Informer()
	op.addEventHandlers(rsInformer, extensions.SchemeGroupVersion.WithKind("ReplicaSet"))
	rsInformer.AddEventHandler(op.extractDockerLabel.ExtractFromReplicaSetHandler())

	daemonSetInformer := op.kubeInformerFactory.Extensions().V1beta1().DaemonSets().Informer()
	op.addEventHandlers(daemonSetInformer, extensions.SchemeGroupVersion.WithKind("DaemonSet"))
	daemonSetInformer.AddEventHandler(op.extractDockerLabel.ExtractFromDaemonSetHandler())

	jobInformer := op.kubeInformerFactory.Batch().V1().Jobs().Informer()
	op.addEventHandlers(jobInformer, batch.SchemeGroupVersion.WithKind("Job"))
	jobInformer.AddEventHandler(op.extractDockerLabel.ExtractFromJobHandler())

	stsInformer := op.kubeInformerFactory.Apps().V1beta1().StatefulSets().Informer()
	op.addEventHandlers(stsInformer, apps.SchemeGroupVersion.WithKind("StatefulSet"))
	stsInformer.AddEventHandler(op.extractDockerLabel.ExtractFromStatefulSetHandler())

	op.kubeInformerFactory.Core().V1().Pods().Informer()
}

func (op *Operator) setupNetworkInformers() {
	svcInformer := op.kubeInformerFactory.Core().V1().Services().Informer()
	op.addEventHandlers(svcInformer, core.SchemeGroupVersion.WithKind("Service"))

	ingressInformer := op.kubeInformerFactory.Extensions().V1beta1().Ingresses().Informer()
	op.addEventHandlers(ingressInformer, extensions.SchemeGroupVersion.WithKind("Ingress"))
}

func (op *Operator) setupConfigInformers() {
	configMapInformer := op.kubeInformerFactory.Core().V1().ConfigMaps().Informer()
	op.addEventHandlers(configMapInformer, core.SchemeGroupVersion.WithKind("ConfigMap"))
	configMapInformer.AddEventHandler(op.configSyncer.ConfigMapHandler())

	secretInformer := op.kubeInformerFactory.Core().V1().Secrets().Informer()
	op.addEventHandlers(secretInformer, core.SchemeGroupVersion.WithKind("Secret"))
	secretInformer.AddEventHandler(op.configSyncer.SecretHandler())

	nsInformer := op.kubeInformerFactory.Core().V1().Namespaces().Informer()
	nsInformer.AddEventHandler(op.configSyncer.NamespaceHandler())
}

func (op *Operator) setupRBACInformers() {
	clusterRoleInformer := op.kubeInformerFactory.Rbac().V1beta1().ClusterRoles().Informer()
	op.addEventHandlers(clusterRoleInformer, rbac.SchemeGroupVersion.WithKind("ClusterRole"))

	clusterRoleBindingInformer := op.kubeInformerFactory.Rbac().V1beta1().ClusterRoleBindings().Informer()
	op.addEventHandlers(clusterRoleBindingInformer, rbac.SchemeGroupVersion.WithKind("ClusterRoleBinding"))

	roleInformer := op.kubeInformerFactory.Rbac().V1beta1().Roles().Informer()
	op.addEventHandlers(roleInformer, rbac.SchemeGroupVersion.WithKind("Role"))

	roleBindingInformer := op.kubeInformerFactory.Rbac().V1beta1().RoleBindings().Informer()
	op.addEventHandlers(roleBindingInformer, rbac.SchemeGroupVersion.WithKind("RoleBinding"))
}

func (op *Operator) setupNodeInformers() {
	nodeInformer := op.kubeInformerFactory.Core().V1().Nodes().Informer()
	op.addEventHandlers(nodeInformer, core.SchemeGroupVersion.WithKind("Node"))
}

func (op *Operator) setupEventInformers() {
	eventInformer := op.kubeInformerFactory.InformerFor(&core.Event{}, func(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return core_informers.NewFilteredEventInformer(
			client,
			core.NamespaceAll,
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			func(options *metav1.ListOptions) {
				options.FieldSelector = fields.OneTermEqualSelector("type", core.EventTypeWarning).String()
			},
		)
	})
	eventInformer.AddEventHandler(op.eventProcessor)
}

func (op *Operator) setupCertificateInformers() {
	csrInformer := op.kubeInformerFactory.Certificates().V1beta1().CertificateSigningRequests().Informer()
	op.addEventHandlers(csrInformer, certificates.SchemeGroupVersion.WithKind("CertificateSigningRequest"))
}

func (op *Operator) setupStorageInformers() {
	pvInformer := op.kubeInformerFactory.Core().V1().PersistentVolumes().Informer()
	op.addEventHandlers(pvInformer, core.SchemeGroupVersion.WithKind("PersistentVolume"))

	pvcInformer := op.kubeInformerFactory.Core().V1().PersistentVolumeClaims().Informer()
	op.addEventHandlers(pvcInformer, core.SchemeGroupVersion.WithKind("PersistentVolumeClaim"))

	storageClassInformer := op.kubeInformerFactory.Storage().V1().StorageClasses().Informer()
	op.addEventHandlers(storageClassInformer, storage_v1.SchemeGroupVersion.WithKind("StorageClass"))
}

func (op *Operator) setupVoyagerInformers() {
	if discovery.IsPreferredAPIResource(op.KubeClient.Discovery(), voyager_api.SchemeGroupVersion.String(), voyager_api.ResourceKindIngress) {
		voyagerIngressInformer := op.voyagerInformerFactory.Voyager().V1beta1().Ingresses().Informer()
		op.addEventHandlers(voyagerIngressInformer, voyager_api.SchemeGroupVersion.WithKind(voyager_api.ResourceKindIngress))

		voyagerCertificateInformer := op.voyagerInformerFactory.Voyager().V1beta1().Certificates().Informer()
		op.addEventHandlers(voyagerCertificateInformer, voyager_api.SchemeGroupVersion.WithKind(voyager_api.ResourceKindCertificate))
	}
}

func (op *Operator) setupStashInformers() {
	if discovery.IsPreferredAPIResource(op.KubeClient.Discovery(), stash_api.SchemeGroupVersion.String(), stash_api.ResourceKindRestic) {
		resticsInformer := op.stashInformerFactory.Stash().V1alpha1().Restics().Informer()
		op.addEventHandlers(resticsInformer, stash_api.SchemeGroupVersion.WithKind(stash_api.ResourceKindRestic))

		recoveryInformer := op.stashInformerFactory.Stash().V1alpha1().Recoveries().Informer()
		op.addEventHandlers(recoveryInformer, stash_api.SchemeGroupVersion.WithKind(stash_api.ResourceKindRecovery))
	}
}

func (op *Operator) setupSearchlightInformers() {
	if discovery.IsPreferredAPIResource(op.KubeClient.Discovery(), searchlight_api.SchemeGroupVersion.String(), searchlight_api.ResourceKindClusterAlert) {
		clusterAlertInformer := op.searchlightInformerFactory.Monitoring().V1alpha1().ClusterAlerts().Informer()
		op.addEventHandlers(clusterAlertInformer, searchlight_api.SchemeGroupVersion.WithKind(searchlight_api.ResourceKindClusterAlert))

		nodeAlertInformer := op.searchlightInformerFactory.Monitoring().V1alpha1().NodeAlerts().Informer()
		op.addEventHandlers(nodeAlertInformer, searchlight_api.SchemeGroupVersion.WithKind(searchlight_api.ResourceKindNodeAlert))

		podAlertInformer := op.searchlightInformerFactory.Monitoring().V1alpha1().PodAlerts().Informer()
		op.addEventHandlers(podAlertInformer, searchlight_api.SchemeGroupVersion.WithKind(searchlight_api.ResourceKindPodAlert))
	}
}

func (op *Operator) setupKubeDBInformers() {
	if discovery.IsPreferredAPIResource(op.KubeClient.Discovery(), kubedb_api.SchemeGroupVersion.String(), kubedb_api.ResourceKindPostgres) {
		pgInformer := op.kubedbInformerFactory.Kubedb().V1alpha1().Postgreses().Informer()
		op.addEventHandlers(pgInformer, kubedb_api.SchemeGroupVersion.WithKind(kubedb_api.ResourceKindPostgres))

		esInformer := op.kubedbInformerFactory.Kubedb().V1alpha1().Elasticsearches().Informer()
		op.addEventHandlers(esInformer, kubedb_api.SchemeGroupVersion.WithKind(kubedb_api.ResourceKindElasticsearch))

		myInformer := op.kubedbInformerFactory.Kubedb().V1alpha1().MySQLs().Informer()
		op.addEventHandlers(myInformer, kubedb_api.SchemeGroupVersion.WithKind(kubedb_api.ResourceKindMySQL))

		mgInformer := op.kubedbInformerFactory.Kubedb().V1alpha1().MongoDBs().Informer()
		op.addEventHandlers(mgInformer, kubedb_api.SchemeGroupVersion.WithKind(kubedb_api.ResourceKindMongoDB))

		rdInformer := op.kubedbInformerFactory.Kubedb().V1alpha1().Redises().Informer()
		op.addEventHandlers(rdInformer, kubedb_api.SchemeGroupVersion.WithKind(kubedb_api.ResourceKindRedis))

		mcInformer := op.kubedbInformerFactory.Kubedb().V1alpha1().Memcacheds().Informer()
		op.addEventHandlers(mcInformer, kubedb_api.SchemeGroupVersion.WithKind(kubedb_api.ResourceKindMemcached))

		dbSnapshotInformer := op.kubedbInformerFactory.Kubedb().V1alpha1().Snapshots().Informer()
		op.addEventHandlers(dbSnapshotInformer, kubedb_api.SchemeGroupVersion.WithKind(kubedb_api.ResourceKindSnapshot))

		dormantDatabaseInformer := op.kubedbInformerFactory.Kubedb().V1alpha1().DormantDatabases().Informer()
		op.addEventHandlers(dormantDatabaseInformer, kubedb_api.SchemeGroupVersion.WithKind(kubedb_api.ResourceKindDormantDatabase))
	}
}

func (op *Operator) setupPrometheusInformers() {
	if discovery.IsPreferredAPIResource(op.KubeClient.Discovery(), prom_util.SchemeGroupVersion.String(), prom.PrometheusesKind) {
		op.promInf = cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc:  op.PromClient.Prometheuses(core.NamespaceAll).List,
				WatchFunc: op.PromClient.Prometheuses(core.NamespaceAll).Watch,
			},
			&prom.Prometheus{}, op.ResyncPeriod, cache.Indexers{},
		)
		op.addEventHandlers(op.promInf, prom_util.SchemeGroupVersion.WithKind(prom.PrometheusesKind))

		op.smonInf = cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc:  op.PromClient.ServiceMonitors(core.NamespaceAll).List,
				WatchFunc: op.PromClient.ServiceMonitors(core.NamespaceAll).Watch,
			},
			&prom.ServiceMonitor{}, op.ResyncPeriod, cache.Indexers{},
		)
		op.addEventHandlers(op.smonInf, prom_util.SchemeGroupVersion.WithKind(prom.ServiceMonitorsKind))

		op.amgrInf = cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc:  op.PromClient.Alertmanagers(core.NamespaceAll).List,
				WatchFunc: op.PromClient.Alertmanagers(core.NamespaceAll).Watch,
			},
			&prom.Alertmanager{}, op.ResyncPeriod, cache.Indexers{},
		)
		op.addEventHandlers(op.amgrInf, prom_util.SchemeGroupVersion.WithKind(prom.AlertmanagersKind))
	}
}

func (op *Operator) addEventHandlers(informer cache.SharedIndexInformer, gvk schema.GroupVersionKind) {
	informer.AddEventHandler(queue.NewVersionedHandler(op.trashCan, gvk))
	informer.AddEventHandler(queue.NewVersionedHandler(op.eventProcessor, gvk))
	informer.AddEventHandler(queue.NewVersionedHandler(op.Indexer, gvk))
}

func (op *Operator) getLoader() (envconfig.LoaderFunc, error) {
	if op.config.NotifierSecretName == "" {
		return func(key string) (string, bool) {
			return "", false
		}, nil
	}
	cfg, err := op.KubeClient.CoreV1().
		Secrets(op.OperatorNamespace).
		Get(op.config.NotifierSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return func(key string) (value string, found bool) {
		var bytes []byte
		bytes, found = cfg.Data[key]
		value = string(bytes)
		return
	}, nil
}

func (op *Operator) RunWatchers(stopCh <-chan struct{}) {
	op.kubeInformerFactory.Start(stopCh)

	var res map[reflect.Type]bool

	res = op.kubeInformerFactory.WaitForCacheSync(stopCh)
	for _, v := range res {
		if !v {
			runtime.HandleError(errors.Errorf("timed out waiting for caches to sync"))
			return
		}
	}
}

func (op *Operator) Run(stopCh <-chan struct{}) error {
	op.RunWatchers(stopCh)

	go op.watcher.Run(stopCh)

	m := pat.New()
	m.Get("/metrics", promhttp.Handler())
	http.Handle("/", m)
	log.Infoln("Listening on", op.OpsAddress)
	return http.ListenAndServe(op.OpsAddress, nil)
}
