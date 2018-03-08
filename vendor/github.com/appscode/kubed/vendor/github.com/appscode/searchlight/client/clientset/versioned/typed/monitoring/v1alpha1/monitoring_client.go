/*
Copyright 2018 The Searchlight Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	v1alpha1 "github.com/appscode/searchlight/apis/monitoring/v1alpha1"
	"github.com/appscode/searchlight/client/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type MonitoringV1alpha1Interface interface {
	RESTClient() rest.Interface
	ClusterAlertsGetter
	NodeAlertsGetter
	PodAlertsGetter
}

// MonitoringV1alpha1Client is used to interact with features provided by the monitoring.appscode.com group.
type MonitoringV1alpha1Client struct {
	restClient rest.Interface
}

func (c *MonitoringV1alpha1Client) ClusterAlerts(namespace string) ClusterAlertInterface {
	return newClusterAlerts(c, namespace)
}

func (c *MonitoringV1alpha1Client) NodeAlerts(namespace string) NodeAlertInterface {
	return newNodeAlerts(c, namespace)
}

func (c *MonitoringV1alpha1Client) PodAlerts(namespace string) PodAlertInterface {
	return newPodAlerts(c, namespace)
}

// NewForConfig creates a new MonitoringV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*MonitoringV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &MonitoringV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new MonitoringV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *MonitoringV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new MonitoringV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *MonitoringV1alpha1Client {
	return &MonitoringV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *MonitoringV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
