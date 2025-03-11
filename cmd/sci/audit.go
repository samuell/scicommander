package main

import (
	"strings"
	"time"
)

func NewAuditInfo(cmd string, inputs []string, outputs []string) *AuditInfo {
	return &AuditInfo{
		Inputs:  inputs,
		Outputs: outputs,
		Executors: []*Executor{{
			Image:   "None",
			Command: strings.Split(cmd, " "),
		}},
		Tags:     &SciTags{},
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

type SciTags struct {
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	DurationSec int           `json:"duration_s"`
}
