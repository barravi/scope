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
		exprs := parseExpressions(getExpressions(r))
		log.Printf("%s: %d expr(s)", r.URL.Path, len(exprs))
		for _, expr := range exprs {
			tpy = expr.eval(tpy)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tpy); err != nil {
			log.Print(err)
			return
		}

	}
}

func getExpressions(r *http.Request) []string {
	a := []string{}
	s := bufio.NewScanner(r.Body)
	for s.Scan() {
		log.Printf("Expression Scan: %s", s.Text())
		a = append(a, s.Text())
	}
	log.Printf("Expression Scan: %s", s.Err())
	return a
}
