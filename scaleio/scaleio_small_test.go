// +build small

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
	"testing"

	"github.com/intelsdi-x/snap/control/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPluginCreation(t *testing.T) {
	Convey("Meta should return the right data", t, func() {
		metadata := Meta()
		// Check to make sure everything matches
		So(metadata.Name, ShouldResemble, name)
		So(metadata.Version, ShouldResemble, version)
		So(metadata.Type, ShouldResemble, pluginType)
		So(metadata.AcceptedContentTypes, ShouldResemble, []string{plugin.SnapGOBContentType})
		So(metadata.Exclusive, ShouldBeTrue)
		So(metadata.RoutingStrategy, ShouldResemble, plugin.StickyRouting)
	})

	Convey("Should return a new ScaleIO plugin instance", t, func() {
		col := NewScaleIOCollector()
		Convey("Plugin should not be nil", func() {
			So(col, ShouldNotBeNil)
		})

		Convey("Plugin should be of the right type", func() {
			So(col, ShouldHaveSameTypeAs, &ScaleIO{})
		})
	})
}

func TestConfigPolicy(t *testing.T) {
	Convey("GetConfigPolicy should return a valid policy", t, func() {
		s := NewScaleIOCollector()
		policy, err := s.GetConfigPolicy()
		Convey("No errors should occur in getting the policy", func() {
			So(err, ShouldBeNil)
		})
		policies := policy.GetAll()
		Convey("Policy node should exist", func() {
			So(policies, ShouldHaveLength, 1)
		})
		node := policies["intel.scaleio"]
		rules := node.RulesAsTable()
		Convey("All rules should exist in policy node", func() {
			So(node.HasRules(), ShouldBeTrue)
			So(rules, ShouldHaveLength, 4)
		})

		Convey("Rules table should contain gateway, username, and password", func() {
			var names []string
			for _, v := range rules {
				names = append(names, v.Name)
			}
			So(names, ShouldContain, "gateway")
			So(names, ShouldContain, "username")
			So(names, ShouldContain, "password")
		})
	})
}

func TestGetMetricTypes(t *testing.T) {
	Convey("GetMetricTypes should return the correct number of metrics", t, func() {
		s := NewScaleIOCollector()
		metrics, err := s.GetMetricTypes(plugin.NewPluginConfigType())
		Convey("Should not return error", func() {
			So(err, ShouldBeNil)
		})
		Convey("Has the correct number of metrics", func() {
			So(metrics, ShouldHaveLength, len(metricKeys))
		})
	})
}
