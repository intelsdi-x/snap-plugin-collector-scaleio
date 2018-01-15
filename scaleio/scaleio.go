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
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

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
	clientCache map[string]SIOClient
}

// SIOClient stores available clients for usage without needing to reauth
type SIOClient struct {
	token   string
	client  *http.Client
	address *url.URL
}

//NewScaleIOCollector returns an instance of scaleIOCollector
func NewScaleIOCollector() *ScaleIO {
	clientCache := make(map[string]SIOClient)
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
	var client SIOClient
	gateway, _ := mts[0].Config.GetString("gateway")
	cachedClient, ok := s.clientCache[gateway]
	if !ok {
		newClient, err := s.initConnection(mts[0].Config)
		if err != nil {
			return nil, err
		}
		s.clientCache[newClient.address.String()] = newClient
		client = newClient
	} else {
		client = cachedClient
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

func (s *ScaleIO) initConnection(cfg plugin.Config) (SIOClient, error) {
	sioClient := SIOClient{}
	var c *http.Client
	verifySSL, _ := cfg.GetBool("verifySSL")
	if !verifySSL {
		c = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	} else {
		c = &http.Client{}
	}
	sioClient.client = c
	gateway, _ := cfg.GetString("gateway")
	u, err := url.Parse(gateway)
	if err != nil {
		return SIOClient{}, fmt.Errorf("Error while parsing gateway URL: %v", err)
	}
	sioClient.address = u
	loginURL := &url.URL{}
	//Make a copy of the base URL
	*loginURL = *sioClient.address
	loginURL.Path = "/api/login"
	req, _ := http.NewRequest("GET", loginURL.String(), nil)
	username, _ := cfg.GetString("username")
	password, _ := cfg.GetString("password")
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	resp, err := sioClient.client.Do(req)
	if err != nil {
		return SIOClient{}, fmt.Errorf("Error while logging in to ScaleIO API: %v", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	// Strip out the quotes
	body = bytes.Trim(body, "\"")
	body = append([]byte(":"), body...)
	sioClient.token = base64.StdEncoding.EncodeToString(body)
	return sioClient, nil
}

func (s *ScaleIO) getAPIResponse(client SIOClient, path string, v interface{}) error {
	fullURL := &url.URL{}
	*fullURL = *client.address
	fullURL.Path = path
	req, _ := http.NewRequest("GET", fullURL.String(), nil)
	req.Header.Add("Authorization", "Basic "+client.token)
	resp, err := client.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error while accessing ScaleIO API: %v", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("Error while parsing data from %s: %v", path, err)
	}
	return nil
}
