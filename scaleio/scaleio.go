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
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	name            = "scaleio"
	version         = 1
	pluginType      = plugin.CollectorPluginType
	statisticsPath  = "/api/instances/StoragePool::%s/relationships/Statistics"
	storagePoolPath = "/api/types/StoragePool/instances"
	NS_VENDOR       = "intel"
	NS_PLUGIN       = "scaleio"
)

//Meta returns plugin metadata
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		name,
		version,
		pluginType,
		[]string{plugin.SnapGOBContentType},
		[]string{plugin.SnapGOBContentType},
		plugin.Exclusive(true),
		plugin.RoutingStrategy(plugin.StickyRouting))
}

type ScaleIO struct {
	token   string
	client  *http.Client
	address *url.URL
}

//NewScaleIOCollector returns an instance of scaleIOCollector
func NewScaleIOCollector() *ScaleIO {
	return &ScaleIO{}

}

func (s *ScaleIO) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	r1, _ := cpolicy.NewStringRule("gateway", true)
	r2, _ := cpolicy.NewStringRule("username", true)
	r3, _ := cpolicy.NewStringRule("password", true)
	r4, _ := cpolicy.NewBoolRule("verifySSL", true, true)

	config.Add(r1)
	config.Add(r2)
	config.Add(r3)
	config.Add(r4)
	cp.Add([]string{""}, config)
	return cp, nil
}

func (s *ScaleIO) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	mts := make([]plugin.MetricType, len(metricKeys))
	for i := 0; i < len(mts); i++ {
		namespace := core.NewNamespace(NS_VENDOR, NS_PLUGIN)
		namespace = namespace.AddDynamicElement("StoragePool", "The specific storage pool ID to collect from")
		namespace = namespace.AddStaticElements(metricKeys[i]...)
		mts[i].Namespace_ = namespace
	}
	return mts, nil
}

func (s *ScaleIO) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	if s.token == "" || s.client == nil || s.address == nil {
		err := s.initConnection(mts[0].Config())
		if err != nil {
			return nil, err
		}
	}
	var returnMts []plugin.MetricType
	// Everything is dynamic right now so get the list of all the StoragePools
	var pools []map[string]interface{}
	err := s.getAPIResponse(storagePoolPath, &pools)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for _, v := range pools {
		id, ok := v["id"].(string)
		if !ok {
			return nil, fmt.Errorf("Found StoragePool entry without an ID")
		}
		var metrics map[string]interface{}
		err := s.getAPIResponse(fmt.Sprintf(statisticsPath, id), &metrics)
		if err != nil {
			return nil, err
		}
		for _, m := range mts {
			// Slice out only the important part for now
			currentNamespace := m.Namespace().Strings()[3:]
			newNS := append([]string{NS_VENDOR, NS_PLUGIN, id}, currentNamespace...)
			var data interface{}
			if len(currentNamespace) == 1 {
				data = metrics[currentNamespace[0]]
			} else if len(currentNamespace) == 2 {
				subMap, ok := metrics[currentNamespace[0]].(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("Invalid data found for %s with on StoragePool %s", m.Namespace(), id)
				}
				data = subMap[currentNamespace[1]]
			} else {
				return nil, fmt.Errorf("Invalid metric namespace given: %v", m.Namespace)
			}
			newMetric := plugin.MetricType{
				Namespace_: core.NewNamespace(newNS...),
				Timestamp_: now,
				Data_:      data,
			}
			returnMts = append(returnMts, newMetric)
		}
	}
	return returnMts, nil
}

func (s *ScaleIO) initConnection(cfg *cdata.ConfigDataNode) error {
	var c *http.Client
	if !cfg.Table()["verifySSL"].(ctypes.ConfigValueBool).Value {
		c = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	} else {
		c = &http.Client{}
	}
	s.client = c
	u, err := url.Parse(cfg.Table()["gateway"].(ctypes.ConfigValueStr).Value)
	if err != nil {
		return fmt.Errorf("Error while parsing gateway URL: %v", err)
	}
	s.address = u
	loginURL := &url.URL{}
	//Make a copy of the base URL
	*loginURL = *s.address
	loginURL.Path = "/api/login"
	req, _ := http.NewRequest("GET", loginURL.String(), nil)
	username := cfg.Table()["username"].(ctypes.ConfigValueStr).Value
	password := cfg.Table()["password"].(ctypes.ConfigValueStr).Value
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error while logging in to ScaleIO API: %v", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	// Strip out the quotes
	body = bytes.Trim(body, "\"")
	body = append([]byte(":"), body...)
	s.token = base64.StdEncoding.EncodeToString(body)
	return nil
}

func (s *ScaleIO) getAPIResponse(path string, v interface{}) error {
	fullURL := &url.URL{}
	*fullURL = *s.address
	fullURL.Path = path
	req, _ := http.NewRequest("GET", fullURL.String(), nil)
	req.Header.Add("Authorization", "Basic "+s.token)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error while accessing ScaleIO API: %v", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("Error while parsing data from %s: %v", path, err)
	}
	return nil
}
