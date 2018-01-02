package scaleio

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	storagePoolIDIdx = 3
)

func (s *ScaleIO) poolMetrics(nss []plugin.Namespace) ([]plugin.Metric, error) {

	results := []plugin.Metric{}

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
		for _, ns := range nss {
			// Slice out only the important part for now
			dyn := make([]plugin.NamespaceElement, len(ns))
			copy(dyn, ns)
			dyn[storagePoolIDIdx].Value = id

			currentNamespace := ns.Strings()[storagePoolIDIdx+1:]
			var data interface{}
			if len(currentNamespace) == 1 {
				data = metrics[currentNamespace[0]]
			} else if len(currentNamespace) == 2 {
				subMap, ok := metrics[currentNamespace[0]].(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("Invalid data found for %s with on StoragePool %s", ns, id)
				}
				data = subMap[currentNamespace[1]]
			} else {
				return nil, fmt.Errorf("Invalid metric namespace given: %v", ns)
			}

			newMetric := plugin.Metric{
				Namespace: dyn,
				Timestamp: now,
				Data:      data,
			}
			results = append(results, newMetric)
		}
	}

	return results, nil
}
