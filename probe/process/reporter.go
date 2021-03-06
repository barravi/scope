package process

import (
	"strconv"

	"github.com/weaveworks/scope/report"
)

// We use these keys in node metadata
const (
	PID     = "pid"
	Comm    = "comm"
	PPID    = "ppid"
	Cmdline = "cmdline"
	Threads = "threads"
)

// Reporter generates Reports containing the Process topology.
type Reporter struct {
	scope  string
	walker Walker
}

// NewReporter makes a new Reporter.
func NewReporter(walker Walker, scope string) *Reporter {
	return &Reporter{
		scope:  scope,
		walker: walker,
	}
}

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	result := report.MakeReport()
	processes, err := r.processTopology()
	if err != nil {
		return result, err
	}
	result.Process.Merge(processes)
	return result, nil
}

func (r *Reporter) processTopology() (report.Topology, error) {
	t := report.NewTopology()
	err := r.walker.Walk(func(p Process) {
		pidstr := strconv.Itoa(p.PID)
		nodeID := report.MakeProcessNodeID(r.scope, pidstr)
		t.NodeMetadatas[nodeID] = report.MakeNodeMetadataWith(map[string]string{
			PID:     pidstr,
			Comm:    p.Comm,
			Cmdline: p.Cmdline,
			Threads: strconv.Itoa(p.Threads),
		})
		if p.PPID > 0 {
			t.NodeMetadatas[nodeID].Metadata[PPID] = strconv.Itoa(p.PPID)
		}
	})

	return t, err
}
