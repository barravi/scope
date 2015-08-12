package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"

	"github.com/weaveworks/scope/report"
)

func handleJSON(tpy report.Topology) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpy := tpy.Copy()
		exprs := parseView(getExpressions(r))
		log.Printf("%s: %d expr(s)", r.URL.Path, len(exprs))
		for _, expr := range exprs {
			tpy = expr.eval(tpy)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tpy.NodeMetadatas); err != nil {
			log.Print(err)
			return
		}
	}
}

func getExpressions(r *http.Request) []string {
	a := []string{}
	s := bufio.NewScanner(r.Body)
	for s.Scan() {
		log.Printf("Provided expression: %s", s.Text())
		a = append(a, s.Text())
	}
	return a
}
