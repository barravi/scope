package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/weaveworks/scope/report"
)

func dot(w io.Writer, tpy report.Topology) {
	fmt.Fprintf(w, "digraph G {\n")
	fmt.Fprintf(w, "\tgraph [ overlap=false ];\n")
	fmt.Fprintf(w, "\tnode [ shape=circle, style=filled ];\n")
	fmt.Fprintf(w, "\toutputorder=edgesfirst;\n")
	fmt.Fprintf(w, "\n")

	for id := range tpy.NodeMetadatas {
		fmt.Fprintf(w, "\t%q [label=%q];\n", id, strings.Join(strings.Split(id, ";"), "\n"))
	}
	fmt.Fprintf(w, "\n")

	for src, dsts := range tpy.Adjacency {
		for _, dst := range dsts {
			fmt.Fprintf(w, "\t%q -> %q;\n", src, dst)
		}
	}

	fmt.Fprintf(w, "}\n")
}
