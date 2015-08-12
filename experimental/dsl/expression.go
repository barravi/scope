package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/weaveworks/scope/report"
)

type expression struct {
	selector
	transformer
}

func (e expression) eval(tpy report.Topology) report.Topology {
	return e.transformer(tpy, e.selector(tpy))
}

type selector func(report.Topology) []string

type transformer func(report.Topology, []string) report.Topology

func selectAll(tpy report.Topology) []string {
	out := make([]string, 0, len(tpy.NodeMetadatas))
	for id := range tpy.NodeMetadatas {
		out = append(out, id)
	}
	log.Printf("All: %d", len(out))
	return out
}

func selectConnected(tpy report.Topology) []string {
	return []string{} // TODO
}

func selectTouched(tpy report.Topology) []string {
	return []string{} // TODO
}

func selectLike(s string) selector {
	return func(tpy report.Topology) []string {
		return []string{} // TODO
	}
}

func selectWith(s string) selector {
	var k, v string
	if fields := strings.SplitN(s, "=", 2); len(fields) == 1 {
		k = strings.TrimSpace(fields[0])
	} else if len(fields) == 2 {
		k, v = strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])
	}

	return func(tpy report.Topology) []string {
		out := []string{}
		for id, md := range tpy.NodeMetadatas {
			if vv, ok := md.Metadata[k]; ok {
				if v == "" || (v != "" && v == vv) {
					out = append(out, id)
				}
			}
		}
		return out
	}
}

func selectNot(s selector) selector {
	return func(tpy report.Topology) []string {
		set := map[string]struct{}{}
		for _, id := range s(tpy) {
			set[id] = struct{}{}
		}
		out := []string{}
		for id := range tpy.NodeMetadatas {
			if _, ok := set[id]; ok {
				continue // selected by that one -> not by this one
			}
			out = append(out, id)
		}
		return out
	}
}

func transformRemove(tpy report.Topology, ids []string) report.Topology {
	toRemove := map[string]struct{}{}
	for _, id := range ids {
		toRemove[id] = struct{}{}
	}
	out := report.NewTopology()
	for id := range tpy.NodeMetadatas {
		if _, ok := toRemove[id]; ok {
			continue
		}
		cp(out, tpy, id)
	}
	log.Printf("Remove %d: in %d, out %d", len(ids), len(tpy.NodeMetadatas), len(out.NodeMetadatas))
	return out
}

func transformShowOnly(tpy report.Topology, ids []string) report.Topology {
	out := report.NewTopology()
	for _, id := range ids {
		if _, ok := tpy.NodeMetadatas[id]; !ok {
			continue
		}
		cp(out, tpy, id)
	}
	log.Printf("ShowOnly %d: in %d, out %d", len(ids), len(tpy.NodeMetadatas), len(out.NodeMetadatas))
	return out
}

func transformMerge(tpy report.Topology, ids []string) report.Topology {
	name := fmt.Sprintf("%x", rand.Int31())
	toMerge := map[string]struct{}{}
	for _, id := range ids {
		toMerge[id] = struct{}{}
	}
	out := report.NewTopology()
	for id := range tpy.NodeMetadatas {
		if _, ok := toMerge[id]; ok {
			merge(out, name, tpy, id)
		} else {
			cp(out, tpy, id)
		}
	}
	return out
}

func transformGroupBy(s string) transformer {
	prefix := fmt.Sprintf("%x", rand.Int31())

	keys := map[string]struct{}{}
	for _, key := range strings.Split(s, ",") {
		keys[strings.TrimSpace(key)] = struct{}{}
	}

	return func(tpy report.Topology, ids []string) report.Topology {
		set := map[string]struct{}{}
		for _, id := range ids {
			set[id] = struct{}{}
		}

		// Identify all nodes that should be grouped.
		toMerge := map[string]string{} // src ID: dst ID
		for id, md := range tpy.NodeMetadatas {
			if _, ok := set[id]; !ok {
				continue // not selected
			}
			for k, v := range md.Metadata {
				if _, ok := keys[k]; ok {
					toMerge[id] = fmt.Sprintf("%s-%s-%s", prefix, k, v)
				}
			}
		}

		// Walk nodes again, merging those that should be grouped.
		out := report.NewTopology()
		for id := range tpy.NodeMetadatas {
			if dstID, ok := toMerge[id]; ok {
				merge(out, dstID, tpy, id)
			} else {
				cp(out, tpy, id)
			}
		}

		log.Printf("GroupBy (%v) %d: in %d, out %d", keys, len(ids), len(tpy.NodeMetadatas), len(out.NodeMetadatas))
		return out
	}
}

func cp(dst report.Topology, src report.Topology, id string) {
	dst.NodeMetadatas[id] = src.NodeMetadatas[id]
	dst.NodeMetadatas[id].Merge(report.MakeNodeMetadataWith(map[string]string{"touched": "true"}))
	dst.Adjacency[id] = src.Adjacency[id]
	for _, otherID := range dst.Adjacency[id] {
		edgeID := report.MakeEdgeID(id, otherID)
		dst.EdgeMetadatas[edgeID] = src.EdgeMetadatas[edgeID]
	}
}

func merge(dst report.Topology, dstID string, src report.Topology, srcID string) {
	md, ok := dst.NodeMetadatas[dstID]
	if !ok {
		md = report.MakeNodeMetadata()
	}
	md.Merge(report.MakeNodeMetadataWith(map[string]string{"touched": "true"}))
	md.Merge(src.NodeMetadatas[srcID])
	dst.NodeMetadatas[dstID] = md

	ids := dst.Adjacency[dstID]
	ids.Merge(src.Adjacency[srcID])
	dst.Adjacency[dstID] = ids

	for _, otherID := range src.Adjacency[srcID] {
		oldEdgeID := report.MakeEdgeID(srcID, otherID)
		newEdgeID := report.MakeEdgeID(dstID, otherID)
		edgeMetadatas := dst.EdgeMetadatas[newEdgeID]
		edgeMetadatas.Merge(src.EdgeMetadatas[oldEdgeID])
		dst.EdgeMetadatas[newEdgeID] = edgeMetadatas
	}
}
