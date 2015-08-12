package report

import (
	"fmt"
	"strings"
)

const localUnknown = "localUnknown"

// Topology describes a specific view of a network. It consists of nodes and
// edges, represented by Adjacency, and metadata about those nodes and edges,
// represented by EdgeMetadatas and NodeMetadatas respectively.
type Topology struct {
	Adjacency
	EdgeMetadatas
	NodeMetadatas
}

// Adjacency is an adjacency-list encoding of the topology. Keys are node IDs,
// as produced by the relevant MappingFunc for the topology.
type Adjacency map[string]IDList

// EdgeMetadatas collect metadata about each edge in a topology. Keys are a
// concatenation of node IDs.
type EdgeMetadatas map[string]EdgeMetadata

// NodeMetadatas collect metadata about each node in a topology. Keys are node
// IDs.
type NodeMetadatas map[string]NodeMetadata

// EdgeMetadata describes a superset of the metadata that probes can possibly
// collect about a directed edge between two nodes in any topology.
type EdgeMetadata struct {
	EgressPacketCount  *uint64 `json:"egress_packet_count,omitempty"`
	IngressPacketCount *uint64 `json:"ingress_packet_count,omitempty"`
	EgressByteCount    *uint64 `json:"egress_byte_count,omitempty"`  // Transport layer
	IngressByteCount   *uint64 `json:"ingress_byte_count,omitempty"` // Transport layer
	MaxConnCountTCP    *uint64 `json:"max_conn_count_tcp,omitempty"`
}

// NodeMetadata describes a superset of the metadata that probes can collect
// about a given node in a given topology.
type NodeMetadata struct {
	Metadata map[string]string
}

// MakeNodeMetadata creates a new NodeMetadata with no initial metadata.
func MakeNodeMetadata() NodeMetadata {
	return MakeNodeMetadataWith(map[string]string{})
}

// MakeNodeMetadataWith creates a new NodeMetadata with the supplied map.
func MakeNodeMetadataWith(m map[string]string) NodeMetadata {
	return NodeMetadata{
		Metadata: m,
	}
}

// Copy returns a value copy, useful for tests.
func (nm NodeMetadata) Copy() NodeMetadata {
	cp := MakeNodeMetadata()
	for k, v := range nm.Metadata {
		cp.Metadata[k] = v
	}
	return cp
}

// NewTopology gives you a Topology.
func NewTopology() Topology {
	return Topology{
		Adjacency:     map[string]IDList{},
		EdgeMetadatas: map[string]EdgeMetadata{},
		NodeMetadatas: map[string]NodeMetadata{},
	}
}

// Validate checks the topology for various inconsistencies.
func (t Topology) Validate() error {
	// Check all edge metadata keys must have the appropriate entries in
	// adjacencies & node metadata.
	var errs []string
	for edgeID := range t.EdgeMetadatas {
		srcNodeID, dstNodeID, ok := ParseEdgeID(edgeID)
		if !ok {
			errs = append(errs, fmt.Sprintf("invalid edge ID %q", edgeID))
			continue
		}
		if _, ok := t.NodeMetadatas[srcNodeID]; !ok {
			errs = append(errs, fmt.Sprintf("node metadata missing for source node ID %q (from edge %q)", srcNodeID, edgeID))
		}
		dstNodeIDs, ok := t.Adjacency[MakeAdjacencyID(srcNodeID)]
		if !ok {
			errs = append(errs, fmt.Sprintf("adjacency entries missing for source node ID %q (from edge %q)", srcNodeID, edgeID))
			continue
		}
		if !dstNodeIDs.Contains(dstNodeID) {
			errs = append(errs, fmt.Sprintf("adjacency destination missing for destination node ID %q (from edge %q)", dstNodeID, edgeID))
		}
	}

	// Check all adjancency keys has entries in NodeMetadata.
	for adjacencyID := range t.Adjacency {
		nodeID, ok := ParseAdjacencyID(adjacencyID)
		if !ok {
			errs = append(errs, fmt.Sprintf("invalid adjacency ID %q", adjacencyID))
			continue
		}
		if _, ok := t.NodeMetadatas[nodeID]; !ok {
			errs = append(errs, fmt.Sprintf("node metadata missing for source node %q (from adjacency %q)", nodeID, adjacencyID))
		}
	}

	// Check all node metadatas are valid, and the keys are parseable, i.e.
	// contain a scope.
	for nodeID := range t.NodeMetadatas {
		if t.NodeMetadatas[nodeID].Metadata == nil {
			errs = append(errs, fmt.Sprintf("node ID %q has nil metadata", nodeID))
		}
		if _, _, ok := ParseNodeID(nodeID); !ok {
			errs = append(errs, fmt.Sprintf("invalid node ID %q", nodeID))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s): %s", len(errs), strings.Join(errs, "; "))
	}

	return nil
}

func (t Topology) Copy() Topology {
	other := NewTopology()
	for id, list := range t.Adjacency {
		other.Adjacency[id] = list
	}
	for id, md := range t.EdgeMetadatas {
		other.EdgeMetadatas[id] = md
	}
	for id, md := range t.NodeMetadatas {
		other.NodeMetadatas[id] = md
	}
	return other
}
