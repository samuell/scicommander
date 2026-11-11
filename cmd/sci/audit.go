package main

import (
	"strings"
	"time"
)

func NewAuditInfo(outFilePath string, cmd string, inPaths []string, outPaths []string) *AuditInfo {
	inputs := []*AuditTarget{}
	for _, inPath := range inPaths {
		inputs = append(inputs, NewAuditTarget(inPath, inPath))
	}
	outputs := []*AuditTarget{}
	for _, outPath := range outPaths {
		outputs = append(outputs, NewAuditTarget(outPath, outPath))
	}
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
	Inputs    []*AuditTarget `json:"inputs"`
	Outputs   []*AuditTarget `json:"outputs"`
	Executors []*Executor    `json:"executors"`
	Tags      *SciTags       `json:"tags"`
	Upstream  []AuditInfo    `json:"upstream"`
}

func NewAuditTarget(url string, path string) *AuditTarget {
	return &AuditTarget{
		Url:  url,
		Path: path,
		Type: "FILE",
	}
}

type AuditTarget struct {
	Url  string `json:url`
	Path string `json:path`
	Type string `json:type`
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
