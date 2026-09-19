package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type request struct {
	Task string `json:"task"`
}

type response struct {
	Summary string   `json:"summary"`
	Changes []change `json:"changes"`
}

type change struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func main() {
	var req request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if req.Task == "" {
		fmt.Fprintln(os.Stderr, "missing task")
		os.Exit(2)
	}
	if err := json.NewEncoder(os.Stdout).Encode(response{
		Summary: "deterministic smoke edit",
		Changes: []change{{Path: "result.txt", Content: "good\n"}},
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
