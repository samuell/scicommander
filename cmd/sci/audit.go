package main

import (
	"os"
	"strconv"
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
		Logs: []*TaskLog{{
			Logs: []*ExecutorLog{{}},
			Metadata: map[string]string{
				"output_path": outFilePath,
			},
			Outputs: []*OutputFileLog{},
		}},
		Upstream: []AuditInfo{},
	}
}

type AuditInfo struct {
	Inputs       []string    `json:"inputs"`
	Outputs      []string    `json:"outputs"`
	Executors    []*Executor `json:"executors"`
	Logs         []*TaskLog  `json:"logs"`
	CreationTime time.Time   `json:"creation_time"`
	Upstream     []AuditInfo `json:"upstream"`

	// Tags is kept for reading audit files written in the old format, where
	// execution metadata was stored under "tags". It is not written to new
	// audit files.
	Tags *SciTags `json:"tags,omitempty"`
}

type Executor struct {
	Image   string   `json:"image"`
	Command []string `json:"command"`
}

// TaskLog contains information about a task execution, following the TES
// "tesTaskLog" schema.
type TaskLog struct {
	Logs       []*ExecutorLog    `json:"logs"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Outputs    []*OutputFileLog  `json:"outputs"`
	SystemLogs []string          `json:"system_logs,omitempty"`
}

// ExecutorLog contains information about a single executor run, following the
// TES "tesExecutorLog" schema.
type ExecutorLog struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Stdout    string    `json:"stdout"`
	Stderr    string    `json:"stderr"`
	ExitCode  int       `json:"exit_code"`
}

// OutputFileLog contains information about an output file, following the TES
// "tesOutputFileLog" schema.
type OutputFileLog struct {
	URL       string `json:"url"`
	Path      string `json:"path"`
	SizeBytes string `json:"size_bytes"`
}

// SciTags is the legacy representation of execution metadata, which was
// stored under the "tags" section of the audit JSON. It is only used when
// reading audit files written by older versions of SciCommander.
type SciTags struct {
	OutPath     string        `json:"output_path"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	DurationSec int           `json:"duration_s"`
	DirDepth    int           `json:"dir_depth"`
}

// SetExecutionInfo stores metadata about an executed command in the audit
// info's log section.
func (ai *AuditInfo) SetExecutionInfo(startTime time.Time, endTime time.Time, exitCode int, dirDepth int) {
	if len(ai.Logs) == 0 {
		ai.Logs = []*TaskLog{{}}
	}
	taskLog := ai.Logs[0]

	if len(taskLog.Logs) == 0 {
		taskLog.Logs = []*ExecutorLog{{}}
	}
	executorLog := taskLog.Logs[0]

	duration := endTime.Sub(startTime)

	taskLog.StartTime = startTime
	taskLog.EndTime = endTime
	executorLog.StartTime = startTime
	executorLog.EndTime = endTime
	executorLog.ExitCode = exitCode

	if taskLog.Metadata == nil {
		taskLog.Metadata = map[string]string{}
	}
	taskLog.Metadata["dir_depth"] = strconv.Itoa(dirDepth)
	taskLog.Metadata["duration"] = strconv.FormatInt(int64(duration), 10)
	taskLog.Metadata["duration_s"] = strconv.FormatFloat(duration.Seconds(), 'f', -1, 64)

	ai.CreationTime = startTime
}

// SetOutputFiles stores information about the files produced by the command in
// the audit info's log section.
func (ai *AuditInfo) SetOutputFiles(paths []string) {
	if len(ai.Logs) == 0 {
		ai.Logs = []*TaskLog{{}}
	}

	outputLogs := []*OutputFileLog{}
	for _, path := range paths {
		sizeBytes := int64(0)
		if stat, err := os.Stat(path); err == nil {
			sizeBytes = stat.Size()
		}
		outputLogs = append(outputLogs, &OutputFileLog{
			URL:       path,
			Path:      path,
			SizeBytes: strconv.FormatInt(sizeBytes, 10),
		})
	}

	ai.Logs[0].Outputs = outputLogs
}

// OutPath returns the path of the file this audit info was created for.
func (ai *AuditInfo) OutPath() string {
	if len(ai.Logs) > 0 {
		if outPath, ok := ai.Logs[0].Metadata["output_path"]; ok {
			return outPath
		}
		if len(ai.Logs[0].Outputs) > 0 {
			return ai.Logs[0].Outputs[0].Path
		}
	}
	if ai.Tags != nil {
		return ai.Tags.OutPath
	}
	return ""
}

// StartTime returns the time the command started.
func (ai *AuditInfo) StartTime() time.Time {
	if len(ai.Logs) > 0 {
		return ai.Logs[0].StartTime
	}
	if ai.Tags != nil {
		return ai.Tags.StartTime
	}
	return time.Time{}
}

// EndTime returns the time the command ended.
func (ai *AuditInfo) EndTime() time.Time {
	if len(ai.Logs) > 0 {
		return ai.Logs[0].EndTime
	}
	if ai.Tags != nil {
		return ai.Tags.EndTime
	}
	return time.Time{}
}

// Duration returns the duration of the command execution.
func (ai *AuditInfo) Duration() time.Duration {
	if len(ai.Logs) > 0 {
		return ai.Logs[0].EndTime.Sub(ai.Logs[0].StartTime)
	}
	if ai.Tags != nil {
		return ai.Tags.Duration
	}
	return 0
}

// DirDepth returns how many directory levels down the audit file is located,
// in relation to the command that produced it.
func (ai *AuditInfo) DirDepth() int {
	if len(ai.Logs) > 0 {
		if dirDepth, ok := ai.Logs[0].Metadata["dir_depth"]; ok {
			if depth, err := strconv.Atoi(dirDepth); err == nil {
				return depth
			}
		}
	}
	if ai.Tags != nil {
		return ai.Tags.DirDepth
	}
	return 0
}
