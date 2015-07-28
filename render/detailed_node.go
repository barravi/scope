package render

import (
	"fmt"
	"log"
	"time"

	"github.com/weaveworks/scope/probe/host"

	"github.com/weaveworks/scope/report"
)

// DetailedNode is the data type that's yielded to the JavaScript layer when
// we want deep information about an individual node.
type DetailedNode struct {
	ID         string             `json:"id"`                    // unique
	LabelMajor string             `json:"label_major"`           // big header, e.g. "container_name"
	LabelMinor string             `json:"label_minor,omitempty"` // small header, e.g. "host.domain.org"
	Pseudo     bool               `json:"pseudo,omitempty"`      // if true, changes rendering
	Sections   map[string]Section `json:"sections,omitempty"`
}

// Section is a simple JSON-marshalable key-value map. It's deliberately kept
// underspecified for right now, as we want some flexibility to change what
// data we submit to the UI for rendering while we figure out what makes
// sense.
type Section map[string]interface{}

// These variables are keys in the sections map.
var (
	SectionHosts      = "hosts"
	SectionEndpoints  = "endpoints"
	SectionAddresses  = "addresses"
	SectionContainers = "containers"
	SectionProcesses  = "processes"
)

// MakeDetailedNode transforms a renderable node to a detailed node. It uses
// the renderable node's metadata to perform lookups in the report struct for
// the various sections.
func MakeDetailedNode(rpt report.Report, n RenderableNode) DetailedNode {
	sections := map[string]Section{}
	for key, f := range map[string]func(report.Report, RenderableNode) Section{
		SectionHosts:      makeHostsSection,      // origin host(s)
		SectionEndpoints:  makeEndpointsSection,  // port-to-port connections
		SectionAddresses:  makeAddressesSection,  // address-to-address connections
		SectionContainers: makeContainersSection, // contributing container(s)
		SectionProcesses:  makeProcessesSection,  // contributing process(es)
	} {
		if section := f(rpt, n); len(section) > 0 {
			sections[key] = section
		}
	}
	return DetailedNode{
		ID:         n.ID,
		LabelMajor: n.LabelMajor,
		LabelMinor: n.LabelMinor,
		Pseudo:     n.Pseudo,
		Sections:   sections,
	}
}

func makeHostsSection(rpt report.Report, n RenderableNode) Section {
	section := Section{}
	for _, id := range n.Origins {
		if nmd, ok := rpt.Host.NodeMetadatas[id]; ok {
			section[id] = map[string]string{
				host.HostName:      nmd.Metadata[host.HostName],
				host.OS:            nmd.Metadata[host.OS],
				host.Load:          nmd.Metadata[host.Load],
				host.Uptime:        nmd.Metadata[host.Uptime],
				host.KernelVersion: nmd.Metadata[host.KernelVersion],
			}
		}
	}
	return section
}

// These are keys for endpoint and address sections.
const (
	ConnectionSrc        = "src"
	ConnectionDst        = "dst"
	ConnectionPacketRate = "packet_rate"
	ConnectionByteRate   = "byte_rate"
)

func makeEndpointsSection(rpt report.Report, n RenderableNode) Section {
	section := Section{}
	for _, srcNodeID := range n.Origins {
		if _, ok := rpt.Endpoint.NodeMetadatas[srcNodeID]; ok {
			adjacencyID := report.MakeAdjacencyID(srcNodeID)
			for _, dstNodeID := range rpt.Endpoint.Adjacency[adjacencyID] {
				edgeID := report.MakeEdgeID(srcNodeID, dstNodeID)
				section[edgeID] = makeEndpointSection(rpt.Endpoint, srcNodeID, dstNodeID, rpt.Window)
			}
		}
	}
	return section
}

