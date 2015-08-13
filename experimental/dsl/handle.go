package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/weaveworks/scope/report"
)

func handleJSON(tpy report.Topology) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(getView(r).eval(tpy.Copy()).NodeMetadatas); err != nil {
			log.Print(err)
			return
		}
	}
}

func handleDOT(tpy report.Topology) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		dot(w, getView(r).eval(tpy.Copy()))
	}
}

func handleSVG(tpy report.Topology) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command(engine(r), "-Tsvg")

		wc, err := cmd.StdinPipe()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cmd.Stdout = w

		dot(wc, getView(r).eval(tpy.Copy()))
		wc.Close()

		w.Header().Set("Content-Type", "image/svg+xml")
		if err := cmd.Run(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handleHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<html><head>\n")
	//fmt.Fprintf(w, `<meta http-equiv="refresh" content="10">`+"\n")
	fmt.Fprintf(w, "</head><body>\n")
	fmt.Fprintf(w, `<center><img src="/svg?%s" width="100%%" height="95%%"></center>`+"\n", r.URL.Query().Encode())
	fmt.Fprintf(w, "</body></html>\n")
}

func getView(r *http.Request) view {
	strs := []string{}
	scanner := bufio.NewScanner(r.Body)
	for scanner.Scan() {
		log.Printf("Provided expression: %s", scanner.Text())
		strs = append(strs, scanner.Text())
	}
	return parseView(strs)
}

func engine(r *http.Request) string {
	engine := r.FormValue("engine")
	if engine == "" {
		engine = "dot"
	}
	return engine
}
