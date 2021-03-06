package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

const (
	PID    = "pid"
	Name   = "name"
	Domain = "domain"
)

func TestMergeAdjacency(t *testing.T) {
	for name, c := range map[string]struct {
		a, b, want report.Adjacency
	}{
		"Empty b": {
			a: report.Adjacency{
				"hostA|:192.168.1.1:12345": report.MakeIDList(":192.168.1.2:80"),
				"hostA|:192.168.1.1:8888":  report.MakeIDList(":1.2.3.4:22"),
				"hostB|:192.168.1.2:80":    report.MakeIDList(":192.168.1.1:12345"),
			},
			b: report.Adjacency{},
			want: report.Adjacency{
				"hostA|:192.168.1.1:12345": report.MakeIDList(":192.168.1.2:80"),
				"hostA|:192.168.1.1:8888":  report.MakeIDList(":1.2.3.4:22"),
				"hostB|:192.168.1.2:80":    report.MakeIDList(":192.168.1.1:12345"),
			},
		},
		"Empty a": {
			a: report.Adjacency{},
			b: report.Adjacency{
				"hostA|:192.168.1.1:12345": report.MakeIDList(":192.168.1.2:80"),
				"hostA|:192.168.1.1:8888":  report.MakeIDList(":1.2.3.4:22"),
				"hostB|:192.168.1.2:80":    report.MakeIDList(":192.168.1.1:12345"),
			},
			want: report.Adjacency{
				"hostA|:192.168.1.1:12345": report.MakeIDList(":192.168.1.2:80"),
				"hostA|:192.168.1.1:8888":  report.MakeIDList(":1.2.3.4:22"),
				"hostB|:192.168.1.2:80":    report.MakeIDList(":192.168.1.1:12345"),
			},
		},
		"Same address": {
			a: report.Adjacency{
				"hostA|:192.168.1.1:12345": report.MakeIDList(":192.168.1.2:80"),
			},
			b: report.Adjacency{
				"hostA|:192.168.1.1:12345": report.MakeIDList(":192.168.1.2:8080"),
			},
			want: report.Adjacency{
				"hostA|:192.168.1.1:12345": report.MakeIDList(
					":192.168.1.2:80", ":192.168.1.2:8080",
				),
			},
		},
		"No duplicates": {
			a: report.Adjacency{
				"hostA|:192.168.1.1:12345": report.MakeIDList(
					":192.168.1.2:80",
					":192.168.1.2:8080",
					":192.168.1.2:555",
				),
			},
			b: report.Adjacency{
				"hostA|:192.168.1.1:12345": report.MakeIDList(
					":192.168.1.2:8080",
					":192.168.1.2:80",
					":192.168.1.2:444",
				),
			},
			want: report.Adjacency{
				"hostA|:192.168.1.1:12345": []string{
					":192.168.1.2:444",
					":192.168.1.2:555",
					":192.168.1.2:80",
					":192.168.1.2:8080",
				},
			},
		},
		"Double keys": {
			a: report.Adjacency{
				"key1": report.MakeIDList("a", "c", "d", "b"),
				"key2": report.MakeIDList("c", "a"),
			},
			b: report.Adjacency{
				"key1": report.MakeIDList("a", "b", "e"),
				"key3": report.MakeIDList("e", "a", "a", "a", "e"),
			},
			want: report.Adjacency{
				"key1": report.MakeIDList("a", "b", "c", "d", "e"),
				"key2": report.MakeIDList("a", "c"),
				"key3": report.MakeIDList("a", "e"),
			},
		},
	} {
		have := c.a
		have.Merge(c.b)
		if !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s: want\n\t%#v\nhave\n\t%#v", name, c.want, have)
		}
	}
}

func TestMergeEdgeMetadatas(t *testing.T) {
	for name, c := range map[string]struct {
		a, b, want report.EdgeMetadatas
	}{
		"Empty a": {
			a: report.EdgeMetadatas{},
			b: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(1),
					MaxConnCountTCP:   newu64(2),
				},
			},
			want: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(1),
					MaxConnCountTCP:   newu64(2),
				},
			},
		},
		"Empty b": {
			a: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(999),
				},
			},
			b: report.EdgeMetadatas{},
			want: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(999),
				},
			},
		},
		"Host merge": {
			a: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(500),
					MaxConnCountTCP:   newu64(4),
				},
			},
			b: report.EdgeMetadatas{
				"hostQ|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(1),
					EgressByteCount:   newu64(2),
					MaxConnCountTCP:   newu64(6),
				},
			},
			want: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(500),
					MaxConnCountTCP:   newu64(4),
				},
				"hostQ|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(1),
					EgressByteCount:   newu64(2),
					MaxConnCountTCP:   newu64(6),
				},
			},
		},
		"Edge merge": {
			a: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(12),
					EgressByteCount:   newu64(1000),
					MaxConnCountTCP:   newu64(7),
				},
			},
			b: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(1),
					IngressByteCount:  newu64(123),
					EgressByteCount:   newu64(2),
					MaxConnCountTCP:   newu64(9),
				},
			},
			want: report.EdgeMetadatas{
				"hostA|:192.168.1.1:12345|:192.168.1.2:80": report.EdgeMetadata{
					EgressPacketCount: newu64(13),
					IngressByteCount:  newu64(123),
					EgressByteCount:   newu64(1002),
					MaxConnCountTCP:   newu64(9),
				},
			},
		},
	} {
		have := c.a
		have.Merge(c.b)

		if !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s: want\n\t%#v, have\n\t%#v", name, c.want, have)
		}
	}
}

func TestMergeNodeMetadatas(t *testing.T) {
	for name, c := range map[string]struct {
		a, b, want report.NodeMetadatas
	}{
		"Empty a": {
			a: report.NodeMetadatas{},
			b: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			want: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Empty b": {
			a: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.NodeMetadatas{},
			want: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Simple merge": {
			a: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.NodeMetadatas{
				":192.168.1.2:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "42",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			want: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
				":192.168.1.2:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "42",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Merge conflict": {
			a: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{ // <-- same ID
					PID:    "0",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			want: report.NodeMetadatas{
				":192.168.1.1:12345": report.MakeNodeMetadataWith(map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
	} {
		have := c.a
		have.Merge(c.b)

		if !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s: want\n\t%#v, have\n\t%#v", name, c.want, have)
		}
	}
}

func newu64(value uint64) *uint64 { return &value }
