package render_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/test"
)

func TestOriginTable(t *testing.T) {
	if _, ok := render.OriginTable(test.Report, "not-found"); ok {
		t.Errorf("unknown origin ID gave unexpected success")
	}
	for originID, want := range map[string]render.Table{
		test.ClientAddressNodeID: {
			Title:   "Origin Address",
			Numeric: false,
			Rows: []render.Row{
				{"Address", test.ClientIP, ""},
			},
		},
		test.ServerProcessNodeID: {
			Title:   "Origin Process",
			Numeric: false,
			Rank:    2,
			Rows: []render.Row{
				{"Name (comm)", "apache", ""},
				{"PID", test.ServerPID, ""},
			},
		},
		test.ServerHostNodeID: {
			Title:   "Origin Host",
			Numeric: false,
			Rank:    1,
			Rows: []render.Row{
				{"Host name", test.ServerHostName, ""},
				{"Load", "0.01 0.01 0.01", ""},
				{"Operating system", "Linux", ""},
			},
		},
	} {
		have, ok := render.OriginTable(test.Report, originID)
		if !ok {
			t.Errorf("%q: not OK", originID)
			continue
		}
		if !reflect.DeepEqual(want, have) {
			t.Errorf("%q: %s", originID, test.Diff(want, have))
		}
	}
}

func TestMakeDetailedNode(t *testing.T) {
	renderableNode := render.ContainerRenderer.Render(test.Report)[test.ServerContainerID]
	have := render.MakeDetailedNode(test.Report, renderableNode)
	want := render.DetailedNode{
		ID:         test.ServerContainerID,
		LabelMajor: "server",
		LabelMinor: test.ServerHostName,
		Pseudo:     false,
		Tables: []render.Table{
			{
				Title:   "Connections",
				Numeric: true,
				Rank:    100,
				Rows: []render.Row{
					{"Egress packet rate", "75", "packets/sec"},
					{"Egress byte rate", "750", "Bps"},
				},
			},
			{
				Title:   "Origin Container",
				Numeric: false,
				Rank:    3,
				Rows: []render.Row{
					{"ID", test.ServerContainerID, ""},
					{"Name", "server", ""},
					{"Image ID", test.ServerContainerImageID, ""},
				},
			},
			{
				Title:   "Origin Process",
				Numeric: false,
				Rank:    2,
				Rows: []render.Row{
					{"Name (comm)", "apache", ""},
					{"PID", test.ServerPID, ""},
				},
			},
			{
				Title:   "Origin Host",
				Numeric: false,
				Rank:    1,
				Rows: []render.Row{
					{"Host name", test.ServerHostName, ""},
					{"Load", "0.01 0.01 0.01", ""},
					{"Operating system", "Linux", ""},
				},
			},
			{
				Title:   "Connection Details",
				Numeric: false,
				Rows: []render.Row{
					{"Local", "Remote", ""},
					{
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
						fmt.Sprintf("%s:%s", test.UnknownClient1IP, test.ClientPort54010),
						"",
					},
					{
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
						fmt.Sprintf("%s:%s", test.UnknownClient1IP, test.ClientPort54020),
						"",
					},
					{
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
						fmt.Sprintf("%s:%s", test.UnknownClient3IP, test.ClientPort54020),
						"",
					},
					{
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
						fmt.Sprintf("%s:%s", test.ClientIP, test.ClientPort54001),
						"",
					},
					{
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
						fmt.Sprintf("%s:%s", test.ClientIP, test.ClientPort54002),
						"",
					},
					{
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
						fmt.Sprintf("%s:%s", test.RandomClientIP, test.ClientPort12345),
						"",
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
