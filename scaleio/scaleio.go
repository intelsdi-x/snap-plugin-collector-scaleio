/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package scaleio

import (
	"fmt"

	sioclient "github.com/intelsdi-x/snap-plugin-collector-scaleio/scaleio/client"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	name            = "scaleio"
	version         = 2
	statisticsPath  = "/api/instances/StoragePool::%s/relationships/Statistics"
	storagePoolPath = "/api/types/StoragePool/instances"
	NS_VENDOR       = "intel"
	NS_PLUGIN       = "scaleio"
	NS_SP           = "storagePool"
)

// ScaleIO struct implements the collector interface and stores the target
// system URL and credentials
type ScaleIO struct {
	clientCache map[string]*sioclient.SIOClient
}

//NewScaleIOCollector returns an instance of scaleIOCollector
func NewScaleIOCollector() *ScaleIO {
	clientCache := make(map[string]*sioclient.SIOClient)
	return &ScaleIO{
		clientCache: clientCache,
	}
}

// GetConfigPolicy implements the collector interface requirements
func (s *ScaleIO) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	config := plugin.NewConfigPolicy()

	config.AddNewStringRule([]string{"intel", "scaleio"}, "gateway", true)
	config.AddNewStringRule([]string{"intel", "scaleio"}, "username", true)
	config.AddNewStringRule([]string{"intel", "scaleio"}, "password", true)
	config.AddNewBoolRule([]string{"intel", "scaleio"}, "verifySSL", true, plugin.SetDefaultBool(true))

	return *config, nil
}

// GetMetricTypes implements the collector interface requirements
func (s *ScaleIO) GetMetricTypes(_ plugin.Config) ([]plugin.Metric, error) {
	mts := make([]plugin.Metric, len(storagePoolMetricKeys))
	for i := 0; i < len(mts); i++ {
		namespace := plugin.NewNamespace(NS_VENDOR, NS_PLUGIN, NS_SP)
		namespace = namespace.AddDynamicElement("storagePoolID", "The specific storage pool ID to collect from")
		namespace = namespace.AddStaticElements(storagePoolMetricKeys[i]...)
		mts[i].Namespace = namespace
	}
	return mts, nil
}

// CollectMetrics implements the collector interface requirements
func (s *ScaleIO) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	// get the client - this a helper that also initialized the client if needed
	// we get a client from a cache if we have to
	client, err := s.GetSIOClient(mts[0].Config)
	if err != nil {
		return nil, err
	}
	// ensure this is called frequently, we will cache the token and handle expiration
	err = client.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("Failed to authenticate SIO API Client")
	}

	poolReqs := []plugin.Namespace{}

	for _, m := range mts {
		ns := m.Namespace
		switch ns[2].Value {
		case NS_SP:
			poolReqs = append(poolReqs, ns)
		default:
			return nil, fmt.Errorf("Requested metric %s does not match any known scaleio metric", m.Namespace.String())
		}
	}

	metrics := []plugin.Metric{}

	poolMts, err := s.poolMetrics(client, poolReqs)
	if err != nil {
		return nil, err
	}
	metrics = append(metrics, poolMts...)

	return metrics, nil
}

// GetSIOClient returns an SIOClient or creates one as needed
func (s *ScaleIO) GetSIOClient(cfg plugin.Config) (*sioclient.SIOClient, error) {
	var client *sioclient.SIOClient
	gateway, err := cfg.GetString("gateway")
	if err != nil {
		return &sioclient.SIOClient{}, err
	}
	username, err := cfg.GetString("username")
	if err != nil {
		return &sioclient.SIOClient{}, err
	}
	password, err := cfg.GetString("password")
	if err != nil {
		return &sioclient.SIOClient{}, err
	}
	verifySSL, err := cfg.GetBool("verifySSL")
	if err != nil {
		return &sioclient.SIOClient{}, err
	}
	cachedClient, ok := s.clientCache[gateway]
	if !ok {
		newClient, err := sioclient.NewSIOClient(gateway, username, password, verifySSL)
		if err != nil {
			return &sioclient.SIOClient{}, err
		}
		s.clientCache[gateway] = newClient
		client = newClient
	} else {
		// TODO: add check for task config updated - new creds/etc
		client = cachedClient
	}
	return client, nil
}
