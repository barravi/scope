package sniff_test

import (
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/gopacket"

	"github.com/weaveworks/scope/probe/sniff"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestSnifferShutdown(t *testing.T) {
	var (
		hostID = "abcd"
		src    = newMockSource([]byte{}, nil)
		on     = time.Millisecond
		off    = time.Millisecond
		s      = sniff.New(hostID, src, on, off)
	)

	// Stopping the source should terminate the sniffer.
	src.Close()
	time.Sleep(10 * time.Millisecond)

	// Try to get a report from the sniffer. It should block forever, as the
	// loop goroutine should have exited.
	report := make(chan struct{})
	go func() { s.Report(); close(report) }()
	select {
	case <-time.After(time.Millisecond):
	case <-report:
		t.Errorf("shouldn't get report after Close")
	}
}

func TestMerge(t *testing.T) {
	var (
		hostID = "xyz"
		src    = newMockSource([]byte{}, nil)
		on     = time.Millisecond
		off    = time.Millisecond
		rpt    = report.MakeReport()
		p      = sniff.Packet{
			SrcIP:     "1.0.0.0",
			SrcPort:   "1000",
			DstIP:     "2.0.0.0",
			DstPort:   "2000",
			Network:   512,
			Transport: 256,
		}
	)
	sniff.New(hostID, src, on, off).Merge(p, rpt)

	var (
		srcEndpointNodeID = report.MakeEndpointNodeID(hostID, p.SrcIP, p.SrcPort)
		dstEndpointNodeID = report.MakeEndpointNodeID(hostID, p.DstIP, p.DstPort)
	)
	if want, have := (report.Topology{
		Adjacency: report.Adjacency{
			report.MakeAdjacencyID(srcEndpointNodeID): report.MakeIDList(
				dstEndpointNodeID,
			),
		},
		EdgeMetadatas: report.EdgeMetadatas{
			report.MakeEdgeID(srcEndpointNodeID, dstEndpointNodeID): report.EdgeMetadata{
				WithBytes:   true,
				BytesEgress: 256,
			},
		},
		NodeMetadatas: report.NodeMetadatas{
			srcEndpointNodeID: report.NodeMetadata{},
			dstEndpointNodeID: report.NodeMetadata{},
		},
	}), rpt.Endpoint; !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}

	var (
		srcAddressNodeID = report.MakeAddressNodeID(hostID, p.SrcIP)
		dstAddressNodeID = report.MakeAddressNodeID(hostID, p.DstIP)
	)
	if want, have := (report.Topology{
		Adjacency: report.Adjacency{
			report.MakeAdjacencyID(srcAddressNodeID): report.MakeIDList(
				dstAddressNodeID,
			),
		},
		EdgeMetadatas: report.EdgeMetadatas{
			report.MakeEdgeID(srcAddressNodeID, dstAddressNodeID): report.EdgeMetadata{
				WithBytes:   true,
				BytesEgress: 512,
			},
		},
		NodeMetadatas: report.NodeMetadatas{
			srcAddressNodeID: report.NodeMetadata{},
			dstAddressNodeID: report.NodeMetadata{},
		},
	}), rpt.Address; !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}

type mockSource struct {
	mtx  sync.RWMutex
	data []byte
	err  error
}

func newMockSource(data []byte, err error) *mockSource {
	return &mockSource{
		data: data,
		err:  err,
	}
}

func (s *mockSource) ZeroCopyReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.data, gopacket.CaptureInfo{
		Timestamp:     time.Now(),
		CaptureLength: len(s.data),
		Length:        len(s.data),
	}, s.err
}

func (s *mockSource) Close() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.err = io.EOF
}
