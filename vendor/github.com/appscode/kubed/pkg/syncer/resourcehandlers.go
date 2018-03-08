package syncer

import (
	"reflect"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func (s *ConfigSyncer) ConfigMapHandler() cache.ResourceEventHandler {
	return &configmapSyncer{s}
}

type configmapSyncer struct {
	*ConfigSyncer
}

var _ cache.ResourceEventHandler = &configmapSyncer{}

func (s *configmapSyncer) OnAdd(obj interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.enable {
		return
	}

	if res, ok := obj.(*core.ConfigMap); ok {
		if err := s.SyncConfigMap(res); err != nil {
			log.Errorln(err)
		}
	}
}

func (s *configmapSyncer) OnUpdate(oldObj, newObj interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.enable {
		return
	}

	oldRes, ok := oldObj.(*core.ConfigMap)
	if !ok {
		return
	}
	newRes, ok := newObj.(*core.ConfigMap)
	if !ok {
		return
	}
	if !reflect.DeepEqual(oldRes.Labels, newRes.Labels) ||
		!reflect.DeepEqual(oldRes.Annotations, newRes.Annotations) ||
		!reflect.DeepEqual(oldRes.Data, newRes.Data) {

		if err := s.SyncConfigMap(newRes); err != nil {
			log.Errorln(err)
		}
	}
}

func (s *configmapSyncer) OnDelete(obj interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.enable {
		return
	}

	if res, ok := obj.(*core.ConfigMap); ok {
		if err := s.SyncDeletedConfigMap(res); err != nil {
			log.Errorln(err)
		}
	}
}

func (s *ConfigSyncer) SecretHandler() cache.ResourceEventHandler {
	return &secretSyncer{s}
}

type secretSyncer struct {
	*ConfigSyncer
}

var _ cache.ResourceEventHandler = &secretSyncer{}

func (s *secretSyncer) OnAdd(obj interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.enable {
		return
	}

	if res, ok := obj.(*core.Secret); ok {
		if err := s.SyncSecret(res); err != nil {
			log.Errorln(err)
		}
	}
}

func (s *secretSyncer) OnUpdate(oldObj, newObj interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.enable {
		return
	}

	oldRes, ok := oldObj.(*core.Secret)
	if !ok {
		return
	}
	newRes, ok := newObj.(*core.Secret)
	if !ok {
		return
	}
	if !reflect.DeepEqual(oldRes.Labels, newRes.Labels) ||
		!reflect.DeepEqual(oldRes.Annotations, newRes.Annotations) ||
		!reflect.DeepEqual(oldRes.Data, newRes.Data) {

		if err := s.SyncSecret(newRes); err != nil {
			log.Errorln(err)
		}
	}
}

func (s *secretSyncer) OnDelete(obj interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.enable {
		return
	}

	if res, ok := obj.(*core.Secret); ok {
		if err := s.SyncDeletedSecret(res); err != nil {
			log.Infoln(err)
		}
	}
}

func (s *ConfigSyncer) NamespaceHandler() cache.ResourceEventHandler {
	return &nsSyncer{s}
}

type nsSyncer struct {
	*ConfigSyncer
}

var _ cache.ResourceEventHandler = &secretSyncer{}

func (s *nsSyncer) OnAdd(obj interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.enable {
		return
	}

	if res, ok := obj.(*core.Namespace); ok {
		s.SyncIntoNamespace(res.Name)
	}
}
func (s *nsSyncer) OnUpdate(oldObj, newObj interface{}) {}
func (s *nsSyncer) OnDelete(obj interface{})            {}
