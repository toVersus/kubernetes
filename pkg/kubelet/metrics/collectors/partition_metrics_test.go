/*
Copyright The Kubernetes Authors.

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

package collectors

import (
	"strings"
	"testing"

	"k8s.io/component-base/metrics/testutil"
	"k8s.io/kubernetes/pkg/kubelet/cm"
	"k8s.io/utils/ptr"
)

func TestPartitionMetricsCollector(t *testing.T) {
	const (
		memoryUsageHeader = `
		# HELP kubelet_partition_memory_usage_bytes [ALPHA] Memory in bytes charged to the partition's cgroup, including its page cache.
		# TYPE kubelet_partition_memory_usage_bytes gauge
`
		memoryLimitHeader = `
		# HELP kubelet_partition_memory_limit_bytes [ALPHA] Memory limit in bytes of the partition's cgroup.
		# TYPE kubelet_partition_memory_limit_bytes gauge
`
		podsHeader = `
		# HELP kubelet_partition_pods [ALPHA] Number of active pods in the partition.
		# TYPE kubelet_partition_pods gauge
`
	)

	tests := []struct {
		name  string
		stats *cm.SystemPartitionStats
		want  string
	}{
		{
			name: "the system partition reports its usage, limit and pods",
			stats: &cm.SystemPartitionStats{
				MemoryUsageBytes: ptr.To[int64](536870912),
				MemoryLimitBytes: ptr.To[int64](1073741824),
				Pods:             3,
			},
			want: memoryUsageHeader + `
		kubelet_partition_memory_usage_bytes{partition="system"} 5.36870912e+08
` + memoryLimitHeader + `
		kubelet_partition_memory_limit_bytes{partition="system"} 1.073741824e+09
` + podsHeader + `
		kubelet_partition_pods{partition="system"} 3
`,
		},
		{
			name:  "an unread usage and an unset limit are left out",
			stats: &cm.SystemPartitionStats{},
			want: podsHeader + `
		kubelet_partition_pods{partition="system"} 0
`,
		},
		{
			name: "a node without a system partition reports nothing",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset descriptors because each subtest uses its own registry.
			descPartitionMemoryUsage = descPartitionMemoryUsage.GetRawDesc()
			descPartitionMemoryLimit = descPartitionMemoryLimit.GetRawDesc()
			descPartitionPods = descPartitionPods.GetRawDesc()

			collector := &partitionMetricsCollector{
				stats: func() *cm.SystemPartitionStats { return tc.stats },
			}

			err := testutil.CustomCollectAndCompare(collector, strings.NewReader(tc.want),
				"kubelet_partition_memory_usage_bytes", "kubelet_partition_memory_limit_bytes", "kubelet_partition_pods")
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