func makeEndpointSection(t report.Topology, srcNodeID, dstNodeID string, window time.Duration) Section {
	_, srcAddr, srcPort, ok := report.ParseEndpointNodeID(srcNodeID)
	if !ok {
		log.Printf("endpoint section: bad src ID %q", srcNodeID)
		return Section{}
	}

	_, dstAddr, dstPort, ok := report.ParseEndpointNodeID(dstNodeID)
	if !ok {
		log.Printf("endpoint section: bad src ID %q", srcNodeID)
		return Section{}
	}

	section := Section{
		ConnectionSrc: fmt.Sprintf("%s:%s", srcAddr, srcPort),
		ConnectionDst: fmt.Sprintf("%s:%s", dstAddr, dstPort),
	}

	edgeID := report.MakeEdgeID(srcNodeID, dstNodeID)
	if emd, ok := t.EdgeMetadatas[edgeID]; ok {
		if emd.PacketCount != nil {
			section[ConnectionPacketRate] = fmt.Sprint(uint64(float64(*emd.PacketCount) / window.Seconds()))
		}
		if emd.ByteCount != nil {
			section[ConnectionByteRate] = fmt.Sprint(uint64(float64(*emd.ByteCount) / window.Seconds()))
		}
	}

	return section
}

func makeAddressesSection(rpt report.Report, n RenderableNode) Section {
	return Section{} // TODO
}

func makeContainersSection(rpt report.Report, n RenderableNode) Section {
	return Section{} // TODO
}

func makeProcessesSection(rpt report.Report, n RenderableNode) Section {
	return Section{} // TODO
}

