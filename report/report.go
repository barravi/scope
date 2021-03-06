package report

import (
	"fmt"
	"strings"
	"time"
)

// Report is the core data type. It's produced by probes, and consumed and
// stored by apps. It's composed of multiple topologies, each representing
// a different (related, but not equivalent) view of the network.
type Report struct {
	// Endpoint nodes are individual (address, port) tuples on each host.
	// They come from inspecting active connections and can (theoretically)
	// be traced back to a process. Edges are present.
	Endpoint Topology

	// Address nodes are addresses (e.g. ifconfig) on each host. Certain
	// information may be present in this topology that can't be mapped to
	// endpoints (e.g. ICMP). Edges are present.
	Address Topology

	// Process nodes are processes on each host. Edges are not present.
	Process Topology

	// Container nodes represent all Docker containers on hosts running probes.
	// Metadata includes things like containter id, name, image id etc.
	// Edges are not present.
	Container Topology

	// ContainerImages nodes represent all Docker containers images on
	// hosts running probes. Metadata includes things like image id, name etc.
	// Edges are not present.
	ContainerImage Topology

	// Host nodes are physical hosts that run probes. Metadata includes things
	// like operating system, load, etc. The information is scraped by the
	// probes with each published report. Edges are not present.
	Host Topology

	// Overlay nodes are active peers in any software-defined network that's
	// overlaid on the infrastructure. The information is scraped by polling
	// their status endpoints. Edges could be present, but aren't currently.
	Overlay Topology

	// Sampling data for this report.
	Sampling Sampling

	// Window is the amount of time that this report purports to represent.
	// Windows must be carefully merged. They should only be added when
	// reports cover non-overlapping periods of time. By default, we assume
	// that's true, and add windows in merge operations. When that's not true,
	// such as in the app, we expect the component to overwrite the window
	// before serving it to consumers.
	Window time.Duration
}

// MakeReport makes a clean report, ready to Merge() other reports into.
func MakeReport() Report {
	return Report{
		Endpoint:       NewTopology(),
		Address:        NewTopology(),
		Process:        NewTopology(),
		Container:      NewTopology(),
		ContainerImage: NewTopology(),
		Host:           NewTopology(),
		Overlay:        NewTopology(),
		Sampling:       Sampling{},
		Window:         0,
	}
}

// Topologies returns a slice of Topologies in this report
func (r Report) Topologies() []Topology {
	return []Topology{
		r.Endpoint,
		r.Address,
		r.Process,
		r.Container,
		r.ContainerImage,
		r.Host,
		r.Overlay,
	}
}

// Validate checks the report for various inconsistencies.
func (r Report) Validate() error {
	var errs []string
	for _, topology := range r.Topologies() {
		if err := topology.Validate(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if r.Sampling.Count > r.Sampling.Total {
		errs = append(errs, fmt.Sprintf("sampling count (%d) bigger than total (%d)", r.Sampling.Count, r.Sampling.Total))
	}
	if len(errs) > 0 {
		return fmt.Errorf("%d error(s): %s", len(errs), strings.Join(errs, "; "))
	}
	return nil
}

// Sampling describes how the packet data sources for this report were
// sampled. It can be used to calculate effective sample rates. We can't
// just put the rate here, because that can't be accurately merged. Counts
// in e.g. edge metadata structures have already been adjusted to
// compensate for the sample rate.
type Sampling struct {
	Count uint64 // observed and processed
	Total uint64 // observed overall
}

// Rate returns the effective sampling rate.
func (s Sampling) Rate() float64 {
	if s.Total <= 0 {
		return 1.0
	}
	return float64(s.Count) / float64(s.Total)
}

const (
	// HostNodeID is a metadata foreign key, linking a node in any topology to
	// a node in the host topology. That host node is the origin host, where
	// the node was originally detected.
	HostNodeID = "host_node_id"
)

// TopologySelector selects a single topology from a report.
type TopologySelector func(r Report) Topology

// SelectEndpoint selects the endpoint topology.
func SelectEndpoint(r Report) Topology {
	return r.Endpoint
}

// SelectProcess selects the process topology.
func SelectProcess(r Report) Topology {
	return r.Process
}

// SelectContainer selects the container topology.
func SelectContainer(r Report) Topology {
	return r.Container
}

// SelectContainerImage selects the container image topology.
func SelectContainerImage(r Report) Topology {
	return r.ContainerImage
}

// SelectAddress selects the address topology.
func SelectAddress(r Report) Topology {
	return r.Address
}

// SelectHost selects the address topology.
func SelectHost(r Report) Topology {
	return r.Host
}
