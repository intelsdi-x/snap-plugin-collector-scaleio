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

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPluginCreation(t *testing.T) {
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
		_, err := s.GetConfigPolicy()
		Convey("No errors should occur in getting the policy", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestGetMetricTypes(t *testing.T) {
	Convey("GetMetricTypes should return the correct number of metrics", t, func() {
		s := NewScaleIOCollector()
		config := plugin.Config{}
		metrics, err := s.GetMetricTypes(config)
		Convey("Should not return error", func() {
			So(err, ShouldBeNil)
		})
		Convey("Has the correct number of metrics", func() {
			So(metrics, ShouldHaveLength, len(storagePoolMetricKeys))
		})
	})
}