/*
const (
	mb                 = 1 << 20
	connectionsRank    = 100
	containerImageRank = 4
	containerRank      = 3
	processRank        = 2
	hostRank           = 1
	endpointRank       = 0 // this is the least important table, so sort to bottom
	addressRank        = 0 // also least important; never merged with endpoints
)

// DetailedNode is the data type that's yielded to the JavaScript layer when
// we want deep information about an individual node.
type DetailedNode struct {
	ID         string  `json:"id"`
	LabelMajor string  `json:"label_major"`
	LabelMinor string  `json:"label_minor,omitempty"`
	Pseudo     bool    `json:"pseudo,omitempty"`
	Tables     []Table `json:"tables"`
}

// Table is a dataset associated with a node. It will be displayed in the
// detail panel when a user clicks on a node.
type Table struct {
	Title   string `json:"title"`   // e.g. Bandwidth
	Numeric bool   `json:"numeric"` // should the major column be right-aligned?
	Rank    int    `json:"-"`       // used to sort tables; not emitted.
	Rows    []Row  `json:"rows"`
}

// Row is a single entry in a Table dataset.
type Row struct {
	Key        string `json:"key"`                   // e.g. Ingress
	ValueMajor string `json:"value_major"`           // e.g. 25
	ValueMinor string `json:"value_minor,omitempty"` // e.g. KB/s
}

type tables []Table

func (t tables) Len() int           { return len(t) }
func (t tables) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t tables) Less(i, j int) bool { return t[i].Rank > t[j].Rank }

// MakeDetailedNode transforms a renderable node to a detailed node. It uses
// aggregate metadata, plus the set of origin node IDs, to produce tables.
func MakeDetailedNode(r report.Report, n RenderableNode) DetailedNode {
	tables := tables{}
	{
		rows := []Row{}
		if n.EdgeMetadata.MaxConnCountTCP != nil {
			rows = append(rows, Row{"TCP connections", strconv.FormatUint(*n.EdgeMetadata.MaxConnCountTCP, 10), ""})
		}
		if n.EdgeMetadata.PacketCount != nil {
			rows = append(rows, Row{"Packets", strconv.FormatUint(*n.EdgeMetadata.PacketCount, 10), ""})
		}
		if n.EdgeMetadata.ByteCount != nil {
			rows = append(rows, Row{"Bytes", strconv.FormatUint(*n.EdgeMetadata.ByteCount, 10), ""})
		}
		if len(rows) > 0 {
			tables = append(tables, Table{"Connections", true, connectionsRank, rows})
		}
	}

	// RenderableNode may be the result of merge operation(s), and so may have
	// multiple origins. The ultimate goal here is to generate tables to view
	// in the UI, so we skip the intermediate representations, but we could
	// add them later.
	connections := []Row{}
	for _, id := range n.Origins {
		if table, ok := OriginTable(r, id); ok {
			tables = append(tables, table)
		} else if nmd, ok := r.Endpoint.NodeMetadatas[id]; ok {
			connections = append(connections, connectionDetailsRows(r.Endpoint, id, nmd)...)
		}
	}
	if len(connections) > 0 {
		tables = append(tables, connectionDetailsTable(connections))
	}

	// Sort tables by rank
	sort.Sort(tables)

	return DetailedNode{
		ID:         n.ID,
		LabelMajor: n.LabelMajor,
		LabelMinor: n.LabelMinor,
		Pseudo:     n.Pseudo,
		Tables:     tables,
	}
}

// OriginTable produces a table (to be consumed directly by the UI) based on
// an origin ID, which is (optimistically) a node ID in one of our topologies.
func OriginTable(r report.Report, originID string) (Table, bool) {
	if nmd, ok := r.Address.NodeMetadatas[originID]; ok {
		return addressOriginTable(nmd)
	}
	if nmd, ok := r.Process.NodeMetadatas[originID]; ok {
		return processOriginTable(nmd)
	}
	if nmd, ok := r.Container.NodeMetadatas[originID]; ok {
		return containerOriginTable(nmd)
	}
	if nmd, ok := r.ContainerImage.NodeMetadatas[originID]; ok {
		return containerImageOriginTable(nmd)
	}
	if nmd, ok := r.Host.NodeMetadatas[originID]; ok {
		return hostOriginTable(nmd)
	}
	return Table{}, false
}

func connectionDetailsRows(endpointTopology report.Topology, originID string, nmd report.NodeMetadata) []Row {
	rows := []Row{}
	local := fmt.Sprintf("%s:%s", nmd.Metadata[docker.Addr], nmd.Metadata[docker.Port])
	adjacencies := endpointTopology.Adjacency[report.MakeAdjacencyID(originID)]
	sort.Strings(adjacencies)
	for _, adj := range adjacencies {
		if _, address, port, ok := report.ParseEndpointNodeID(adj); ok {
			rows = append(rows, Row{
				Key:        local,
				ValueMajor: fmt.Sprintf("%s:%s", address, port),
			})
		}
	}
	return rows
}

func connectionDetailsTable(connectionRows []Row) Table {
	return Table{
		Title:   "Connection Details",
		Numeric: false,
		Rows:    append([]Row{{Key: "Local", ValueMajor: "Remote"}}, connectionRows...),
		Rank:    endpointRank,
	}
}

func addressOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	if val, ok := nmd.Metadata["addr"]; ok {
		rows = append(rows, Row{"Address", val, ""})
	}
	return Table{
		Title:   "Origin Address",
		Numeric: false,
		Rows:    rows,
		Rank:    addressRank,
	}, len(rows) > 0
}

func processOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{process.Comm, "Name (comm)"},
		{process.PID, "PID"},
		{process.PPID, "Parent PID"},
		{process.Cmdline, "Command"},
		{process.Threads, "# Threads"},
	} {
		if val, ok := nmd.Metadata[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}

	return Table{
		Title:   "Origin Process",
		Numeric: false,
		Rows:    rows,
		Rank:    processRank,
	}, len(rows) > 0
}

func containerOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{docker.ContainerID, "ID"},
		{docker.ContainerName, "Name"},
		{docker.ImageID, "Image ID"},
		{docker.ContainerPorts, "Ports"},
		{docker.ContainerCreated, "Created"},
		{docker.ContainerCommand, "Command"},
	} {
		if val, ok := nmd.Metadata[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}

	if val, ok := nmd.Metadata[docker.MemoryUsage]; ok {
		memory, err := strconv.ParseFloat(val, 64)
		if err == nil {
			memoryStr := fmt.Sprintf("%0.2f", memory/float64(mb))
			rows = append(rows, Row{Key: "Memory Usage (MB):", ValueMajor: memoryStr, ValueMinor: ""})
		}
	}

	return Table{
		Title:   "Origin Container",
		Numeric: false,
		Rows:    rows,
		Rank:    containerRank,
	}, len(rows) > 0
}

func containerImageOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{docker.ImageID, "Image ID"},
		{docker.ImageName, "Image name"},
	} {
		if val, ok := nmd.Metadata[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}
	return Table{
		Title:   "Origin Container Image",
		Numeric: false,
		Rows:    rows,
		Rank:    containerImageRank,
	}, len(rows) > 0
}

func hostOriginTable(nmd report.NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{host.HostName, "Host name"},
		{host.Load, "Load"},
		{host.OS, "Operating system"},
		{host.KernelVersion, "Kernel version"},
		{host.Uptime, "Uptime"},
	} {
		if val, ok := nmd.Metadata[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}

	return Table{
		Title:   "Origin Host",
		Numeric: false,
		Rows:    rows,
		Rank:    hostRank,
	}, len(rows) > 0
}
*/
