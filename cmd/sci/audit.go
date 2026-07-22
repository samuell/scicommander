package main

import (
	"strings"
	"time"
)

func NewAuditInfo(outFilePath string, cmd string, inputs []string, outputs []string) *AuditInfo {
	return &AuditInfo{
		Inputs:  inputs,
		Outputs: outputs,
		Executors: []*Executor{{
			Image:   "None",
			Command: strings.Split(cmd, " "),
		}},
		Tags:     &SciTags{OutPath: outFilePath},
		Upstream: []AuditInfo{},
	}
}

type AuditInfo struct {
	Inputs    []string    `json:"inputs"`
	Outputs   []string    `json:"outputs"`
	Executors []*Executor `json:"executors"`
	Tags      *SciTags    `json:"tags"`
	Upstream  []AuditInfo `json:"upstream"`
}

type Executor struct {
	Image   string   `json:"image"`
	Command []string `json:"command"`
}

// SciTags contains key-value tags, which is here used to store metadata about
// command executions. Documentation for some of the fields are below:
//   - DirDepth stores into on how many directory levels down an audit file is
//     located, in relation to the command that produced it.
type SciTags struct {
	OutPath     string        `json:"output_path"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	DurationSec int           `json:"duration_s"`
	DirDepth    int           `json:"dir_depth"`
}
