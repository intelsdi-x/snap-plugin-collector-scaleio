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
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

const (
	name       = "scaleio"
	version    = 1
	pluginType = plugin.CollectorPluginType
)

//Meta returns plugin metadata
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		name,
		version,
		pluginType,
		[]string{plugin.SnapGOBContentType},
		[]string{plugin.SnapGOBContentType})
}

type ScaleIO struct {
}

//NewScaleIOCollector returns an instance of scaleIOCollector
func NewScaleIOCollector() *ScaleIO {
	return &ScaleIO{}
}

func (s *ScaleIO) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	r1, _ := cpolicy.NewStringRule("master", true)
	r2, _ := cpolicy.NewStringRule("username", true)
	r3, _ := cpolicy.NewStringRule("password", true)

	config.Add(r1)
	config.Add(r2)
	config.Add(r3)
	cp.Add([]string{""}, config)
	return cp, nil
}

func (s *ScaleIO) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	return []plugin.MetricType{}, nil
}

func (s *ScaleIO) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	return []plugin.MetricType{}, nil
}
