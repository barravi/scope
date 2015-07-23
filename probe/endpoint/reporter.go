package endpoint

import (
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

// Reporter generates Reports containing the Endpoint topology.
type Reporter struct {
	hostID           string
	hostName         string
	includeProcesses bool
	includeNAT       bool
}

// SpyDuration is an exported prometheus metric
var SpyDuration = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Namespace: "scope",
		Subsystem: "probe",
		Name:      "spy_time_nanoseconds",
		Help:      "Total time spent spying on active connections.",
		MaxAge:    10 * time.Second, // like statsd
	},
	[]string{},
)

// NewReporter creates a new Reporter that invokes procspy.Connections to
// generate a report.Report that contains every discovered (spied) connection
// on the host machine, at the granularity of host and port. That information
// is stored in the Endpoint topology. It optionally enriches that topology
// with process (PID) information.
func NewReporter(hostID, hostName string, includeProcesses bool) *Reporter {
	return &Reporter{
		hostID:           hostID,
		hostName:         hostName,
		includeProcesses: includeProcesses,
		includeNAT:       conntrackModulePresent(),
	}
}

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	defer func(begin time.Time) {
		SpyDuration.WithLabelValues().Observe(float64(time.Since(begin)))
	}(time.Now())

	rpt := report.MakeReport()
	conns, err := procspy.Connections(r.includeProcesses)
	if err != nil {
		return rpt, err
	}

	for conn := conns.Next(); conn != nil; conn = conns.Next() {
		r.addConnection(&rpt, conn)
	}

	if r.includeNAT {
		err = applyNAT(rpt, r.hostID)
	}

	return rpt, err
}

func (r *Reporter) addConnection(rpt *report.Report, c *procspy.Connection) {
	var (
		localAddressNodeID  = report.MakeAddressNodeID(r.hostID, c.LocalAddress.String())
		remoteAddressNodeID = report.MakeAddressNodeID(r.hostID, c.RemoteAddress.String())
		adjecencyID         = report.MakeAdjacencyID(localAddressNodeID)
		edgeID              = report.MakeEdgeID(localAddressNodeID, remoteAddressNodeID)
	)

	rpt.Address.Adjacency[adjecencyID] = rpt.Address.Adjacency[adjecencyID].Add(remoteAddressNodeID)

	if _, ok := rpt.Address.NodeMetadatas[localAddressNodeID]; !ok {
		rpt.Address.NodeMetadatas[localAddressNodeID] = report.NewNodeMetadata(map[string]string{
			docker.Name: r.hostName,
			docker.Addr: c.LocalAddress.String(),
		})
	}

	countTCPConnection(rpt.Address.EdgeMetadatas, edgeID)

	if c.Proc.PID > 0 {
		var (
			localEndpointNodeID  = report.MakeEndpointNodeID(r.hostID, c.LocalAddress.String(), strconv.Itoa(int(c.LocalPort)))
			remoteEndpointNodeID = report.MakeEndpointNodeID(r.hostID, c.RemoteAddress.String(), strconv.Itoa(int(c.RemotePort)))
			adjecencyID          = report.MakeAdjacencyID(localEndpointNodeID)
			edgeID               = report.MakeEdgeID(localEndpointNodeID, remoteEndpointNodeID)
		)

		rpt.Endpoint.Adjacency[adjecencyID] = rpt.Endpoint.Adjacency[adjecencyID].Add(remoteEndpointNodeID)

		if _, ok := rpt.Endpoint.NodeMetadatas[localEndpointNodeID]; !ok {
			// First hit establishes NodeMetadata for scoped local address + port
			md := report.NewNodeMetadata(map[string]string{
				"addr": c.LocalAddress.String(),
				"port": strconv.Itoa(int(c.LocalPort)),
				"pid":  fmt.Sprintf("%d", c.Proc.PID),
			})

			rpt.Endpoint.NodeMetadatas[localEndpointNodeID] = md
		}

		countTCPConnection(rpt.Endpoint.EdgeMetadatas, edgeID)
	}
}

func countTCPConnection(mds report.EdgeMetadatas, key string) {
	md := mds[key]
	if md.MaxConnCountTCP == nil {
		md.MaxConnCountTCP = new(uint64)
	}
	*md.MaxConnCountTCP++
	mds[key] = md
}
